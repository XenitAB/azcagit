// Orignial copyright from Flux (commit: 2c69c84): https://github.com/fluxcd/notification-controller/blob/main/internal/notifier/azure_devops.go
// /*
// Copyright 2020 The Flux authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v6"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/git"
)

type AzureDevOpsNotification struct {
	Project string
	Repo    string
	Client  git.Client
}

var _ Notification = (*AzureDevOpsNotification)(nil)

func NewAzureDevOpsNotification(address string, token string) (*AzureDevOpsNotification, error) {
	if len(token) == 0 {
		return nil, errors.New("azure devops token cannot be empty")
	}

	host, id, err := parseGitAddress(address)
	if err != nil {
		return nil, err
	}

	comp := strings.Split(id, "/")
	if len(comp) != 4 {
		return nil, fmt.Errorf("invalid repository id %q", id)
	}
	org := comp[0]
	proj := comp[1]
	repo := comp[3]

	orgURL := fmt.Sprintf("%v/%v", host, org)
	connection := azuredevops.NewPatConnection(orgURL, token)
	client := connection.GetClientByUrl(orgURL)
	gitClient := &git.ClientImpl{
		Client: *client,
	}
	return &AzureDevOpsNotification{
		Project: proj,
		Repo:    repo,
		Client:  gitClient,
	}, nil
}

func (a *AzureDevOpsNotification) Send(ctx context.Context, event NotificationEvent) error {
	rev, err := parseRevision(event.Revision)
	if err != nil {
		return err
	}
	state, err := toAzureDevOpsState(event.State)
	if err != nil {
		return err
	}

	// Check if the exact status is already set
	createArgs := git.CreateCommitStatusArgs{
		Project:      &a.Project,
		RepositoryId: &a.Repo,
		CommitId:     &rev,
		GitCommitStatusToCreate: &git.GitStatus{
			Description: &event.Description,
			State:       &state,
			Context: &git.GitStatusContext{
				Genre: toPtr("azcagit"),
				Name:  &event.Name,
			},
		},
	}
	getArgs := git.GetStatusesArgs{
		Project:      &a.Project,
		RepositoryId: &a.Repo,
		CommitId:     &rev,
	}
	statuses, err := a.Client.GetStatuses(ctx, getArgs)
	if err != nil {
		return fmt.Errorf("could not list commit statuses: %v", err)
	}
	if duplicateAzureDevOpsStatus(statuses, createArgs.GitCommitStatusToCreate) {
		return nil
	}

	// Create a new status
	_, err = a.Client.CreateCommitStatus(ctx, createArgs)
	if err != nil {
		return fmt.Errorf("could not create commit status: %v", err)
	}
	return nil
}

func toAzureDevOpsState(state NotificationState) (git.GitStatusState, error) {
	switch state {
	case NotificationStateSuccess:
		return git.GitStatusStateValues.Succeeded, nil
	case NotificationStateFailure:
		return git.GitStatusStateValues.Error, nil
	default:
		return "", errors.New("can't convert to azure devops state")
	}
}

// duplicateStatus return true if the latest status
// with a matching context has the same state and description
func duplicateAzureDevOpsStatus(statuses *[]git.GitStatus, status *git.GitStatus) bool {
	for _, s := range *statuses {
		if *s.Context.Name == *status.Context.Name && *s.Context.Genre == *status.Context.Genre {
			if *s.State == *status.State && *s.Description == *status.Description {
				return true
			}

			return false
		}
	}

	return false
}
