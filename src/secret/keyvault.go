package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/xenitab/azcagit/src/config"
)

type KeyVaultSecret struct {
	client *azsecrets.Client
}

var _ Secret = (*KeyVaultSecret)(nil)

func NewKeyVaultSecret(cfg config.ReconcileConfig, cred azcore.TokenCredential) (*KeyVaultSecret, error) {
	vaultUrl := fmt.Sprintf("https://%s.vault.azure.net", cfg.KeyVaultName)
	client, err := azsecrets.NewClient(vaultUrl, cred, nil)
	if err != nil {
		return nil, err
	}

	return &KeyVaultSecret{
		client,
	}, nil
}

func (s *KeyVaultSecret) ListItems(ctx context.Context) (*Items, error) {
	items := make(Items)
	pager := s.client.NewListSecretsPager(&azsecrets.ListSecretsOptions{})
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range nextResult.Value {
			changedAt := item.Attributes.Updated
			if changedAt == nil {
				if item.Attributes.Created == nil {
					return nil, fmt.Errorf("both UpdatedOn and CreatedOn are nil")
				}
				changedAt = item.Attributes.Created
			}

			items[item.ID.Name()] = Item{
				name:      item.ID.Name(),
				changedAt: *changedAt,
			}
		}
	}

	return &items, nil
}

func (s *KeyVaultSecret) Get(ctx context.Context, name string) (string, time.Time, error) {
	res, err := s.client.GetSecret(ctx, name, "", &azsecrets.GetSecretOptions{})
	if err != nil {
		return "", time.Time{}, err
	}

	if res.Value == nil {
		return "", time.Time{}, fmt.Errorf("value for secret %q is nil", name)
	}

	changedAt := res.Attributes.Updated
	if changedAt == nil {
		if res.Attributes.Created == nil {
			return "", time.Time{}, fmt.Errorf("both UpdatedOn and CreatedOn are nil")
		}
		changedAt = res.Attributes.Created
	}

	return *res.Value, *changedAt, nil
}
