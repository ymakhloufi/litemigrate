package fsutils

import (
	"os"
	"sort"
	"strings"
)

type DirElements []os.DirEntry

func (s DirElements) Len() int           { return len(s) }
func (s DirElements) Less(i, j int) bool { return strings.Compare(s[i].Name(), s[j].Name()) == -1 }
func (s DirElements) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type FsUtils struct {
	SkipDownFiles bool
}

func (s *FsUtils) GetMigrationFileList(dir string) (DirElements, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := DirElements{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			if s.SkipDownFiles && strings.HasSuffix(strings.ToLower(file.Name()), "down.sql") {
				continue
			}
			result = append(result, file)
		}
	}

	sort.Sort(result)

	return result, nil
}

func (s *FsUtils) ReadFileContent(pathToFile string) (string, error) {
	rawSQL, err := os.ReadFile(pathToFile)
	return string(rawSQL), err
}
