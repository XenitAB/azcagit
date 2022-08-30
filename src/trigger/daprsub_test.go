package trigger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenitab/azcagit/src/config"
	"golang.org/x/sync/errgroup"

	"github.com/dapr/go-sdk/service/common"
)

func TestDaprSubTrigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	cfg := config.Config{
		DaprAppPort:    8080,
		DaprPubsubName: "sb",
		DaprTopic:      "azcagit_trigger",
	}
	trigger, err := NewDaprSubTrigger(cfg)
	require.NoError(t, err)

	g.Go(func() error {
		return trigger.Start()
	})

	for start := time.Now(); time.Since(start) < 5*time.Second; {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", cfg.DaprAppPort))
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	gClient, ctxClient := errgroup.WithContext(ctx)
	httpClient := &http.Client{}
	gClient.Go(func() error {
		triggerData := struct {
			Trigger bool `json:"trigger"`
		}{
			Trigger: true,
		}
		event := &common.TopicEvent{
			PubsubName: "test_name",
			ID:         "test_id",
			Data:       triggerData,
		}
		b, err := json.Marshal(event)
		require.NoError(t, err)
		t.Logf("Event: %v", string(b))
		req, err := http.NewRequestWithContext(ctxClient, http.MethodPost, "http://localhost:8080/trigger", bytes.NewBuffer(b))
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")

		res, err := httpClient.Do(req)
		require.NoError(t, err)
		b, err = io.ReadAll(res.Body)
		require.NoError(t, err)
		defer res.Body.Close()

		resData := struct {
			Status string `json:"status"`
		}{}

		err = json.Unmarshal(b, &resData)
		require.NoError(t, err)

		require.Equal(t, "SUCCESS", resData.Status)

		return err
	})

	timeout := time.NewTimer(50 * time.Millisecond)
	select {
	case <-timeout.C:
		t.Logf("Waiting for trigger timed out")
		t.Fail()
	case triggeredBy := <-trigger.WaitForTrigger():
		t.Logf("Received trigger: %s", triggeredBy)
		require.Equal(t, TriggeredBy("DaprSub"), triggeredBy)
	}

	require.NoError(t, gClient.Wait())

	g.Go(func() error {
		return trigger.Stop()
	})

	err = g.Wait()
	require.NoError(t, err)
}
