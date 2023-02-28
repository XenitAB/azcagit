package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
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
		"RECONCILE_INTERVAL",
		"CHECKOUT_PATH",
		"GIT_URL",
		"GIT_BRANCH",
		"GIT_YAML_ROOT",
		"DAPR_APP_PORT",
		"DAPR_PUBSUB_NAME",
		"DAPR_TOPIC_NAME",
		"NOTIFICATIONS_ENABLED",
		"NOTIFICATION_GROUP",
		"DEBUG",
	}

	for _, envVar := range envVarsToClear {
		restore := testTempUnsetEnv(t, envVar)
		defer restore()
	}

	args := []string{
		"/foo/bar/bin",
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
		"--dapr-topic-name",
		"ze-topic",
	}
	cfg, err := NewConfig(args[1:])
	require.NoError(t, err)
	require.Equal(t, Config{
		ResourceGroupName:    "foo",
		Environment:          "foobar",
		SubscriptionID:       "bar",
		ManagedEnvironmentID: "baz",
		KeyVaultName:         "ze-keyvault",
		OwnContainerAppName:  "azcagit",
		OwnResourceGroupName: "platform",
		Location:             "westeurope",
		ReconcileInterval:    "5m",
		CheckoutPath:         "/tmp",
		GitUrl:               "https://github.com/foo/bar.git",
		GitBranch:            "main",
		DaprAppPort:          8080,
		DaprPubsubName:       "azcagit-trigger",
		DaprTopic:            "ze-topic",
		NotificationGroup:    "apps",
	}, cfg)
}

func TestRedactedConfig(t *testing.T) {
	cfgWithUserAndPass := Config{
		ContainerRegistryPassword: "secret",                            // secretlint-disable
		GitUrl:                    "https://foo:bar@foobar.io/abc.git", // secretlint-disable
	}
	require.Equal(t, "redacted", cfgWithUserAndPass.Redacted().ContainerRegistryPassword)
	require.Equal(t, "https://foo:redacted@foobar.io/abc.git", cfgWithUserAndPass.Redacted().GitUrl) // secretlint-disable

	cfg := Config{
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
