package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	db "github.com/valkyraycho/bank/db/sqlc"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskHandler interface {
	Start() error
	Shutdown()
	HandleTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskHandler struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskHandler(redisOpt *asynq.RedisClientOpt, store db.Store) TaskHandler {
	return &RedisTaskHandler{
		server: asynq.NewServer(
			redisOpt,
			asynq.Config{
				Queues: map[string]int{
					QueueCritical: 10,
					QueueDefault:  5,
				},
				ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
					log.Error().
						Err(err).
						Str("type", task.Type()).
						Bytes("payload", task.Payload()).
						Msg("task failed")
				}),
				Logger: NewAsynqLogger(),
			},
		),
		store: store,
	}
}

func (handler *RedisTaskHandler) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, handler.HandleTaskSendVerifyEmail)

	return handler.server.Start(mux)
}

func (handler *RedisTaskHandler) Shutdown() {
	handler.server.Shutdown()
}
