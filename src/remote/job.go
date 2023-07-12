package remote

import (
	"sort"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type RemoteJob struct {
	Job     *armappcontainers.Job
	Managed bool
}

type RemoteJobs map[string]RemoteJob

func (jobs *RemoteJobs) GetSortedNames() []string {
	names := []string{}
	for name := range *jobs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (jobs *RemoteJobs) Get(name string) (RemoteJob, bool) {
	job, ok := (*jobs)[name]
	return job, ok
}
