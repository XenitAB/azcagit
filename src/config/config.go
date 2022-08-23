package config

import (
	"github.com/alexflint/go-arg"
)

type Config struct {
	ResourceGroupName    string `arg:"-g,--resource-group-name,env:RESOURCE_GROUP_NAME,required" help:"Azure Resource Group Name"`
	SubscriptionID       string `arg:"-s,--subscription-id,env:AZURE_SUBSCRIPTION_ID,required" help:"Azure Subscription ID"`
	ManagedEnvironmentID string `arg:"-m,--managed-environment-id,env:MANAGED_ENVIRONMENT_ID,required" help:"Azure Container Apps Managed Environment ID"`
	Location             string `arg:"-l,--location,env:LOCATION,required" help:"Azure Region (location)"`
	ReconcileInterval    string `arg:"-i,--reconcile-interval,env:RECONCILE_INTERVAL" default:"5m" help:"The interval between reconciles"`
	CheckoutPath         string `arg:"-c,--checkout-path,env:CHECKOUT_PATH,required" help:"The local path where the git repository should be checked out"`
	GitUrl               string `arg:"-u,--git-url,env:GIT_URL,required" help:"The git url to checkout"`
	GitBranch            string `arg:"-b,--git-branch,env:GIT_BRANCH" default:"main" help:"The git branch to checkout"`
	GitYamlPath          string `arg:"--git-yaml-path,env:GIT_YAML_ROOT" default:"" help:"The path where the yaml files are located"`
}

func NewConfig(args []string) (Config, error) {
	cfg := Config{}

	parser, err := arg.NewParser(arg.Config{
		Program:   "aca-gitops-engine",
		IgnoreEnv: false,
	}, &cfg)
	if err != nil {
		return Config{}, err
	}

	err = parser.Parse(args)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
