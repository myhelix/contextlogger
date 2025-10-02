package mockprovider

import (
	"context"

	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
	"github.com/stretchr/testify/mock"
)

func LogProvider(nextProvider providers.LogProvider) *MockProvider {
	return &MockProvider{LogProvider: chaining.LogProvider(nextProvider)}
}

type MockProvider struct {
	mock.Mock
	providers.LogProvider
}

func (p *MockProvider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Error(ctx, report, args...)
}

func (p *MockProvider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Warn(ctx, report, args...)
}

func (p *MockProvider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Info(ctx, report, args...)
}

func (p *MockProvider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.Called(ctx, report, args)
	p.LogProvider.Debug(ctx, report, args...)
}

func (p *MockProvider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.Called(ctx, metrics)
	p.LogProvider.Record(ctx, metrics)
}

func (p *MockProvider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.Called(ctx, eventName, metrics)
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}

func (p *MockProvider) Wait() {
	p.Called()
	p.LogProvider.Wait()
}
