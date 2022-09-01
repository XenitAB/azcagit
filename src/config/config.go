package config

import (
	"net/url"

	"github.com/alexflint/go-arg"
)

type Config struct {
	ResourceGroupName    string `json:"resource_group_name" arg:"-g,--resource-group-name,env:RESOURCE_GROUP_NAME,required" help:"Azure Resource Group Name"`
	SubscriptionID       string `json:"subscription_id" arg:"-s,--subscription-id,env:AZURE_SUBSCRIPTION_ID,required" help:"Azure Subscription ID"`
	ManagedEnvironmentID string `json:"managed_environment_id" arg:"-m,--managed-environment-id,env:MANAGED_ENVIRONMENT_ID,required" help:"Azure Container Apps Managed Environment ID"`
	KeyVaultName         string `json:"key_vault_name" arg:"-k,--key-vault-name,env:KEY_VAULT_NAME,required" help:"Azure KeyVault name to extract secrets from"`
	Location             string `json:"location" arg:"-l,--location,env:LOCATION,required" help:"Azure Region (location)"`
	ReconcileInterval    string `json:"reconcile_interval" arg:"-i,--reconcile-interval,env:RECONCILE_INTERVAL" default:"5m" help:"The interval between reconciles"`
	CheckoutPath         string `json:"checkout_path" arg:"-c,--checkout-path,env:CHECKOUT_PATH,required" help:"The local path where the git repository should be checked out"`
	GitUrl               string `json:"git_url" arg:"-u,--git-url,env:GIT_URL,required" help:"The git url to checkout"`
	GitBranch            string `json:"git_branch" arg:"-b,--git-branch,env:GIT_BRANCH" default:"main" help:"The git branch to checkout"`
	GitYamlPath          string `json:"git_yaml_path" arg:"--git-yaml-path,env:GIT_YAML_ROOT" default:"" help:"The path where the yaml files are located"`
	DaprAppPort          int    `json:"dapr_app_port" arg:"--dapr-app-port,env:DAPR_APP_PORT" default:"8080" help:"The port Dapr service should listen to"`
	DaprPubsubName       string `json:"dapr_pubsub_name" arg:"--dapr-pubsub-name,env:DAPR_PUBSUB_NAME" default:"azcagit-trigger" help:"The PubSub name for the trigger"`
	DaprTopic            string `json:"dapr_topic" arg:"--dapr-topic-name,env:DAPR_TOPIC_NAME" default:"azcagit_trigger" help:"The PubSub topic name for the trigger"`
}

func (cfg *Config) Redacted() Config {
	if cfg == nil {
		return Config{}
	}

	redactedCfg := *cfg
	redactedCfg.GitUrl = redactUrl(redactedCfg.GitUrl)

	return redactedCfg
}

func redactUrl(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}

	_, ok := parsed.User.Password()
	if ok {
		parsed.User = url.UserPassword(parsed.User.Username(), "redacted")
	}

	if !ok && parsed.User.Username() != "" {
		parsed.User = url.User("redacted")
	}

	return parsed.String()
}

func NewConfig(args []string) (Config, error) {
	cfg := Config{}

	parser, err := arg.NewParser(arg.Config{
		Program:   "azcagit",
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
