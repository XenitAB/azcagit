package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/alexflint/go-arg"
)

type ReconcileConfig struct {
	ResourceGroupName                  string `json:"resource_group_name" arg:"-g,--resource-group-name,env:RESOURCE_GROUP_NAME,required" help:"Azure Resource Group Name"`
	Environment                        string `json:"environment" arg:"--environment,env:ENVIRONMENT,required" help:"The current environment that azcagit is running in"`
	SubscriptionID                     string `json:"subscription_id" arg:"-s,--subscription-id,env:AZURE_SUBSCRIPTION_ID,required" help:"Azure Subscription ID"`
	ManagedEnvironmentID               string `json:"managed_environment_id" arg:"-m,--managed-environment-id,env:MANAGED_ENVIRONMENT_ID,required" help:"Azure Container Apps Managed Environment ID"`
	KeyVaultName                       string `json:"key_vault_name" arg:"-k,--key-vault-name,env:KEY_VAULT_NAME,required" help:"Azure KeyVault name to extract secrets from"`
	OwnContainerJobName                string `json:"own_container_job_name" arg:"--own-container-job-name,env:OWN_CONTAINER_JOB_NAME" default:"azcagit-reconcile" help:"The name of the Container App job that is running azcagit"`
	OwnResourceGroupName               string `json:"own_resource_group" arg:"--own-resource-group-name,env:OWN_RESOURCE_GROUP_NAME,required" help:"The name of the resource group that the azcagit Container App is located in"`
	ContainerRegistryServer            string `json:"container_registry_server" arg:"--container-registry-server,env:CONTAINER_REGISTRY_SERVER" help:"The container registry server"`
	ContainerRegistryUsername          string `json:"container_registry_username" arg:"--container-registry-username,env:CONTAINER_REGISTRY_USERNAME" help:"The container registry username"`
	ContainerRegistryPassword          string `json:"container_registry_password" arg:"--container-registry-password,env:CONTAINER_REGISTRY_PASSWORD" help:"The container registry password"`
	Location                           string `json:"location" arg:"-l,--location,env:LOCATION,required" help:"Azure Region (location)"`
	CheckoutPath                       string `json:"checkout_path" arg:"-c,--checkout-path,env:CHECKOUT_PATH" default:"/tmp" help:"The local path where the git repository should be checked out"`
	GitUrl                             string `json:"git_url" arg:"-u,--git-url,env:GIT_URL,required" help:"The git url to checkout"`
	GitBranch                          string `json:"git_branch" arg:"-b,--git-branch,env:GIT_BRANCH" default:"main" help:"The git branch to checkout"`
	GitYamlPath                        string `json:"git_yaml_path" arg:"--git-yaml-path,env:GIT_YAML_ROOT" default:"" help:"The path where the yaml files are located"`
	NotificationsEnabled               bool   `json:"notifications_enabled" arg:"--notifications-enabled,env:NOTIFICATIONS_ENABLED" default:"false" help:"Sets if Notifications should be sent to the git provider, should be disabled if no token is provided in git url"`
	NotificationGroup                  string `json:"notification_group" arg:"--notification-group,env:NOTIFICATION_GROUP" default:"apps" help:"The notification group used by gitops-promotion"`
	CosmosDBAccount                    string `json:"cosmosdb_account" arg:"--cosmosdb-account,env:COSMOSDB_ACCOUNT,required" help:"The CosmosDB account to be used for cache"`
	CosmosDBSqlDb                      string `json:"cosmosdb_sql_db" arg:"--cosmosdb-sql-db,env:COSMOSDB_SQL_DB" default:"azcagit" help:"The CosmosDB SQL database to be used for cache"`
	CosmosDBAppCacheContainer          string `json:"cosmosdb_app_cache_container" arg:"--cosmosdb-app-cache-container,env:COSMOSDB_APP_CACHE_CONTAINER" default:"app-cache" help:"The CosmosDB container used for the app cache"`
	CosmosDBJobCacheContainer          string `json:"cosmosdb_job_cache_container" arg:"--cosmosdb-job-cache-container,env:COSMOSDB_JOB_CACHE_CONTAINER" default:"job-cache" help:"The CosmosDB container used for the job cache"`
	CosmosDBNotificationCacheContainer string `json:"cosmosdb_notification_cache_container" arg:"--cosmosdb-notification-cache-container,env:COSMOSDB_NOTIFICATION_CACHE_CONTAINER" default:"notification-cache" help:"The CosmosDB container used for the notification cache"`
	DebugEnabled                       bool   `json:"debug_enabled" arg:"--debug,env:DEBUG" default:"false" help:"Enabled debug logging"`
}

func (cfg *ReconcileConfig) Redacted() ReconcileConfig {
	if cfg == nil {
		return ReconcileConfig{}
	}

	redactedCfg := *cfg
	redactedCfg.GitUrl = redactUrl(redactedCfg.GitUrl)
	if redactedCfg.ContainerRegistryPassword != "" {
		redactedCfg.ContainerRegistryPassword = "redacted"
	}

	return redactedCfg
}

func redactUrl(u string) string {
	if u == "" {
		return ""
	}

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

type TriggerConfig struct {
	SubscriptionID      string `json:"subscription_id" arg:"-s,--subscription-id,env:AZURE_SUBSCRIPTION_ID,required" help:"Azure Subscription ID"`
	JobName             string `json:"job_name" arg:"--job-name,env:JOB_NAME,required" help:"The name of the container app job running azcagit"`
	ResourceGroupName   string `json:"resource_group_name" arg:"--resource-group-name,env:RESOURCE_GROUP_NAME,required" help:"The resource group name of where container app job running azcagit is located"`
	ServiceBusNamespace string `json:"service_bus_namespace" arg:"--service-bus-namespace,env:SERVICE_BUS_NAMESPACE,required" help:"The namespace of the service bus"`
	ServiceBusQueue     string `json:"service_bus_queue" arg:"--service-bus-queue,env:SERVICE_BUS_QUEUE,required" help:"The queue name of where to consume the messages from the service bus"`
}

type Config struct {
	ReconcileCfg *ReconcileConfig `arg:"subcommand:reconcile" help:"run reconciliation"`
	TriggerCfg   *TriggerConfig   `arg:"subcommand:trigger" help:"run trigger"`
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

	if parser.Subcommand() == nil {
		parser.WriteHelp(os.Stderr)
		return Config{}, fmt.Errorf("missing subcommand")
	}

	return cfg, nil
}
