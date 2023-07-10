// Orignial copyright from Flux (commit: f5ada74): https://github.com/fluxcd/source-controller/blob/main/pkg/git/libgit2/checkout_test.go
// /*
// Copyright 2020 The Flux authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package source

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluxcd/pkg/git"
	gg "github.com/fluxcd/pkg/git/gogit"
	"github.com/fluxcd/pkg/git/repository"
	"github.com/fluxcd/pkg/gittestserver"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/azcagit/src/config"
)

const (
	testFixtureYAML1 = `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo1
spec:
  app:
    properties:
      template:
        containers:
        - name: foo
          image: foo:latest
          resources:
            cpu: 0.25
            memory: .5Gi
        scale:
          minReplicas: 1
          maxReplicas: 1`
	testFixtureYAML2 = `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo2
spec:
  app:
    properties:
      template:
        containers:
        - name: foo
          image: foo:latest
          resources:
            cpu: 0.25
            memory: .5Gi
        scale:
          minReplicas: 1
          maxReplicas: 1`
)

func TestGitSource(t *testing.T) {
	server, err := gittestserver.NewTempGitServer()
	require.NoError(t, err)
	defer os.RemoveAll(server.Root())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = server.StartHTTP()
	require.NoError(t, err)
	defer server.StopHTTP()

	repoPath := "test.git"
	defaultBranch := "master"

	// the fixture path can't be empty or it will return an error: clean working tree
	tmpFixtureDir := t.TempDir()
	err = os.WriteFile(filepath.Clean(fmt.Sprintf("%s/foo.txt", tmpFixtureDir)), []byte("test file"), 0600)
	require.NoError(t, err)
	err = server.InitRepo(tmpFixtureDir, defaultBranch, repoPath)
	require.NoError(t, err)

	repoURL := server.HTTPAddress() + "/" + repoPath
	sourceClient, err := NewGitSource(config.Config{
		GitUrl:               repoURL,
		GitBranch:            defaultBranch,
		ManagedEnvironmentID: "ze-managed-id",
		Location:             "ze-location",
	})
	require.NoError(t, err)

	tmp := t.TempDir()
	ggc, err := gg.NewClient(tmp, &git.AuthOptions{
		Transport: git.HTTP,
	})
	require.NoError(t, err)
	defer ggc.Close()

	// an initial clone is required, or else the client won't have a repository and commands will fail
	_, err = ggc.Clone(ctx, repoURL, repository.CloneConfig{})
	require.NoError(t, err)

	firstCommit, err := testCommitFile(t, ctx, ggc, "foo1.yaml", testFixtureYAML1)
	require.NoError(t, err)

	firstSourceApps, firstRevision, err := sourceClient.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, firstCommit, firstRevision)
	require.NotNil(t, firstSourceApps)
	require.NoError(t, firstSourceApps.Error())
	require.Len(t, firstSourceApps.GetSortedNames(), 1)

	secondCommit, err := testCommitFile(t, ctx, ggc, "foo2.yaml", testFixtureYAML2)
	require.NoError(t, err)

	secondSourceApps, secondRevision, err := sourceClient.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, secondCommit, secondRevision)
	require.NotNil(t, secondSourceApps)
	require.NoError(t, secondSourceApps.Error())
	require.Len(t, secondSourceApps.GetSortedNames(), 2)
}

func testCommitFile(t *testing.T, ctx context.Context, ggc *gg.Client, path, content string) (string, error) {
	t.Helper()

	ref, err := ggc.Head()
	require.NoError(t, err)

	_, err = ggc.Commit(
		git.Commit{
			Author: git.Signature{
				Name:  "Jane Doe",
				Email: "author@example.com",
			},
			Message: "testing",
		},
		repository.WithFiles(map[string]io.Reader{
			path: strings.NewReader(content),
		}),
	)
	require.NoError(t, err)

	// the commit needs to be pushed
	err = ggc.Push(ctx, repository.PushConfig{})
	require.NoError(t, err)

	newRef, err := ggc.Head()
	require.NoError(t, err)
	require.NotEqual(t, ref, newRef)

	return newRef, nil
}

func TestRedactGitSecretFromError(t *testing.T) {
	cases := []struct {
		testDescription  string
		gitUrl           string
		inputErrorString string
		expectedResult   string
	}{
		{
			testDescription: "redact secret",
			// secretlint-disable
			gitUrl: "https://foo:supersecret@foobar.net",
			// secretlint-disable
			inputErrorString: "unable to clone https://foo:supersecret@foobar.net",
			// secretlint-disable
			expectedResult: "unable to clone https://foo:redacted@foobar.net",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		inputError := fmt.Errorf(c.inputErrorString)
		result := redactGitSecretFromError(c.gitUrl, inputError)
		require.Equal(t, c.expectedResult, result.Error())
	}

}
