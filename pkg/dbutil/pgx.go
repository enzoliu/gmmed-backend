package dbutil

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func GetColumn[T any](ctx context.Context, db PgxReaderItf, builder QueryBuilderItf) (T, error) {
	var zeroValue T

	rows, err := Query(ctx, db, builder)
	if err != nil {
		return zeroValue, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[T])
	if err != nil {
		return zeroValue, errors.Wrap(err, "failed to collect one column")
	}
	return result, nil
}

func GetColumnToAddrOf[T any](ctx context.Context, db PgxReaderItf, builder QueryBuilderItf) (*T, error) {
	rows, err := Query(ctx, db, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOf[T])
	if err != nil {
		return nil, errors.Wrap(err, "failed to collect one column")
	}
	return result, nil
}

func GetAllColumns[T any](ctx context.Context, db PgxReaderItf, builder QueryBuilderItf) ([]T, error) {
	rows, err := Query(ctx, db, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectRows(rows, pgx.RowTo[T])
	if err != nil {
		return nil, errors.Wrap(err, "failed to collect columns")
	}
	return result, nil
}

func GetOne[T any](ctx context.Context, db PgxReaderItf, builder QueryBuilderItf) (*T, error) {
	rows, err := Query(ctx, db, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[T])
}

func GetOneOptional[T any](ctx context.Context, db PgxReaderItf, builder QueryBuilderItf) (*T, error) {
	results, err := GetAll[T](ctx, db, builder)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func GetAll[T any](ctx context.Context, db PgxReaderItf, builder QueryBuilderItf) ([]*T, error) {
	rows, err := Query(ctx, db, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[T])
	if err != nil {
		return nil, errors.Wrap(err, "failed to collect rows")
	}
	return result, nil
}

func GetPage[T any](ctx context.Context, db PgxReaderItf, builder bob.BaseQuery[*dialect.SelectQuery], limit int) ([]*T, bool, error) {
	return getPage(ctx, db, pgx.RowToAddrOfStructByName[T], builder, limit)
}

func GetPageOfColumn[T any](ctx context.Context, db PgxReaderItf, builder bob.BaseQuery[*dialect.SelectQuery], limit int) ([]T, bool, error) {
	return getPage(ctx, db, pgx.RowTo[T], builder, limit)
}

func GetPageOfAddrColumn[T any](ctx context.Context, db PgxReaderItf, builder bob.BaseQuery[*dialect.SelectQuery], limit int) ([]*T, bool, error) {
	return getPage(ctx, db, pgx.RowToAddrOf[T], builder, limit)
}

func getPage[T any](ctx context.Context, db PgxReaderItf, rowToFunc pgx.RowToFunc[T], builder bob.BaseQuery[*dialect.SelectQuery], limit int) ([]T, bool, error) {
	// build limit clause
	var limitClause bob.Mod[*dialect.SelectQuery]
	if limit <= -1 {
		limitClause = sm.Limit("ALL")
	} else {
		limitClause = sm.Limit(limit + 1)
	}
	builder.Apply(limitClause)

	// query
	rows, err := Query(ctx, db, builder)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, rowToFunc)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to collect rows")
	}

	// build result
	hasNext := false
	if limit >= 0 && len(results) > limit {
		hasNext = true
		results = results[:limit]
	}
	return results, hasNext, nil
}

func Query(ctx context.Context, db PgxReaderItf, builder QueryBuilderItf) (pgx.Rows, error) {
	query, args, err := builder.Build(ctx)
	if err != nil {

		return nil, errors.Wrap(err, "failed to build query")
	}

	if IsDebugMode() {
		queryLog := strings.ReplaceAll(query, "\n", " ")
		slog.DebugContext(ctx, queryLog, "args", args)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query")
	}
	return rows, nil
}

// won't return error when the query affected 0 rows
func MayExec(ctx context.Context, db PgxWriterItf, builder QueryBuilderItf) error {
	if _, err := Exec(ctx, db, builder); err != nil {
		return err
	}
	return nil
}

// Won't return error when the query affected 0 rows, but will return an empty string
func MayExecForAuditLog(ctx context.Context, db PgxWriterItf, builder QueryBuilderItf) (string, error) {
	query, args, err := builder.Build(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to build query")
	}

	if IsDebugMode() {
		queryLog := strings.ReplaceAll(query, "\n", " ")
		slog.DebugContext(ctx, queryLog, "args", args)
	}

	// Use Query instead of QueryRow since we might not have QueryRow in the interface
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute query with returning")
	}
	defer rows.Close()

	// No rows returned means the update affected 0 rows
	if !rows.Next() {
		return "", nil
	}

	// Scan the first column of the first row
	var returnValue string
	if err := rows.Scan(&returnValue); err != nil {
		return "", errors.Wrap(err, "failed to scan returned value")
	}

	return returnValue, nil
}

// return pgx.ErrNoRows when the query affected 0 rows
func ShouldExec(ctx context.Context, db PgxWriterItf, builder QueryBuilderItf) error {
	result, err := Exec(ctx, db, builder)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.Wrap(pgx.ErrNoRows, "query affected 0 rows")
	}
	return nil
}

func Exec(ctx context.Context, db PgxWriterItf, builder QueryBuilderItf) (pgconn.CommandTag, error) {
	query, args, err := builder.Build(ctx)
	if err != nil {
		return pgconn.CommandTag{}, errors.Wrap(err, "failed to build query")
	}

	if IsDebugMode() {
		queryLog := strings.ReplaceAll(query, "\n", " ")
		slog.DebugContext(ctx, queryLog, "args", args)
	}

	result, err := db.Exec(ctx, query, args...)
	if err != nil {
		return pgconn.CommandTag{}, errors.Wrap(err, "failed to execute")
	}
	return result, nil
}
