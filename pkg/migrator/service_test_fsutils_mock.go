package migrator

import (
	"io/fs"

	"github.com/ymakhloufi/light-migrate/internal/pkg/fsutils"
)

type fsUtilsMock struct {
	getMigrationFileListCalls uint
	readFileContentCalls      uint

	GetMigrationFileListFunc func(dir string) (fsutils.DirElements, error)
	ReadFileContentFunc      func(pathToFile string) (string, error)
}

func (f *fsUtilsMock) GetMigrationFileList(dir string) (fsutils.DirElements, error) {
	f.getMigrationFileListCalls++
	return f.GetMigrationFileListFunc(dir)
}

func (f *fsUtilsMock) ReadFileContent(pathToFile string) (string, error) {
	f.readFileContentCalls++
	return f.ReadFileContentFunc(pathToFile)
}

// fakeDirElement is a mock implementation of os.DirEntry (which in turn is an alias for fs.DirEntry)
type fakeDirElement struct {
	name  string
	isDir bool
}

func (f fakeDirElement) Name() string {
	return f.name
}

func (f fakeDirElement) IsDir() bool {
	return f.isDir
}

func (f fakeDirElement) Type() fs.FileMode {
	return 0
}

func (f fakeDirElement) Info() (fs.FileInfo, error) {
	return nil, nil
}
