// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package provides newrelic metric/request reporting via ContextLogger
*/
package newrelic

import (
	"github.com/newrelic/go-agent"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"

	"context"
	"errors"
)

type provider struct {
	newRelicApp  newrelic.Application // This has to be passed in; we can't import package config
	nextProvider providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider, newRelicApp newrelic.Application) (providers.LogProvider, error) {
	if nextProvider == nil {
		return nil, errors.New("Newrelic log provider requires a base provider")
	}
	if newRelicApp == nil {
		return nil, errors.New("newRelicApp is required")
	}
	return provider{newRelicApp, nextProvider}, nil
}

type contextNewRelicTxnKey struct{}

func WithTransaction(ctx context.Context, txn newrelic.Transaction) log.ContextLogger {
	return log.FromContext(context.WithValue(ctx, contextNewRelicTxnKey{}, txn))
}

func TxnFrom(ctx context.Context) newrelic.Transaction {
	if txn, ok := ctx.Value(contextNewRelicTxnKey{}).(newrelic.Transaction); ok {
		return txn
	}
	return nil
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	if txn := TxnFrom(ctx); txn != nil {
		for k, v := range metrics {
			txn.AddAttribute(k, v)
		}
	} else {
		log.FromContext(ctx).ErrorReport(errors.New("Attempted to record metric in context without NewRelic transaction"))
	}
	p.nextProvider.Record(ctx, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.newRelicApp.RecordCustomEvent(eventName, map[string]interface{}(metrics))
	p.nextProvider.RecordEvent(ctx, eventName, metrics)
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Error(ctx, report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Warn(ctx, report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Info(ctx, report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Debug(ctx, report, args...)
}

func (p provider) Wait() {
	p.nextProvider.Wait()
}
