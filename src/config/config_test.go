package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	args := []string{
		"/foo/bar/bin",
		"--resource-group-name",
		"foo",
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
	}, cfg)
}

func TestRedactedConfig(t *testing.T) {
	cfgWithUserAndPass := Config{
		ContainerRegistryPassword: "secret",
		GitUrl:                    "https://foo:bar@foobar.io/abc.git",
	}
	require.Equal(t, "redacted", cfgWithUserAndPass.Redacted().ContainerRegistryPassword)
	require.Equal(t, "https://foo:redacted@foobar.io/abc.git", cfgWithUserAndPass.Redacted().GitUrl)

	cfg := Config{
		ContainerRegistryPassword: "",
		GitUrl:                    "https://foobar.io/abc.git",
	}
	require.Equal(t, "", cfg.Redacted().ContainerRegistryPassword)
	require.Equal(t, "https://foobar.io/abc.git", cfg.Redacted().GitUrl)
}

func TestRedactUrl(t *testing.T) {
	require.Equal(t, "https://foobar.io/abc.git", redactUrl("https://foobar.io/abc.git"))
	require.Equal(t, "https://foo:redacted@foobar.io/abc.git", redactUrl("https://foo:bar@foobar.io/abc.git"))
	require.Equal(t, "https://redacted@foobar.io/abc.git", redactUrl("https://foo@foobar.io/abc.git"))
}
