package dbutil

import (
	"context"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryBuilderItf interface {
	Build(ctx context.Context) (string, []any, error)
}

type PgxClientItf interface {
	PgxWriterItf
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close() error
}

type PgxWriterItf interface {
	PgxReaderItf
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type PgxReaderItf interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func IsDebugMode() bool {
	return strings.ToLower(os.Getenv("DEBUG_DB_UTIL")) == "true"
}
