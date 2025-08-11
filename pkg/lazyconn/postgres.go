package lazyconn

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func (cfg *SQLConfig) ToPostgreSQL() string {
	// NOTE: The order of the arguments is important.
	//       Password must be the first one, otherwise, it will be printed in the error if other arguments are empty.
	if strings.HasPrefix(cfg.Host, "/") {
		return fmt.Sprintf("password=%s host=%s user=%s dbname=%s", cfg.Password, cfg.Host, cfg.User, cfg.Database)
	}
	return fmt.Sprintf("password=%s host=%s port=%s user=%s dbname=%s", cfg.Password, cfg.Host, cfg.Port, cfg.User, cfg.Database)
}

type PostgresClient struct {
	lazyClient LazyClient[*pgxpool.Pool]
}

func NewPostgresClient(ctx context.Context, cfg SQLConfig, options ...Option) *PostgresClient {
	opts := applyOptions(options)

	lazyClient := NewLazyClient(func(ctx context.Context) (*pgxpool.Pool, error) {
		poolConfig, err := pgxpool.ParseConfig(cfg.ToPostgreSQL())
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse postgres config")
		}

		client, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create postgres client")
		}

		logger := slog.With("host", cfg.Host, "port", cfg.Port, "database", cfg.Database)
		logger.InfoContext(ctx, "connecting to postgres")

		pingFunc := func() (*pgxpool.Pool, error) {
			if err := client.Ping(ctx); err != nil {
				logger.WarnContext(ctx, "failed to connect postgres, retrying", "err", err.Error())
				return nil, errors.Wrap(err, "failed to ping db")
			}
			return client, nil
		}

		backOffInst := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), opts.MaxRetry)
		return backoff.RetryWithData(pingFunc, backOffInst)
	})

	lazyClient.Preload(ctx)

	return &PostgresClient{
		lazyClient: lazyClient,
	}
}

func (c *PostgresClient) GetClient(ctx context.Context) (*pgxpool.Pool, error) {
	return c.lazyClient.Load(ctx)
}

func (c *PostgresClient) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	client, err := c.lazyClient.Load(ctx)
	if err != nil {
		return nil, err
	}
	return client.Query(ctx, sql, args...)
}

func (c *PostgresClient) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	client, err := c.lazyClient.Load(ctx)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return client.Exec(ctx, sql, arguments...)
}

func (c *PostgresClient) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	client, err := c.lazyClient.Load(ctx)
	if err != nil {
		return 0, err
	}
	return client.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (c *PostgresClient) Begin(ctx context.Context) (pgx.Tx, error) {
	client, err := c.lazyClient.Load(ctx)
	if err != nil {
		return nil, err
	}
	return client.Begin(ctx)
}

func (c *PostgresClient) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	client, err := c.lazyClient.Load(ctx)
	if err != nil {
		return nil, err
	}
	return client.BeginTx(ctx, txOptions)
}

func (c *PostgresClient) Close() error {
	client, err := c.lazyClient.Load(context.Background())
	if err != nil {
		return err
	}
	client.Close()
	return nil
}
