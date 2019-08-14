package statsd

import (
	"context"
	"errors"
	"fmt"

	"github.com/calm/contextlogger/providers"
	"github.com/calm/contextlogger/providers/chaining"
	etsystatsd "github.com/etsy/statsd/examples/go"
)

type provider struct {
	sharedStatsdClient *etsystatsd.StatsdClient // This has to be passed in; we can't import package config
	providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider, client *etsystatsd.StatsdClient) (providers.LogProvider, error) {
	if client == nil {
		return nil, errors.New("statsd client is required")
	}
	return provider{client, chaining.LogProvider(nextProvider)}, nil
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	statsToSend := make(map[string]string)
	for k, v := range metrics {
		statsToSend[k] = fmt.Sprintf("%d|g", v)
	}

	p.sharedStatsdClient.Send(statsToSend, 1)
	p.LogProvider.Record(ctx, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.sharedStatsdClient.Increment(eventName)
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}
