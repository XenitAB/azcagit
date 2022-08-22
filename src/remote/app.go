package remote

import "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"

type RemoteApp struct {
	app     *armappcontainers.ContainerApp
	managed bool
}

func (r *RemoteApp) Managed() bool {
	return r.managed
}

func (r *RemoteApp) App() *armappcontainers.ContainerApp {
	return r.app
}

type RemoteApps map[string]RemoteApp
