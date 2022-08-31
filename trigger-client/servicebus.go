package main

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/dapr/go-sdk/service/common"
)

type serviceBusClient struct {
	client *azservicebus.Client
	sender *azservicebus.Sender
}

func newServiceBusClient(cfg config) (*serviceBusClient, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := azservicebus.NewClient(cfg.FullyQualifiedNamespace, credential, nil)
	if err != nil {
		return nil, err
	}

	sender, err := client.NewSender(cfg.Topic, &azservicebus.NewSenderOptions{})
	if err != nil {
		return nil, err
	}

	return &serviceBusClient{
		client,
		sender,
	}, nil
}

func (c *serviceBusClient) Trigger(ctx context.Context) error {
	triggerData := struct {
		Trigger bool `json:"trigger"`
	}{
		Trigger: true,
	}
	event := &common.TopicEvent{
		PubsubName: "azcagit_trigger",
		ID:         "azcagit_trigger",
		Data:       triggerData,
	}

	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &azservicebus.Message{Body: b}
	err = c.sender.SendMessage(ctx, msg, &azservicebus.SendMessageOptions{})
	if err != nil {
		return err
	}

	return nil
}
