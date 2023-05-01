package store

import (
	rand2 "crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ymakhloufi/light-migrate/internal/pkg/migration/model"
)

// tests for postgresstore.go
// Path: internal/pkg/migration/store/postgresstore.go

var connectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
	"myuser", "mypass", "127.0.0.1", 55432, "mydb", "disable")

const filename = "myFileName"

func TestPostgresStore_EnsureMigrationTableExists(t *testing.T) {
	migrationTableName := "test_migration_" + randomString(10)
	conn, err := newPgConnection(nil, connectionString)
	require.NoError(t, err)

	store := &PostgresStore{SQLStore: SQLStore{conn: conn, tableName: migrationTableName}}

	_, err = store.GetLatestFailedMigration()
	require.Error(t, err)

	err = store.EnsureMigrationTableExists()
	require.NoError(t, err)

	_, err = store.GetLatestFailedMigration()
	require.NoError(t, err)

	_, err = store.conn.Exec("DROP TABLE " + migrationTableName)
	require.NoError(t, err)
}

func TestPostgresStore_InsertMigration(t *testing.T) {
	pg := makeTestStoreWithEphemeralTable(t)

	migration, err := pg.InsertMigration(filename)
	require.NoError(t, err)

	require.Equal(t, filename, migration.Filename)
	require.NotZero(t, migration.ID)
	require.NotZero(t, migration.StartedAt)
	require.Nil(t, migration.CompletedAt)

	migrationFromDB := pg.getMigrationByID(t, migration.ID)
	require.Equal(t, migration, migrationFromDB)
}

func TestPostgresStore_MarkMigrationCompleted(t *testing.T) {
	pg := makeTestStoreWithEphemeralTable(t)

	insertedMigration, err := pg.InsertMigration(filename)
	require.Equal(t, filename, insertedMigration.Filename)
	require.NoError(t, err)

	fetchedMigrationAfterInsertion := pg.getMigrationByID(t, insertedMigration.ID)
	require.Equal(t, filename, fetchedMigrationAfterInsertion.Filename)
	require.Nil(t, fetchedMigrationAfterInsertion.CompletedAt)

	completedMigration, err := pg.MarkMigrationCompleted(insertedMigration.ID)
	require.NoError(t, err)
	require.Equal(t, filename, completedMigration.Filename)
	require.NotNil(t, completedMigration.CompletedAt)

	fetchedMigrationAfterCompletion := pg.getMigrationByID(t, insertedMigration.ID)
	require.Equal(t, completedMigration, fetchedMigrationAfterCompletion)
}

func TestPostgresStore_GetLatestFailedMigration(t *testing.T) {
	pg := makeTestStoreWithEphemeralTable(t)

	migration, err := pg.InsertMigration("myFileName")
	require.NoError(t, err)

	latestFailedMigration, err := pg.GetLatestFailedMigration()
	require.NoError(t, err)
	require.NotNil(t, latestFailedMigration)
	require.Equal(t, migration.ID, latestFailedMigration.ID)

	_, err = pg.MarkMigrationCompleted(latestFailedMigration.ID)
	require.NoError(t, err)

	latestFailedMigration, err = pg.GetLatestFailedMigration()
	require.NoError(t, err)
	require.Nil(t, latestFailedMigration)
}

func TestPostgresStore_HasMigrationRun(t *testing.T) {
	pg := makeTestStoreWithEphemeralTable(t)

	hasMigrationRun, err := pg.HasMigrationRun(filename)
	require.NoError(t, err)
	require.False(t, hasMigrationRun)

	_, err = pg.InsertMigration(filename)
	require.NoError(t, err)

	hasMigrationRun, err = pg.HasMigrationRun(filename)
	require.NoError(t, err)
	require.True(t, hasMigrationRun)
}

func randomString(length int) string {
	b := make([]byte, length+2)
	_, _ = rand2.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

func makeTestStoreWithEphemeralTable(t *testing.T) *PostgresStore {
	migrationTableName := "test_migration_" + randomString(16)

	conn, err := newPgConnection(nil, connectionString)
	require.NoError(t, err)

	pg := &PostgresStore{SQLStore: SQLStore{conn: conn, tableName: migrationTableName}}
	err = pg.EnsureMigrationTableExists()
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := conn.Exec("DROP TABLE " + migrationTableName)
		require.NoError(t, err)
	})

	return pg
}

func (pg *PostgresStore) getMigrationByID(t *testing.T, id uint) model.Migration {
	var migration model.Migration
	err := pg.conn.
		QueryRow("SELECT id, filename, started_at, completed_at FROM "+pg.tableName+" WHERE id = $1", id).
		Scan(&migration.ID, &migration.Filename, &migration.StartedAt, &migration.CompletedAt)

	require.NoError(t, err)

	return migration
}
