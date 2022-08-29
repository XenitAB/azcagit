package main

import (
	"context"
	"fmt"
	"os"
)

type config struct {
	FullyQualifiedNamespace string
	QueueOrTopic            string
}

func main() {
	cfg := config{
		FullyQualifiedNamespace: "sbcontainerapps.servicebus.windows.net",
		QueueOrTopic:            "azcagit_trigger",
	}
	err := run(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "application returned an error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	sbClient, err := newServiceBusClient(cfg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return sbClient.Trigger(ctx)
}
