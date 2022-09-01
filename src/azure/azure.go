package azure

import "github.com/Azure/azure-sdk-for-go/sdk/azidentity"

func NewAzureCredential() (*azidentity.DefaultAzureCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	return cred, nil
}
