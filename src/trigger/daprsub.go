package trigger

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"

	"github.com/xenitab/azcagit/src/config"
)

type DaprSubTrigger struct {
	service   common.Service
	triggerCh chan TriggeredBy
}

var _ Trigger = (*DaprSubTrigger)(nil)

var TriggeredByDaprSub TriggeredBy = "DaprSub"

func NewDaprSubTrigger(cfg config.Config) (*DaprSubTrigger, error) {
	service := daprd.NewService(fmt.Sprintf(":%d", cfg.DaprHttpPort))
	triggerCh := make(chan TriggeredBy)
	trigger := &DaprSubTrigger{
		service,
		triggerCh,
	}

	var subscription = &common.Subscription{
		PubsubName: cfg.DaprPubsubName,
		Topic:      cfg.DaprTopic,
		Route:      "/trigger",
	}

	err := service.AddTopicEventHandler(subscription, trigger.triggerHandler)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (t *DaprSubTrigger) WaitForTrigger() <-chan TriggeredBy {
	return t.triggerCh
}

func (t *DaprSubTrigger) Start() error {
	if err := t.service.Start(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (t *DaprSubTrigger) Stop() error {
	return t.service.GracefulStop()
}

func (t *DaprSubTrigger) triggerHandler(ctx context.Context, e *common.TopicEvent) (bool, error) {
	fmt.Printf("Data: %v\n", e.Data)
	triggerData := struct {
		Trigger *bool `json:"trigger"`
	}{}

	err := e.Struct(&triggerData)
	if err != nil {
		return false, err
	}

	if triggerData.Trigger == nil {
		return false, fmt.Errorf("trigger data not set")
	}

	if !*triggerData.Trigger {
		return false, fmt.Errorf("trigger data set to false")
	}

	t.triggerCh <- TriggeredByDaprSub

	return false, nil
}
