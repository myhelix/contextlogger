// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This extracts merry Values into logger Fields, then passes along to the base logger
*/
package merry

import (
	"github.com/ansel1/merry"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"

	"context"
)

type provider struct {
	nextProvider providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider) (providers.LogProvider, merry.Error) {
	if nextProvider == nil {
		return nil, merry.New("Merry log provider requires a base provider")
	}
	return provider{nextProvider}, nil
}

// Extract fields from merry error values if input was exactly one error
func (p provider) extractContextAndError(ctx context.Context, args []interface{}) (context.Context, merry.Error) {
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			fields := make(log.Fields)
			for key, val := range merry.Values(err) {
				if key, ok := key.(string); ok {
					switch key {
					case "stack", "message":
					// Merry built-ins; ignore
					default:
						fields[key] = val
					}
				}
			}
			// Always turn the error into a merry error, to provide best-effort traceback for basic errors
			return log.ContextWithFields(ctx, fields), merry.Wrap(err)
		}
	}
	return ctx, nil
}

// We always extract merry Values from an error, but only for Error level do we print a traceback
func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	ctx, err := p.extractContextAndError(ctx, args)
	if err != nil {
		args = []interface{}{merry.Details(err)}
	}
	p.nextProvider.Error(ctx, report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	ctx, _ = p.extractContextAndError(ctx, args)
	p.nextProvider.Warn(ctx, report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	ctx, _ = p.extractContextAndError(ctx, args)
	p.nextProvider.Info(ctx, report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	ctx, _ = p.extractContextAndError(ctx, args)
	p.nextProvider.Debug(ctx, report, args...)
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.nextProvider.Record(ctx, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.nextProvider.RecordEvent(ctx, eventName, metrics)
}

func (p provider) Wait() {
	p.nextProvider.Wait()
}
