package remote

import (
	"sort"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

type RemoteApp struct {
	App     *armappcontainers.ContainerApp
	Managed bool
}

type RemoteApps map[string]RemoteApp

func (apps *RemoteApps) GetSortedNames() []string {
	names := []string{}
	for name := range *apps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (apps *RemoteApps) Get(name string) (RemoteApp, bool) {
	app, ok := (*apps)[name]
	return app, ok
}
