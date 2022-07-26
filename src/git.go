package main

import (
	"context"
	"fmt"

	"github.com/fluxcd/source-controller/pkg/git"
	"github.com/fluxcd/source-controller/pkg/git/libgit2"
	"github.com/fluxcd/source-controller/pkg/git/strategy"
)

func checkout(ctx context.Context, path string, url string, branch string, lastHash string) (*YAMLFiles, string, error) {
	strat, err := strategy.CheckoutStrategyForImplementation(ctx, libgit2.Implementation, git.CheckoutOptions{
		Branch: branch,
	})
	if err != nil {
		return nil, "", err
	}

	fmt.Println("Got strategy")

	commit, err := strat.Checkout(ctx, path, url, &git.AuthOptions{})
	if err != nil {
		return nil, "", err
	}

	fmt.Println("Checkout complete")

	yamlFiles, err := listYamlFromPath(path)
	if err != nil {
		return nil, "", err
	}

	return yamlFiles, commit.Hash.String(), nil
}
