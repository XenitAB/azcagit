package source

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fluxcd/source-controller/pkg/git"
	"github.com/fluxcd/source-controller/pkg/git/libgit2"
	"github.com/fluxcd/source-controller/pkg/git/strategy"
	"github.com/go-logr/logr"
	"github.com/xenitab/azcagit/src/config"
)

type GitSource struct {
	cfg            config.Config
	lastCommitHash string
}

var _ Source = (*GitSource)(nil)

func NewGitSource(cfg config.Config) (*GitSource, error) {
	return &GitSource{
		cfg:            cfg,
		lastCommitHash: "",
	}, nil
}

func (s *GitSource) Get(ctx context.Context) (*SourceApps, error) {
	yamlFiles, err := s.checkout(ctx)
	if err != nil {
		return nil, err
	}

	apps := getSourceAppsFromFiles(yamlFiles, s.cfg)
	return apps, nil
}

func (s *GitSource) checkout(ctx context.Context) (*map[string][]byte, error) {
	log := logr.FromContextOrDiscard(ctx)

	strat, err := strategy.CheckoutStrategyForImplementation(ctx, libgit2.Implementation, git.CheckoutOptions{
		Branch: s.cfg.GitBranch,
	})
	if err != nil {
		return nil, err
	}

	commit, err := strat.Checkout(ctx, s.cfg.CheckoutPath, s.cfg.GitUrl, &git.AuthOptions{
		TransportOptionsURL: s.cfg.GitUrl,
	})
	if err != nil {
		return nil, err
	}

	newCommitHash := string(commit.Hash)
	if newCommitHash != s.lastCommitHash {
		log.Info("new commit hash", "newCommitHash", newCommitHash, "lastCommitHash", s.lastCommitHash)
		s.lastCommitHash = newCommitHash
	}

	yamlPath := filepath.Clean(s.cfg.CheckoutPath)
	if s.cfg.GitYamlPath != "" {
		yamlPath = filepath.Clean(fmt.Sprintf("%s/%s", yamlPath, s.cfg.GitYamlPath))
	}

	yamlFiles, err := listYamlFromPath(yamlPath)
	if err != nil {
		return nil, err
	}

	return yamlFiles, nil
}
