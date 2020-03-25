package statsd

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/calm/contextlogger/v2/providers"
	"github.com/calm/contextlogger/v2/providers/chaining"
)

type provider struct {
	sharedStatsdClient *statsd.Client // This has to be passed in; we can't import package config
	providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider, client *statsd.Client) (providers.LogProvider, error) {
	if client == nil {
		return nil, errors.New("statsd client is required")
	}
	return provider{client, chaining.LogProvider(nextProvider)}, nil
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	for key, value := range metrics {
		p.sharedStatsdClient.SimpleEvent(key, fmt.Sprintf("%d|g", value))
	}
	p.LogProvider.Record(ctx, metrics)
}

// RecordEvent the same as statsd.Increment()
func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.sharedStatsdClient.Incr(eventName, make([]string, 0), float64(1))
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}
