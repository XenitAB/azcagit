package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fluxcd/source-controller/pkg/git"
	"github.com/fluxcd/source-controller/pkg/git/libgit2"
	"github.com/fluxcd/source-controller/pkg/git/strategy"
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

	checkoutOpts := git.CheckoutOptions{
		Branch:       s.cfg.GitBranch,
		LastRevision: s.lastRevision,
	}

	strat, err := strategy.CheckoutStrategyForImplementation(ctx, libgit2.Implementation, checkoutOpts)
	if err != nil {
		log.V(1).Error(err, "failed to set checkout strategy", "git_branch", s.cfg.GitBranch)
		return nil, "", err
	}

	commit, err := strat.Checkout(ctx, tmpDir, s.cfg.GitUrl, &git.AuthOptions{
		TransportOptionsURL: s.cfg.GitUrl,
	})
	if err != nil {
		log.V(1).Error(err, "failed to checkout")
		return nil, "", err
	}

	log.V(1).Info("commit data", "ShortMessage", commit.ShortMessage(), "String", commit.String(), "commit", commit)

	revision := string(commit.Hash)
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
