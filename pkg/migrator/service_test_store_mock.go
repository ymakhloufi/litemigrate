package migrator

import (
	"github.com/ymakhloufi/light-migrate/internal/pkg/migration/model"
)

type storeMock struct {
	hasMigrationRunCalls            uint
	insertMigrationCalls            uint
	rawExecCalls                    uint
	markMigrationCompletedCalls     uint
	ensureMigrationTableExistsCalls uint
	getLatestFailedMigrationCalls   uint

	HasMigrationRunFunc            func(filename string) (bool, error)
	InsertMigrationFunc            func(filename string) (model.Migration, error)
	RawExecFunc                    func(rawSql string) error
	MarkMigrationCompletedFunc     func(id uint) (model.Migration, error)
	EnsureMigrationTableExistsFunc func() error
	GetLatestFailedMigrationFunc   func() (*model.Migration, error)
}

func (s *storeMock) Close() error {
	return nil
}

func (s *storeMock) HasMigrationRun(filename string) (bool, error) {
	s.hasMigrationRunCalls++
	return s.HasMigrationRunFunc(filename)
}

func (s *storeMock) InsertMigration(filename string) (model.Migration, error) {
	s.insertMigrationCalls++
	return s.InsertMigrationFunc(filename)
}

func (s *storeMock) RawExec(rawSQL string) error {
	s.rawExecCalls++
	return s.RawExecFunc(rawSQL)
}

func (s *storeMock) MarkMigrationCompleted(id uint) (model.Migration, error) {
	s.markMigrationCompletedCalls++
	return s.MarkMigrationCompletedFunc(id)
}

func (s *storeMock) EnsureMigrationTableExists() error {
	s.ensureMigrationTableExistsCalls++
	return s.EnsureMigrationTableExistsFunc()
}

func (s *storeMock) GetLatestFailedMigration() (*model.Migration, error) {
	s.getLatestFailedMigrationCalls++
	return s.GetLatestFailedMigrationFunc()
}
