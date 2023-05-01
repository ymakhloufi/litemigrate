package migrator

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ymakhloufi/litemigrate/internal/pkg/fsutils"
	"github.com/ymakhloufi/litemigrate/internal/pkg/migration/model"
	"go.uber.org/zap"
)

func TestService_runMigration(t *testing.T) {
	myErr := errors.New("my error")

	type args struct {
		filename string
		rawSQL   string
	}
	tests := []struct {
		name           string
		args           args
		store          Store
		wantStoreCalls *storeMock
		want           model.Migration
		wantErr        error
	}{
		{
			name: "happy path",
			args: args{
				filename: "myFilename",
				rawSQL:   "myRawSQL",
			},
			store: &storeMock{
				InsertMigrationFunc: func(filename string) (model.Migration, error) {
					require.Equal(t, "myFilename", filename)
					return model.Migration{ID: uint(1234)}, nil
				},
				RawExecFunc: func(rawSql string) error {
					require.Equal(t, "myRawSQL", rawSql)
					return nil
				},
				MarkMigrationCompletedFunc: func(id uint) (model.Migration, error) {
					require.Equal(t, uint(1234), id)
					return model.Migration{ID: uint(1234), Filename: "abc"}, nil
				},
			},
			want: model.Migration{ID: uint(1234), Filename: "abc"},
			wantStoreCalls: &storeMock{
				insertMigrationCalls:        1,
				rawExecCalls:                1,
				markMigrationCompletedCalls: 1,
			},
		},
		{
			name: "doesn't call rawExec when insertion fails",
			store: &storeMock{
				InsertMigrationFunc: func(filename string) (model.Migration, error) {
					return model.Migration{}, myErr
				},
				RawExecFunc: func(rawSql string) error {
					require.FailNow(t, "should not have called markCompleted")
					return nil
				},
				MarkMigrationCompletedFunc: func(id uint) (model.Migration, error) {
					require.FailNow(t, "should not have called markCompleted")
					return model.Migration{}, nil
				}},
			wantStoreCalls: &storeMock{
				insertMigrationCalls:          1,
				rawExecCalls:                  0,
				getLatestFailedMigrationCalls: 0,
			},
			wantErr: myErr,
		},
		{
			name: "doesn't call markCompleted when execution fails",
			store: &storeMock{
				InsertMigrationFunc: func(filename string) (model.Migration, error) {
					return model.Migration{ID: uint(1234)}, nil
				},
				RawExecFunc: func(rawSql string) error {
					return myErr
				},
				MarkMigrationCompletedFunc: func(id uint) (model.Migration, error) {
					require.FailNow(t, "should not have called markCompleted")
					return model.Migration{}, nil
				}},
			wantStoreCalls: &storeMock{
				insertMigrationCalls:          1,
				rawExecCalls:                  1,
				getLatestFailedMigrationCalls: 0,
			},
			wantErr: myErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{logger: zap.NewNop(), store: tt.store}
			got, err := s.runMigration(tt.args.filename, tt.args.rawSQL)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)

			require.Equal(t, tt.wantStoreCalls.getLatestFailedMigrationCalls, tt.store.(*storeMock).getLatestFailedMigrationCalls)
			require.Equal(t, tt.wantStoreCalls.insertMigrationCalls, tt.store.(*storeMock).insertMigrationCalls)
			require.Equal(t, tt.wantStoreCalls.hasMigrationRunCalls, tt.store.(*storeMock).hasMigrationRunCalls)
			require.Equal(t, tt.wantStoreCalls.rawExecCalls, tt.store.(*storeMock).rawExecCalls)
			require.Equal(t, tt.wantStoreCalls.markMigrationCompletedCalls, tt.store.(*storeMock).markMigrationCompletedCalls)
			require.Equal(t, tt.wantStoreCalls.ensureMigrationTableExistsCalls, tt.store.(*storeMock).ensureMigrationTableExistsCalls)
		})
	}
}

func TestService_ensureNoDirtyMigrationsExist(t *testing.T) {
	myErr := errors.New("my error")

	tests := []struct {
		name      string
		store     Store
		wantModel model.Migration
		wantErr   error
	}{
		{
			name: "happy path",
			store: &storeMock{
				GetLatestFailedMigrationFunc: func() (*model.Migration, error) {
					return nil, nil
				},
			},
			wantModel: model.Migration{},
			wantErr:   nil,
		},
		{
			name: "returns error when GetLatestFailedMigration returns error",
			store: &storeMock{
				GetLatestFailedMigrationFunc: func() (*model.Migration, error) {
					return nil, myErr
				},
			},
			wantModel: model.Migration{},
			wantErr:   myErr,
		},
		{
			name: "returns error when GetLatestFailedMigration returns a migration",
			store: &storeMock{
				GetLatestFailedMigrationFunc: func() (*model.Migration, error) {
					return &model.Migration{ID: uint(1234)}, nil
				},
			},
			wantModel: model.Migration{ID: uint(1234)},
			wantErr:   ErrDirtyMigrationExists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{logger: zap.NewNop(), store: tt.store}
			migration, err := s.ensureNoDirtyMigrationsExist()
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.wantModel, migration)
		})
	}
}

func TestService_Run(t *testing.T) {
	myErr := errors.New("my error")
	myDir := "myDir"

	tests := []struct {
		name           string
		store          Store
		wantStoreCalls *storeMock
		fsUtils        FSUtils
		wantErr        error
	}{
		{
			name: "happy path",
			store: &storeMock{
				EnsureMigrationTableExistsFunc: func() error { return nil },
				GetLatestFailedMigrationFunc:   func() (*model.Migration, error) { return nil, nil },
				InsertMigrationFunc:            func(filename string) (model.Migration, error) { return model.Migration{ID: uint(1234)}, nil },
				HasMigrationRunFunc:            func(filename string) (bool, error) { return false, nil },
				RawExecFunc: func(rawSql string) error {
					require.Equal(t, "select * from foo", rawSql)
					return nil
				},
				MarkMigrationCompletedFunc: func(id uint) (model.Migration, error) {
					require.Equal(t, uint(1234), id)
					return model.Migration{}, nil
				},
			},
			wantStoreCalls: &storeMock{
				ensureMigrationTableExistsCalls: 1,
				getLatestFailedMigrationCalls:   1,
				insertMigrationCalls:            2,
				rawExecCalls:                    2,
				markMigrationCompletedCalls:     2,
				hasMigrationRunCalls:            2,
			},
			fsUtils: &fsUtilsMock{
				GetMigrationFileListFunc: func(dir string) (fsutils.DirElements, error) {
					require.Equal(t, myDir, dir)
					return []os.DirEntry{
						fakeDirElement{name: "1.sql", isDir: false},
						fakeDirElement{name: "2.sql", isDir: false},
					}, nil
				},
				ReadFileContentFunc: func(filename string) (string, error) { return "select * from foo", nil },
			},
		},
		{
			name:           "returns error when ensureMigrationTableExists returns error",
			store:          &storeMock{EnsureMigrationTableExistsFunc: func() error { return myErr }},
			wantStoreCalls: &storeMock{ensureMigrationTableExistsCalls: 1},
			wantErr:        myErr,
		},
		{
			name: "returns error when GetLatestFailedMigration returns error",
			store: &storeMock{
				EnsureMigrationTableExistsFunc: func() error { return nil },
				GetLatestFailedMigrationFunc:   func() (*model.Migration, error) { return nil, myErr },
			},
			wantStoreCalls: &storeMock{ensureMigrationTableExistsCalls: 1, getLatestFailedMigrationCalls: 1},
			wantErr:        myErr,
		},
		{
			name: "returns error when GetLatestFailedMigration returns a migration",
			store: &storeMock{
				EnsureMigrationTableExistsFunc: func() error { return nil },
				GetLatestFailedMigrationFunc:   func() (*model.Migration, error) { return &model.Migration{ID: uint(1234)}, nil },
			},
			wantStoreCalls: &storeMock{ensureMigrationTableExistsCalls: 1, getLatestFailedMigrationCalls: 1},
			wantErr:        ErrDirtyMigrationExists,
		},
		{
			name: "returns error when GetMigrationFileList returns error",
			store: &storeMock{
				EnsureMigrationTableExistsFunc: func() error { return nil },
				GetLatestFailedMigrationFunc:   func() (*model.Migration, error) { return nil, nil },
			},
			fsUtils: &fsUtilsMock{GetMigrationFileListFunc: func(dir string) (fsutils.DirElements, error) { return nil, myErr }},
			wantStoreCalls: &storeMock{
				ensureMigrationTableExistsCalls: 1,
				getLatestFailedMigrationCalls:   1,
			},
			wantErr: myErr,
		},
		{
			name: "skips migration when it has already run",
			store: &storeMock{
				EnsureMigrationTableExistsFunc: func() error { return nil },
				GetLatestFailedMigrationFunc:   func() (*model.Migration, error) { return nil, nil },
				HasMigrationRunFunc:            func(filename string) (bool, error) { return filename == "2.sql", nil },
				RawExecFunc:                    func(rawSql string) error { return nil },
				MarkMigrationCompletedFunc:     func(id uint) (model.Migration, error) { return model.Migration{}, nil },
				InsertMigrationFunc: func(filename string) (model.Migration, error) {
					require.NotEqual(t, "2.sql", filename) // only the first one should run
					return model.Migration{}, nil
				},
			},
			fsUtils: &fsUtilsMock{
				GetMigrationFileListFunc: func(dir string) (fsutils.DirElements, error) {
					require.Equal(t, myDir, dir)
					list := []os.DirEntry{
						fakeDirElement{name: "1.sql", isDir: false},
						fakeDirElement{name: "2.sql", isDir: false},
						fakeDirElement{name: "3.sql", isDir: false},
					}
					return list, nil
				},
				ReadFileContentFunc: func(filename string) (string, error) { return "select * from foo", nil },
			},
			wantStoreCalls: &storeMock{
				ensureMigrationTableExistsCalls: 1,
				getLatestFailedMigrationCalls:   1,
				hasMigrationRunCalls:            3,
				insertMigrationCalls:            2,
				rawExecCalls:                    2,
				markMigrationCompletedCalls:     2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:        zap.NewNop(),
				store:         tt.store,
				fsUtils:       tt.fsUtils,
				migrationPath: myDir,
			}
			err := s.Up()
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.wantStoreCalls.getLatestFailedMigrationCalls, tt.store.(*storeMock).getLatestFailedMigrationCalls)
			require.Equal(t, tt.wantStoreCalls.insertMigrationCalls, tt.store.(*storeMock).insertMigrationCalls)
			require.Equal(t, tt.wantStoreCalls.hasMigrationRunCalls, tt.store.(*storeMock).hasMigrationRunCalls)
			require.Equal(t, tt.wantStoreCalls.rawExecCalls, tt.store.(*storeMock).rawExecCalls)
			require.Equal(t, tt.wantStoreCalls.markMigrationCompletedCalls, tt.store.(*storeMock).markMigrationCompletedCalls)
			require.Equal(t, tt.wantStoreCalls.ensureMigrationTableExistsCalls, tt.store.(*storeMock).ensureMigrationTableExistsCalls)
		})
	}
}
