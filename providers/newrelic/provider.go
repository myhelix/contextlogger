// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package provides newrelic metric/request reporting via ContextLogger
*/
package newrelic

import (
	newrelic "github.com/newrelic/go-agent"

	"github.com/calm/contextlogger/v2/log"
	"github.com/calm/contextlogger/v2/providers"
	"github.com/calm/contextlogger/v2/providers/chaining"

	"context"
	"errors"
)

type provider struct {
	newRelicApp newrelic.Application // This has to be passed in; we can't import package config
	providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider, newRelicApp newrelic.Application) (providers.LogProvider, error) {
	if newRelicApp == nil {
		return nil, errors.New("newRelicApp is required")
	}
	return provider{newRelicApp, chaining.LogProvider(nextProvider)}, nil
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
	p.LogProvider.Record(ctx, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.newRelicApp.RecordCustomEvent(eventName, metrics)
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}
