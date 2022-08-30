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
		"--location",
		"westeurope",
		"--checkout-path",
		"/tmp/foo",
		"--git-url",
		"https://github.com/foo/bar.git",
	}
	cfg, err := NewConfig(args[1:])
	require.NoError(t, err)
	require.Equal(t, Config{
		ResourceGroupName:    "foo",
		SubscriptionID:       "bar",
		ManagedEnvironmentID: "baz",
		Location:             "westeurope",
		ReconcileInterval:    "5m",
		CheckoutPath:         "/tmp/foo",
		GitUrl:               "https://github.com/foo/bar.git",
		GitBranch:            "main",
		DaprAppPort:          3501,
		DaprPubsubName:       "sb",
		DaprTopic:            "azcagit_trigger",
	}, cfg)
}
