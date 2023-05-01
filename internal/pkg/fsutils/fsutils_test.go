package fsutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_getMigrationFileList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		existingFiles       []string
		existingDirectories []string
		want                []string
	}{
		{
			name: "returns empty list if directory is empty",
			want: []string{},
		},
		{
			name:          "doesn't return files without .sql suffix",
			existingFiles: []string{"foo.other", "bar.sql", "baz.sql", "quo.sql.other"},
			want:          []string{"bar.sql", "baz.sql"},
		},
		{
			name:                "doesn't return folders",
			existingFiles:       []string{"foo.sql"},
			existingDirectories: []string{"bar.sql"},
			want:                []string{"foo.sql"},
		},
		{
			name:          "returns files in sorted order",
			existingFiles: []string{"bbb.sql", "ccc.sql", "aaa.sql", "qqq.sql", "111.sql"},
			want:          []string{"111.sql", "aaa.sql", "bbb.sql", "ccc.sql", "qqq.sql"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// create temp dir to test in
			dir, err := os.MkdirTemp("", "")
			defer func() { _ = os.RemoveAll(dir) }()
			require.NoError(t, err)

			// create existing directories
			for _, v := range tt.existingDirectories {
				err := os.Mkdir(filepath.Join(dir, v), 0777)
				require.NoError(t, err)
			}

			// create existing files
			for _, v := range tt.existingFiles {
				_, err := os.Create(filepath.Join(dir, v))
				require.NoError(t, err)
			}

			s := &FsUtils{}
			got, err := s.GetMigrationFileList(dir)
			require.NoError(t, err)

			gotStrings := []string{}
			for _, v := range got {
				gotStrings = append(gotStrings, v.Name())
			}

			require.Equal(t, tt.want, gotStrings)
		})
	}
}
