package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type CosmosDBClient struct {
	containerClient *azcosmos.ContainerClient
}

func NewCosmosDBClient(account string, db string, container string, cred azcore.TokenCredential) (*CosmosDBClient, error) {
	endpoint := fmt.Sprintf("https://%s.documents.azure.com:443/", account)
	client, err := azcosmos.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, err
	}

	containerClient, err := client.NewContainer(db, container)
	if err != nil {
		return nil, err
	}

	return &CosmosDBClient{
		containerClient,
	}, nil
}

type CosmosDBContainerClient[T any] struct {
	client       *azcosmos.ContainerClient
	partitionKey string
	ttl          *int
}

type cosmosDBEntry[T any] struct {
	Id          string `json:"id"`
	ParitionKey string `json:"partition_key"`
	TTL         *int   `json:"ttl"`
	Value       T      `json:"value"`
}

func NewCosmosDBContainerClient[T any](cosmosDBClient *CosmosDBClient, partitionKey string, ttl *int) (*CosmosDBContainerClient[T], error) {
	return &CosmosDBContainerClient[T]{
		client:       cosmosDBClient.containerClient,
		partitionKey: partitionKey,
		ttl:          ttl,
	}, nil
}

func (client *CosmosDBContainerClient[T]) getId(key string) string {
	return fmt.Sprintf("%s-%s", client.partitionKey, key)
}

func (client *CosmosDBContainerClient[T]) Get(ctx context.Context, key string) (*T, error) {
	item, err := client.client.ReadItem(ctx, azcosmos.NewPartitionKeyString(client.partitionKey), client.getId(key), &azcosmos.ItemOptions{})
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

	entry := &cosmosDBEntry[T]{}
	err = json.Unmarshal(item.Value, entry)
	if err != nil {
		return new(T), err
	}

	return &entry.Value, nil
}

func (client *CosmosDBContainerClient[T]) Set(ctx context.Context, key string, value T) error {
	b, err := json.Marshal(cosmosDBEntry[T]{
		Id:          client.getId(key),
		ParitionKey: client.partitionKey,
		TTL:         client.ttl,
		Value:       value,
	})
	if err != nil {
		return err
	}

	_, err = client.client.UpsertItem(ctx, azcosmos.NewPartitionKeyString(client.partitionKey), b, &azcosmos.ItemOptions{})
	if err != nil {
		return err
	}

	return nil
}
