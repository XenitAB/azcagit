package notification

import (
	"context"
	"fmt"
	"net/url"
)

type Notification interface {
	Send(ctx context.Context, event NotificationEvent) error
}

type NotificationEvent struct {
	Revision    string
	State       NotificationState
	Name        string
	Description string
}

func (e *NotificationEvent) Equal(other NotificationEvent) bool {
	if e.Revision != other.Revision {
		return false
	}

	if e.State != other.State {
		return false
	}

	if e.Name != other.Name {
		return false
	}

	if e.Description != other.Description {
		return false
	}

	return true
}

type NotificationState int

const (
	NotificationStateSuccess NotificationState = iota
	NotificationStateFailure
)

type NotificationProvider int

const (
	NotificationProviderAzureDevOps NotificationProvider = iota
	NotificationProviderGitHub
	NotificationProviderUnknown
)

func NewNotificationClient(gitUrl string) (Notification, error) {
	parsedGitUrl, err := url.Parse(gitUrl)
	if err != nil {
		return nil, err
	}

	address, token, err := parseGitAddressAndToken(parsedGitUrl)
	if err != nil {
		return nil, err
	}

	switch parseNotificationProvider(parsedGitUrl) {
	case NotificationProviderAzureDevOps:
		return NewAzureDevOpsNotification(address, token)
	case NotificationProviderGitHub:
		return NewGitHubNotification(address, token)
	}

	return nil, fmt.Errorf("can't find notification provider for hostname: %s", parsedGitUrl.Hostname())
}

func parseGitAddressAndToken(gitUrl *url.URL) (string, string, error) {
	token, ok := gitUrl.User.Password()
	if !ok {
		if gitUrl.User.Username() == "" {
			return "", "", fmt.Errorf("unable to parse token from gitUrl")
		}
		token = gitUrl.User.Username()
	}

	newGitUrl := *gitUrl
	newGitUrl.User = &url.Userinfo{}
	address := newGitUrl.String()

	if address == "" {
		return "", "", fmt.Errorf("unable to parse address from gitUrl")
	}

	return address, token, nil
}

func parseNotificationProvider(gitUrl *url.URL) NotificationProvider {
	switch gitUrl.Hostname() {
	case "github.com":
		return NotificationProviderGitHub
	case "dev.azure.com":
		return NotificationProviderAzureDevOps
	}

	return NotificationProviderUnknown
}
