package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
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

	namespaceFqdn := fmt.Sprintf("%s.servicebus.windows.net", cfg.ServiceBusNamespace)
	client, err := azservicebus.NewClient(namespaceFqdn, credential, nil)
	if err != nil {
		return nil, err
	}

	sender, err := client.NewSender(cfg.ServiceBusQueue, &azservicebus.NewSenderOptions{})
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

	b, err := json.Marshal(triggerData)
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
