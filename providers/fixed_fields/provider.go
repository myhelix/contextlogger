// © 2026 Helix OpCo LLC. All rights reserved.

package fixed_fields

import (
	"context"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
)

type provider struct {
	providers.LogProvider
	fields log.Fields
}

func LogProvider(nextProvider providers.LogProvider, fields log.Fields) providers.LogProvider {
	return provider{
		LogProvider: chaining.LogProvider(nextProvider),
		fields:      fields,
	}
}

func (p provider) context(ctx context.Context) context.Context {
	return log.ContextWithFields(ctx, p.fields)
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Error(p.context(ctx), report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Warn(p.context(ctx), report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Info(p.context(ctx), report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Debug(p.context(ctx), report, args...)
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.LogProvider.Record(ctx, p.metrics(metrics))
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.LogProvider.RecordEvent(ctx, eventName, p.metrics(metrics))
}

func (p provider) metrics(metrics map[string]interface{}) map[string]interface{} {
	withFields := make(map[string]interface{}, len(metrics)+len(p.fields))
	for key, value := range metrics {
		withFields[key] = value
	}
	for key, value := range p.fields {
		withFields[key] = value
	}
	return withFields
}
