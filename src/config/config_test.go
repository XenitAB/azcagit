package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewReconcileConfig(t *testing.T) {
	envVarsToClear := []string{
		"RESOURCE_GROUP_NAME",
		"ENVIRONMENT",
		"AZURE_SUBSCRIPTION_ID",
		"MANAGED_ENVIRONMENT_ID",
		"KEY_VAULT_NAME",
		"OWN_CONTAINER_APP_NAME",
		"OWN_RESOURCE_GROUP_NAME",
		"CONTAINER_REGISTRY_SERVER",
		"CONTAINER_REGISTRY_USERNAME",
		"CONTAINER_REGISTRY_PASSWORD",
		"CHECKOUT_PATH",
		"GIT_URL",
		"GIT_BRANCH",
		"GIT_YAML_ROOT",
		"NOTIFICATIONS_ENABLED",
		"NOTIFICATION_GROUP",
		"DEBUG",
		"COSMOSDB_ACCOUNT",
		"COSMOSDB_SQL_DB",
		"COSMOSDB_APP_CACHE_CONTAINER",
		"COSMOSDB_JOB_CACHE_CONTAINER",
	}

	for _, envVar := range envVarsToClear {
		restore := testTempUnsetEnv(t, envVar)
		defer restore()
	}

	args := []string{
		"/foo/bar/bin",
		"reconcile",
		"--resource-group-name",
		"foo",
		"--environment",
		"foobar",
		"--subscription-id",
		"bar",
		"--managed-environment-id",
		"baz",
		"--key-vault-name",
		"ze-keyvault",
		"--own-resource-group-name",
		"platform",
		"--location",
		"westeurope",
		"--git-url",
		"https://github.com/foo/bar.git",
		"--cosmosdb-account",
		"ze-cosmosdb-account",
	}
	cfg, err := NewConfig(args[1:])
	require.NoError(t, err)
	require.Equal(t, ReconcileConfig{
		ResourceGroupName:                  "foo",
		Environment:                        "foobar",
		SubscriptionID:                     "bar",
		ManagedEnvironmentID:               "baz",
		KeyVaultName:                       "ze-keyvault",
		OwnContainerJobName:                "azcagit-reconcile",
		OwnResourceGroupName:               "platform",
		Location:                           "westeurope",
		CheckoutPath:                       "/tmp",
		GitUrl:                             "https://github.com/foo/bar.git",
		GitBranch:                          "main",
		NotificationGroup:                  "apps",
		CosmosDBAccount:                    "ze-cosmosdb-account",
		CosmosDBSqlDb:                      "azcagit",
		CosmosDBAppCacheContainer:          "app-cache",
		CosmosDBJobCacheContainer:          "job-cache",
		CosmosDBNotificationCacheContainer: "notification-cache",
	}, *cfg.ReconcileCfg)
}

func TestRedactedReconcileConfig(t *testing.T) {
	cfgWithUserAndPass := ReconcileConfig{
		ContainerRegistryPassword: "secret",                            // secretlint-disable
		GitUrl:                    "https://foo:bar@foobar.io/abc.git", // secretlint-disable
	}
	require.Equal(t, "redacted", cfgWithUserAndPass.Redacted().ContainerRegistryPassword)
	require.Equal(t, "https://foo:redacted@foobar.io/abc.git", cfgWithUserAndPass.Redacted().GitUrl) // secretlint-disable

	cfg := ReconcileConfig{
		ContainerRegistryPassword: "",
		GitUrl:                    "https://foobar.io/abc.git", // secretlint-disable
	}
	require.Equal(t, "", cfg.Redacted().ContainerRegistryPassword)
	require.Equal(t, "https://foobar.io/abc.git", cfg.Redacted().GitUrl) // secretlint-disable
}

func TestRedactUrl(t *testing.T) {
	require.Equal(t, "https://foobar.io/abc.git", redactUrl("https://foobar.io/abc.git"))
	require.Equal(t, "https://foo:redacted@foobar.io/abc.git", redactUrl("https://foo:bar@foobar.io/abc.git")) // secretlint-disable
	require.Equal(t, "https://redacted@foobar.io/abc.git", redactUrl("https://foo@foobar.io/abc.git"))
}

func testTempUnsetEnv(t *testing.T, key string) func() {
	t.Helper()

	oldEnv := os.Getenv(key)
	os.Unsetenv(key)
	return func() { os.Setenv(key, oldEnv) }
}
