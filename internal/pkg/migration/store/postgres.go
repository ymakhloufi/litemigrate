package store

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

type PostgresStore struct {
	SQLStore
}

func NewPostgresStore(logger *zap.Logger, migrationTableName, connectionString string) (*PostgresStore, error) {
	conn, err := newPgConnection(logger, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create new postgres connection: %w", err)
	}

	store := &PostgresStore{
		SQLStore: SQLStore{
			conn:      conn,
			tableName: migrationTableName},
	}
	return store, nil
}

func newPgConnection(logger *zap.Logger, connectionString string) (conn *sql.DB, err error) {
	config, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	if logger != nil {
		config.Logger = zapadapter.NewLogger(logger)
	}

	return stdlib.OpenDB(*config), nil
}

func (pg *PostgresStore) EnsureMigrationTableExists() error {
	qry := `CREATE TABLE IF NOT EXISTS ` + pg.tableName + ` (
		    id serial primary key, 
		    filename text unique not null, 
		    started_at timestamp not null default now(), 
		    completed_at timestamp
		)`
	_, err := pg.conn.Exec(qry)
	return err
}
