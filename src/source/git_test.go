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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluxcd/pkg/gittestserver"
	git2go "github.com/libgit2/git2go/v33"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/azcagit/src/config"
)

const (
	testFixtureYAML1 = `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
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
apiVersion: aca.xenit.io/v1alpha1
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

	err = server.StartHTTP()
	require.NoError(t, err)
	defer server.StopHTTP()

	repoPath := "test.git"
	defaultBranch := "master"
	tmpDir := t.TempDir()
	err = server.InitRepo(tmpDir, defaultBranch, repoPath)
	require.NoError(t, err)

	repo, err := git2go.OpenRepository(filepath.Join(server.Root(), repoPath))
	require.NoError(t, err)
	defer repo.Free()

	repoURL := server.HTTPAddress() + "/" + repoPath
	sourceClient, err := NewGitSource(config.Config{
		GitUrl:               repoURL,
		GitBranch:            defaultBranch,
		ManagedEnvironmentID: "ze-managed-id",
		Location:             "ze-location",
	})
	require.NoError(t, err)

	firstCommit, err := testCommitFile(t, repo, "foo1.yaml", testFixtureYAML1, time.Now())
	require.NoError(t, err)

	firstSourceApps, firstRevision, err := sourceClient.Get(context.TODO())
	require.NoError(t, err)
	require.Equal(t, firstCommit.String(), firstRevision)
	require.NotNil(t, firstSourceApps)
	require.NoError(t, firstSourceApps.Error())
	require.Len(t, firstSourceApps.GetSortedNames(), 1)

	secondCommit, err := testCommitFile(t, repo, "foo2.yaml", testFixtureYAML2, time.Now())
	require.NoError(t, err)

	secondSourceApps, secondRevision, err := sourceClient.Get(context.TODO())
	require.NoError(t, err)
	require.Equal(t, secondCommit.String(), secondRevision)
	require.NotNil(t, secondSourceApps)
	require.NoError(t, secondSourceApps.Error())
	require.Len(t, secondSourceApps.GetSortedNames(), 2)
}

func testCommitFile(t *testing.T, repo *git2go.Repository, path, content string, time time.Time) (*git2go.Oid, error) {
	t.Helper()

	var parentC []*git2go.Commit
	head, err := testHeadCommit(t, repo)
	if err == nil {
		defer head.Free()
		parentC = append(parentC, head)
	}

	index, err := repo.Index()
	if err != nil {
		return nil, err
	}
	defer index.Free()

	blobOID, err := repo.CreateBlobFromBuffer([]byte(content))
	if err != nil {
		return nil, err
	}

	entry := &git2go.IndexEntry{
		Mode: git2go.FilemodeBlob,
		Id:   blobOID,
		Path: path,
	}

	if err := index.Add(entry); err != nil {
		return nil, err
	}
	if err := index.Write(); err != nil {
		return nil, err
	}

	treeID, err := index.WriteTree()
	if err != nil {
		return nil, err
	}

	tree, err := repo.LookupTree(treeID)
	if err != nil {
		return nil, err
	}
	defer tree.Free()

	c, err := repo.CreateCommit("HEAD", testMockSignature(t, time), testMockSignature(t, time), "Committing "+path, tree, parentC...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func testMockSignature(t *testing.T, time time.Time) *git2go.Signature {
	t.Helper()

	return &git2go.Signature{
		Name:  "Jane Doe",
		Email: "author@example.com",
		When:  time,
	}
}

func testHeadCommit(t *testing.T, repo *git2go.Repository) (*git2go.Commit, error) {
	t.Helper()

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	defer head.Free()
	c, err := repo.LookupCommit(head.Target())
	if err != nil {
		return nil, err
	}
	return c, nil
}
