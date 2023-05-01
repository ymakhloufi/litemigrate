package migrator

import (
	"fmt"
	"path/filepath"

	"github.com/ymakhloufi/litemigrate/internal/pkg/fsutils"
	"github.com/ymakhloufi/litemigrate/internal/pkg/migration/model"
	"go.uber.org/zap"
)

var ErrDirtyMigrationExists = fmt.Errorf("dirty migration")

type Store interface {
	HasMigrationRun(filename string) (bool, error)
	InsertMigration(filename string) (model.Migration, error)
	RawExec(s string) error
	MarkMigrationCompleted(id uint) (model.Migration, error)
	EnsureMigrationTableExists() error
	GetLatestFailedMigration() (*model.Migration, error)
	Close() error
}

type FSUtils interface {
	GetMigrationFileList(migrationsDir string) (fsutils.DirElements, error)
	ReadFileContent(pathToFile string) (string, error)
}

type Service struct {
	logger        *zap.Logger
	store         Store
	fsUtils       FSUtils
	migrationPath string
}

// New returns a new migrator service
func New(logger *zap.Logger, store Store, migrationPath string, skipDownFiles bool) *Service {
	return &Service{
		logger:        logger,
		store:         store,
		fsUtils:       &fsutils.FsUtils{SkipDownFiles: skipDownFiles},
		migrationPath: migrationPath,
	}
}

func (s *Service) Up() error {
	if err := s.store.EnsureMigrationTableExists(); err != nil {
		return fmt.Errorf("failed to ensure migrations table exists: %w", err)
	}
	if migration, err := s.ensureNoDirtyMigrationsExist(); err != nil {
		return fmt.Errorf("dirty migration %s found: %w", migration.Filename, err)
	}

	// todo: rw-lock migrations table ?
	files, err := s.fsUtils.GetMigrationFileList(s.migrationPath)
	if err != nil {
		return fmt.Errorf("failed to get migration file list in dir %s: %w", s.migrationPath, err)
	}
	for _, file := range files {
		s.logger.Info("running migration", zap.String("filename", file.Name()))

		if wasRun, err := s.wasMigrationPreviouslyRun(file.Name()); err != nil {
			return fmt.Errorf("failed to check if migration %s was previously run: %w", file.Name(), err)
		} else if wasRun {
			s.logger.Info("Skipped: skipping migration, already run", zap.String("filename", file.Name()))
			continue
		}

		s.logger.Info("running migration", zap.String("filename", file.Name()))
		rawSQL, err := s.fsUtils.ReadFileContent(filepath.Join(s.migrationPath, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migration, err := s.runMigration(file.Name(), rawSQL)
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file.Name(), err)
		}

		s.logger.Info("migration has run successfully", zap.Any("migration", migration))
	}

	return nil
}

func (s *Service) Close() (error, error) {
	return nil, s.store.Close() // two errors, to comply with golang-migrate's interface for the migrator's Close() method
}

func (s *Service) runMigration(filename, rawSQL string) (model.Migration, error) {
	migration, err := s.store.InsertMigration(filename)
	if err != nil {
		return model.Migration{}, fmt.Errorf("failed to insert migration into migrations table: %w", err)
	}

	if err := s.store.RawExec(rawSQL); err != nil {
		return model.Migration{}, fmt.Errorf("failed to execute migration: %w", err)
	}

	if migration, err = s.store.MarkMigrationCompleted(migration.ID); err != nil {
		return model.Migration{}, fmt.Errorf("failed to update migrations table: %w", err)
	}

	return migration, nil
}

func (s *Service) wasMigrationPreviouslyRun(filename string) (bool, error) {
	hasRun, err := s.store.HasMigrationRun(filename)
	if err != nil {
		return false, fmt.Errorf("failed to check if migration %s was previously run: %w", filename, err)
	}
	return hasRun, nil
}

func (s *Service) ensureNoDirtyMigrationsExist() (model.Migration, error) {
	migration, err := s.store.GetLatestFailedMigration()
	if err != nil {
		return model.Migration{}, fmt.Errorf("failed to check whether any dirty migrations exist: %w", err)
	}

	if migration != nil {
		return *migration, ErrDirtyMigrationExists
	}

	return model.Migration{}, nil
}
