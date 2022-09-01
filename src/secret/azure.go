package secret

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/xenitab/azcagit/src/config"
)

type AzureKeyVaultSecret struct {
	client *azsecrets.Client
}

var _ Secret = (*AzureKeyVaultSecret)(nil)

func NewAzureKeyVaultSecret(cfg config.Config, cred azcore.TokenCredential) (*AzureKeyVaultSecret, error) {
	vaultUrl := fmt.Sprintf("https://%s.vault.azure.net", cfg.KeyVaultName)
	client, err := azsecrets.NewClient(vaultUrl, cred, nil)
	if err != nil {
		return nil, err
	}

	return &AzureKeyVaultSecret{
		client,
	}, nil
}

func (s *AzureKeyVaultSecret) ListItems(ctx context.Context) (*Items, error) {
	items := make(Items)
	pager := s.client.ListPropertiesOfSecrets(&azsecrets.ListSecretsOptions{})
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range nextResult.Secrets {
			changedAt := item.Properties.UpdatedOn
			if changedAt == nil {
				if item.Properties.CreatedOn == nil {
					return nil, fmt.Errorf("both UpdatedOn and CreatedOn are nil")
				}
				changedAt = item.Properties.CreatedOn
			}

			items[*item.Name] = Item{
				name:      *item.Name,
				changedAt: *changedAt,
			}
		}
	}

	return &items, nil
}

func (s *AzureKeyVaultSecret) Get(ctx context.Context, name string) (string, error) {
	res, err := s.client.GetSecret(ctx, name, &azsecrets.GetSecretOptions{})
	if err != nil {
		return "", err
	}

	if res.Value == nil {
		return "", fmt.Errorf("value for secret %q is nil", name)
	}

	return *res.Value, nil
}
