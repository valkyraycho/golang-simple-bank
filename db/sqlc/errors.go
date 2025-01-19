package db

import "github.com/lib/pq"

const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)

var ErrUniqueViolation = &pq.Error{
	Code: UniqueViolation,
}

var ErrForeignKeyViolation = &pq.Error{
	Code: ForeignKeyViolation,
}
