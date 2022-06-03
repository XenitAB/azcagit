package main

import (
	"io/fs"
	"os"
	"strings"
)

func listYamlFromPath(path string) ([]string, error) {
	fsys := os.DirFS(path)
	return listYamlFromFS(fsys)
}

func listYamlFromFS(fsys fs.FS) ([]string, error) {
	var files []string
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir():
			return nil
		case !strings.HasSuffix(path, ".yaml"):
			return nil
		}

		files = append(files, path)

		return nil
	})

	return files, nil
}
