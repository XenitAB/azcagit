package source

import (
	"context"
	"sort"
)

type Sources struct {
	Apps *SourceApps
	Jobs *SourceJobs
}

func (srcs *Sources) GetUniqueRemoteSecretNames() []string {
	secretsMap := make(map[string]struct{})

	if srcs == nil {
		return nil
	}

	if srcs.Apps != nil {
		for _, remoteSecretName := range srcs.Apps.GetUniqueRemoteSecretNames() {
			secretsMap[remoteSecretName] = struct{}{}
		}
	}

	if srcs.Jobs != nil {
		for _, remoteSecretName := range srcs.Jobs.GetUniqueRemoteSecretNames() {
			secretsMap[remoteSecretName] = struct{}{}
		}
	}

	secrets := []string{}
	for secret := range secretsMap {
		secrets = append(secrets, secret)
	}
	sort.Strings(secrets)

	return secrets
}

type Source interface {
	Get(ctx context.Context) (*Sources, string, error)
}
