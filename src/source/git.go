package source

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fluxcd/pkg/git"
	"github.com/fluxcd/pkg/git/gogit"
	"github.com/fluxcd/pkg/git/repository"
	"github.com/go-logr/logr"
	"github.com/xenitab/azcagit/src/config"
)

type GitSource struct {
	cfg          config.Config
	lastRevision string
}

var _ Source = (*GitSource)(nil)

func NewGitSource(cfg config.Config) (*GitSource, error) {
	return &GitSource{
		cfg:          cfg,
		lastRevision: "",
	}, nil
}

func (s *GitSource) Get(ctx context.Context) (*SourceApps, string, error) {
	yamlFiles, revision, err := s.checkout(ctx)
	if err != nil {
		return nil, "", err
	}

	apps := getSourceAppsFromFiles(yamlFiles, s.cfg)
	return apps, revision, nil
}

func (s *GitSource) checkout(ctx context.Context) (*map[string][]byte, string, error) {
	log := logr.FromContextOrDiscard(ctx)

	tmpDir, tmpDirCleanup, err := createTemporaryDirectory(ctx, s.cfg.CheckoutPath)
	if err != nil {
		return nil, "", err
	}

	defer tmpDirCleanup()

	gitUrl, err := url.Parse(s.cfg.GitUrl)
	if err != nil {
		log.V(1).Error(err, "failed to parse git url")
		return nil, "", err
	}

	authOpts, err := git.NewAuthOptions(*gitUrl, nil)
	if err != nil {
		log.V(1).Error(err, "failed to parse auth options")
		return nil, "", err
	}

	clientOpts := []gogit.ClientOption{gogit.WithDiskStorage()}
	if authOpts.Transport == git.HTTP {
		clientOpts = append(clientOpts, gogit.WithInsecureCredentialsOverHTTP())
	}

	gitReader, err := gogit.NewClient(tmpDir, authOpts, clientOpts...)
	if err != nil {
		log.V(1).Error(err, "failed to create git client")
		return nil, "", err
	}
	defer gitReader.Close()

	cloneOpts := repository.CloneOptions{
		ShallowClone:      true,
		RecurseSubmodules: true,
		CheckoutStrategy: repository.CheckoutStrategy{
			Branch: s.cfg.GitBranch,
		},
	}
	commit, err := gitReader.Clone(ctx, s.cfg.GitUrl, cloneOpts)
	if err != nil {
		log.V(1).Error(err, "failed to clone")
		return nil, "", err
	}

	log.V(1).Info("commit data", "ShortMessage", commit.ShortMessage(), "String", commit.String(), "commit", commit)

	revision := commit.Hash.String()
	log.V(1).Info("current revision", "revision", revision)
	if revision != s.lastRevision {
		log.Info("new commit hash", "new_revision", revision, "last_revision", s.lastRevision)
		s.lastRevision = revision
	}

	yamlPath := filepath.Clean(tmpDir)
	if s.cfg.GitYamlPath != "" {
		yamlPath = filepath.Clean(fmt.Sprintf("%s/%s", yamlPath, s.cfg.GitYamlPath))
	}

	yamlFiles, err := listYamlFromPath(yamlPath)
	if err != nil {
		log.V(1).Error(err, "failed to list yamls from path", "yaml_path", yamlPath)
		return nil, revision, err
	}

	return yamlFiles, revision, nil
}

func createTemporaryDirectory(ctx context.Context, path string) (string, func(), error) {
	log := logr.FromContextOrDiscard(ctx)

	tmpDir, err := os.MkdirTemp(path, "azcagit*")
	if err != nil {
		log.V(1).Error(err, "unable to create temporary working directory", "checkout_path", path)
		return "", nil, err
	}
	cleanup := func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Error(err, "failed to remove temporary working directory", "tmp_dir", tmpDir)
		}
	}

	return tmpDir, cleanup, nil
}
