package store

import (
	"database/sql"
	"errors"

	"github.com/ymakhloufi/litemigrate/internal/pkg/migration/model"
)

type SQLStore struct {
	conn      *sql.DB
	tableName string
}

func (s *SQLStore) Close() error {
	return s.conn.Close()
}

func (s *SQLStore) HasMigrationRun(filename string) (bool, error) {
	qry := `SELECT EXISTS (
    			SELECT 1 
    			FROM ` + s.tableName + ` 
    			WHERE filename = $1
    		)`
	row := s.conn.QueryRow(qry, filename)

	var exists bool
	err := row.Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return exists, nil
}

func (s *SQLStore) InsertMigration(filename string) (model.Migration, error) {
	qry := `INSERT INTO ` + s.tableName + ` (filename) 
		VALUES ($1) 
		RETURNING id, filename, started_at, completed_at`
	row := s.conn.QueryRow(qry, filename)

	result := model.Migration{}
	err := row.Scan(&result.ID, &result.Filename, &result.StartedAt, &result.CompletedAt)
	return result, err
}

func (s *SQLStore) MarkMigrationCompleted(id uint) (model.Migration, error) {
	qry := `UPDATE ` + s.tableName + `
		SET completed_at = now() 
		WHERE id = $1
		RETURNING id, filename, started_at, completed_at`
	row := s.conn.QueryRow(qry, id)

	result := model.Migration{}
	err := row.Scan(&result.ID, &result.Filename, &result.StartedAt, &result.CompletedAt)
	return result, err
}

func (s *SQLStore) GetLatestFailedMigration() (*model.Migration, error) {
	qry := `SELECT id, filename, started_at 
		FROM ` + s.tableName + `
		WHERE completed_at IS NULL 
		ORDER BY id DESC 
		LIMIT 1`
	row := s.conn.QueryRow(qry)

	latestFailedMigration := model.Migration{}
	err := row.Scan(&latestFailedMigration.ID, &latestFailedMigration.Filename, &latestFailedMigration.StartedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // all good, no dirty migrations found
		}
		return nil, err
	}
	return &latestFailedMigration, nil
}

func (s *SQLStore) RawExec(rawSQL string) error {
	_, err := s.conn.Exec(rawSQL)
	return err
}
