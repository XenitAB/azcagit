package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type CosmosDBClient[T any] struct {
	containerClient *azcosmos.ContainerClient
}

func NewCosmosDBClient[T any](account string, db string, container string, cred azcore.TokenCredential) (*CosmosDBClient[T], error) {
	endpoint := fmt.Sprintf("https://%s.documents.azure.com:443/", account)
	client, err := azcosmos.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, err
	}

	containerClient, err := client.NewContainer(db, container)
	if err != nil {
		return nil, err
	}

	return &CosmosDBClient[T]{
		containerClient,
	}, nil
}

func (cache *CosmosDBClient[T]) Get(ctx context.Context, key string) (*T, error) {
	item, err := cache.containerClient.ReadItem(ctx, azcosmos.NewPartitionKeyString(key), key, &azcosmos.ItemOptions{})
	isNotFound := err != nil && strings.Contains(err.Error(), "404 Not Found")
	if err != nil && !isNotFound {
		return new(T), err
	}

	if isNotFound {
		return nil, nil
	}

	if len(item.Value) == 0 {
		return nil, nil
	}

	value := new(T)
	err = json.Unmarshal(item.Value, value)
	if err != nil {
		return new(T), err
	}

	return value, nil
}

func (cache *CosmosDBClient[T]) Set(ctx context.Context, key string, value T) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = cache.containerClient.UpsertItem(ctx, azcosmos.NewPartitionKeyString(key), b, &azcosmos.ItemOptions{})
	if err != nil {
		return err
	}

	return nil
}
