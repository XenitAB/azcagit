package config

type Config struct {
	ResourceGroupName    string
	SubscriptionID       string
	ManagedEnvironmentID string
	Location             string
	ReconcileInterval    string
	CheckoutPath         string
	GitUrl               string
	GitBranch            string
}
