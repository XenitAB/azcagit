package main

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestListYamlFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"foo.yaml": {
			Data: []byte("foo"),
		},
		"bar.yaml": {
			Data: []byte("bar"),
		},
		"baz/foobar.yaml": {
			Data: []byte("foobar"),
		},
		"a.txt": {
			Data: []byte("a"),
		},
		"b/c.txt": {
			Data: []byte("c"),
		},
	}
	yamlFiles, err := listYamlFromFS(fsys)
	require.NoError(t, err)
	require.Len(t, yamlFiles, 3)
}
