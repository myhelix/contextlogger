package mockprovider

import (
	"context"

	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
	"github.com/stretchr/testify/mock"
)

func LogProvider(nextProvider providers.LogProvider) *provider {
	return &provider{LogProvider: chaining.LogProvider(nextProvider)}
}

type provider struct {
	mock.Mock
	providers.LogProvider
}

func (p *provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Error(ctx, report, args...)
}

func (p *provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Warn(ctx, report, args...)
}

func (p *provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Info(ctx, report, args...)
}

func (p *provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Debug(ctx, report, args...)
}

func (p *provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.Called(ctx, metrics)
	p.LogProvider.Record(ctx, metrics)
}

func (p *provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.Called(ctx, eventName, metrics)
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}

func (p *provider) Wait() {
	p.Called()
	p.LogProvider.Wait()
}
