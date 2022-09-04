// Orignial copyright from Flux (commit: 2c69c84): https://github.com/fluxcd/notification-controller/blob/main/internal/notifier/github.go
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

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type GitHubNotification struct {
	Owner  string
	Repo   string
	Client *github.Client
}

var _ Notification = (*GitHubNotification)(nil)

func NewGitHubNotification(address string, token string) (*GitHubNotification, error) {
	if len(token) == 0 {
		return nil, errors.New("github token cannot be empty")
	}

	_, id, err := parseGitAddress(address)
	if err != nil {
		return nil, err
	}

	comp := strings.Split(id, "/")
	if len(comp) != 2 {
		return nil, fmt.Errorf("invalid repository id %q", id)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &GitHubNotification{
		Owner:  comp[0],
		Repo:   comp[1],
		Client: client,
	}, nil
}

func (g *GitHubNotification) Send(ctx context.Context, event NotificationEvent) error {
	state, err := toGitHubState(event.State)
	if err != nil {
		return err
	}
	status := &github.RepoStatus{
		State:       &state,
		Context:     &event.Name,
		Description: toGitHubDescription(event.Description),
	}

	opts := &github.ListOptions{PerPage: 50}
	statuses, _, err := g.Client.Repositories.ListStatuses(ctx, g.Owner, g.Repo, event.Revision, opts)
	if err != nil {
		return fmt.Errorf("could not list commit statuses: %v", err)
	}
	if duplicateGithubStatus(statuses, status) {
		return nil
	}

	_, _, err = g.Client.Repositories.CreateStatus(ctx, g.Owner, g.Repo, event.Revision, status)
	if err != nil {
		return fmt.Errorf("could not create commit status: %v", err)
	}

	return nil
}

func toGitHubDescription(description string) *string {
	if len(description) <= 140 {
		return &description
	}

	strippedDescription := description[0:140]
	return &strippedDescription
}

func toGitHubState(state NotificationState) (string, error) {
	switch state {
	case NotificationStateSuccess:
		return "success", nil
	case NotificationStateFailure:
		return "failure", nil
	default:
		return "", errors.New("can't convert to github state")
	}
}

// duplicateStatus return true if the latest status
// with a matching context has the same state and description
func duplicateGithubStatus(statuses []*github.RepoStatus, status *github.RepoStatus) bool {
	for _, s := range statuses {
		if *s.Context == *status.Context {
			if *s.State == *status.State && *s.Description == *status.Description {
				return true
			}

			return false
		}
	}

	return false
}
