// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package assists with chaining together providers by providing default implementations for
LogProvider interface methods that don't panic if we're the last in the chain.
*/
package chaining

import (
	"github.com/calm/contextlogger/v2/providers"

	"context"
)

type provider struct {
	nextProvider providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider) providers.LogProvider {
	return provider{nextProvider}
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	if p.nextProvider != nil {
		p.nextProvider.Error(ctx, report, args...)
	}
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	if p.nextProvider != nil {
		p.nextProvider.Warn(ctx, report, args...)
	}
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	if p.nextProvider != nil {
		p.nextProvider.Info(ctx, report, args...)
	}
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	if p.nextProvider != nil {
		p.nextProvider.Debug(ctx, report, args...)
	}
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	if p.nextProvider != nil {
		p.nextProvider.Record(ctx, metrics)
	}
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	if p.nextProvider != nil {
		p.nextProvider.RecordEvent(ctx, eventName, metrics)
	}
}

func (p provider) Wait() {
	if p.nextProvider != nil {
		p.nextProvider.Wait()
	}
}
