package main

import (
	"io/fs"
	"os"
	"strings"
)

type YAMLFiles map[string][]byte

func listYamlFromPath(path string) (*YAMLFiles, error) {
	fsys := os.DirFS(path)
	return listYamlFromFS(fsys)
}

func listYamlFromFS(fsys fs.FS) (*YAMLFiles, error) {
	files := make(YAMLFiles)
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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

	return &files, nil
}
