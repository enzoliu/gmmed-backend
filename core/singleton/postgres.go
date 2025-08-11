package singleton

import (
	"breast-implant-warranty-system/pkg/dbutil"
	"breast-implant-warranty-system/pkg/lazyconn"
	"context"
)

type ReadDBConfigItf interface {
	PostgresReadDBHost() string
	PostgresReadDBPort() string
	PostgresReadDBUser() string
	PostgresReadDBPassword() string
	PostgresDBName() string
}

type WriteDBConfigItf interface {
	PostgresWriteDBHost() string
	PostgresWriteDBPort() string
	PostgresWriteDBUser() string
	PostgresWriteDBPassword() string
	PostgresDBName() string
}

var _ ReadDBConfigItf = (*PostgresDBConfig)(nil)
var _ WriteDBConfigItf = (*PostgresDBConfig)(nil)

type PostgresDBConfig struct {
	POSTGRES_READ_DB_HOST     string `env:"POSTGRES_READ_DB_HOST"`
	POSTGRES_READ_DB_PORT     string `env:"POSTGRES_READ_DB_PORT" envDefault:"5432"`
	POSTGRES_READ_DB_USER     string `env:"POSTGRES_READ_DB_USER" envDefault:"root"`
	POSTGRES_READ_DB_PASSWORD string `env:"POSTGRES_READ_DB_PASSWORD" envSecretPath:"POSTGRES_READ_DB_PASSWORD_SECRET_PATH"`

	POSTGRES_WRITE_DB_HOST     string `env:"POSTGRES_WRITE_DB_HOST"`
	POSTGRES_WRITE_DB_PORT     string `env:"POSTGRES_WRITE_DB_PORT" envDefault:"5432"`
	POSTGRES_WRITE_DB_USER     string `env:"POSTGRES_WRITE_DB_USER" envDefault:"root"`
	POSTGRES_WRITE_DB_PASSWORD string `env:"POSTGRES_WRITE_DB_PASSWORD" envSecretPath:"POSTGRES_WRITE_DB_PASSWORD_SECRET_PATH"`

	POSTGRES_DB_NAME string `env:"POSTGRES_DB_NAME"`
}

func (c *PostgresDBConfig) PostgresReadDBHost() string {
	return c.POSTGRES_READ_DB_HOST
}

func (c *PostgresDBConfig) PostgresReadDBPort() string {
	return c.POSTGRES_READ_DB_PORT
}

func (c *PostgresDBConfig) PostgresReadDBUser() string {
	return c.POSTGRES_READ_DB_USER
}

func (c *PostgresDBConfig) PostgresReadDBPassword() string {
	return c.POSTGRES_READ_DB_PASSWORD
}

func (c *PostgresDBConfig) PostgresWriteDBHost() string {
	return c.POSTGRES_WRITE_DB_HOST
}

func (c *PostgresDBConfig) PostgresWriteDBPort() string {
	return c.POSTGRES_WRITE_DB_PORT
}

func (c *PostgresDBConfig) PostgresWriteDBUser() string {
	return c.POSTGRES_WRITE_DB_USER
}

func (c *PostgresDBConfig) PostgresWriteDBPassword() string {
	return c.POSTGRES_WRITE_DB_PASSWORD
}

func (c *PostgresDBConfig) PostgresDBName() string {
	return c.POSTGRES_DB_NAME
}

func (g *Group) GetReadDB(ctx context.Context, cfg ReadDBConfigItf) dbutil.PgxReaderItf {
	g.readDBOnce.Do(func() {
		g.readDB = lazyconn.NewPostgresClient(ctx, lazyconn.SQLConfig{
			Host:     cfg.PostgresReadDBHost(),
			Port:     cfg.PostgresReadDBPort(),
			User:     cfg.PostgresReadDBUser(),
			Password: cfg.PostgresReadDBPassword(),
			Database: cfg.PostgresDBName(),
		})
	})
	return g.readDB
}

func (g *Group) GetWriteDB(ctx context.Context, cfg WriteDBConfigItf) dbutil.PgxClientItf {
	g.writeDBOnce.Do(func() {
		g.writeDB = lazyconn.NewPostgresClient(ctx, lazyconn.SQLConfig{
			Host:     cfg.PostgresWriteDBHost(),
			Port:     cfg.PostgresWriteDBPort(),
			User:     cfg.PostgresWriteDBUser(),
			Password: cfg.PostgresWriteDBPassword(),
			Database: cfg.PostgresDBName(),
		})
	})
	return g.writeDB
}
