package ginkgoprovider

import (
	"context"
	"fmt"
	"io"

	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
	"github.com/onsi/ginkgo/v2"
)

func LogProvider(nextProvider providers.LogProvider) providers.LogProvider {
	return &provider{
		Writer:      ginkgo.GinkgoWriter,
		LogProvider: chaining.LogProvider(nextProvider),
	}
}

type provider struct {
	io.Writer
	providers.LogProvider
}

func (p *provider) Error(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
	p.LogProvider.Error(ctx, report, args...)
}

func (p *provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
	p.LogProvider.Warn(ctx, report, args...)
}

func (p *provider) Info(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
	p.LogProvider.Info(ctx, report, args...)
}

func (p *provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
	p.LogProvider.Debug(ctx, report, args...)
}

func (p *provider) Record(ctx context.Context, metrics map[string]interface{}) {
	fmt.Fprintln(p, metrics)
	p.LogProvider.Record(ctx, metrics)
}

func (p *provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	fmt.Fprintln(p, eventName, metrics)
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}
