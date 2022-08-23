package source

import (
	"io/fs"
	"os"
	"strings"
)

func listYamlFromPath(path string) (*map[string][]byte, error) {
	fsys := os.DirFS(path)
	return listYamlFromFS(fsys)
}

func listYamlFromFS(fsys fs.FS) (*map[string][]byte, error) {
	files := make(map[string][]byte)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir():
			return nil
		case !strings.HasSuffix(path, ".yaml"):
			return nil
		}

		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		files[path] = b

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &files, nil
}
