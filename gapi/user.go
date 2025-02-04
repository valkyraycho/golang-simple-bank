package gapi

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	db "github.com/valkyraycho/bank/db/sqlc"
	"github.com/valkyraycho/bank/pb"
	"github.com/valkyraycho/bank/utils"
	"github.com/valkyraycho/bank/validator"
	"github.com/valkyraycho/bank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)

	if len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := utils.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	log.Info().Msg(">>> creating user...")
	time.Sleep(10 * time.Second)

	createUserTxResult, err := s.store.CreateUserTx(ctx, db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashedPassword,
			FullName:       req.GetFullName(),
			Email:          req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return s.taskDistributor.DistributeTaskSendVerifyEmail(
				ctx,
				&worker.PayloadSendVerifyEmail{Username: user.Username},
				opts...)
		},
	})
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Code {
			case db.UniqueViolation:
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}
	log.Info().Msg(">>> done creating user")

	return &pb.CreateUserResponse{User: convertUser(createUserTxResult.User)}, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := validator.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	if err := validator.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	return violations
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)

	if len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

	user, err := s.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %s", err)
	}
	if err := utils.CheckPassword(req.GetPassword(), user.HashedPassword); err != nil {
		return nil, status.Errorf(codes.NotFound, "incorrect password: %s", err)
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.cfg.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to create access token: %s", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.cfg.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to create refresh token: %s", err)
	}

	mtdt := s.extractMetadata(ctx)
	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to create session: %s", err)
	}
	return &pb.LoginUserResponse{
		User:                  convertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return violations
}

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
	}
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	payload, err := s.authorizeUser(ctx, []string{utils.BankerRole, utils.DepositorRole})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %s", err)
	}

	violations := validateUpdateUserRequest(req)
	if len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

	if payload.Role != utils.BankerRole && payload.Username != req.GetUsername() {
		return nil, status.Error(codes.PermissionDenied, "no permission to update other user's info")
	}

	args := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
		}
		args.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}
		args.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := s.store.UpdateUser(ctx, args)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}
	return &pb.UpdateUserResponse{User: convertUser(user)}, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.Password != nil {
		if err := validator.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if req.Email != nil {
		if err := validator.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	if req.FullName != nil {
		if err := validator.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	return violations
}
