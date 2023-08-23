package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

func main() {
	cfg, err := newConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to load configuration: %v\n", err)
		os.Exit(1)
	}

	err = run(cfg)
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

type config struct {
	ServiceBusNamespace string `arg:"-n,--namespace,required" help:"Service Bus namespace"`
	ServiceBusQueue     string `arg:"-q,--queue,required" help:"Service Bus queue"`
}

func newConfig(args []string) (config, error) {
	cfg := config{}

	parser, err := arg.NewParser(arg.Config{
		Program:   "azcagit-trigger-client",
		IgnoreEnv: false,
	}, &cfg)
	if err != nil {
		return config{}, err
	}

	err = parser.Parse(args)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}
