// © 2017 Helix OpCo LLC. All rights reserved.
// Initial Author: jpecknerhelix

package structured

import (
	"context"
	"sync"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
)

type LogCallArgs struct {
	ContextFields log.Fields
	Report        bool
	Args          []interface{}
	Level         providers.LogLevel
}

type RecordCallArgs struct {
	ContextFields log.Fields
	Metrics       log.Metrics
	EventName     string
}

type StructuredOutputLogProvider struct {
	providers.LogProvider

	logCalls    []*LogCallArgs
	logMutex    sync.RWMutex
	recordCalls []*RecordCallArgs
	recordMutex sync.RWMutex
}

// Return list of log calls, filtered to only selected levels (if any present)
func (p StructuredOutputLogProvider) LogCalls(levels ...providers.LogLevel) (result []*LogCallArgs) {
	p.logMutex.RLock()
	defer p.logMutex.RUnlock()

	allLevels := len(levels) == 0
	for _, call := range p.logCalls {
		if allLevels {
			result = append(result, call)
		} else {
			for _, level := range levels {
				if call.Level == level {
					result = append(result, call)
					break
				}
			}
		}
	}
	return
}

func (p StructuredOutputLogProvider) RecordCalls() []*RecordCallArgs {
	p.recordMutex.RLock()
	defer p.recordMutex.RUnlock()

	result := make([]*RecordCallArgs, len(p.recordCalls))

	for i, call := range p.recordCalls {
		result[i] = call
	}
	return result
}

func LogProvider(nextProvider providers.LogProvider) *StructuredOutputLogProvider {
	return &StructuredOutputLogProvider{
		LogProvider: chaining.LogProvider(nextProvider),
		logCalls:    []*LogCallArgs{},
		recordCalls: []*RecordCallArgs{},
	}
}

func (p *StructuredOutputLogProvider) saveLogCallArgs(
	level providers.LogLevel,
	ctx context.Context,
	report bool,
	args ...interface{},
) {
	callArgs := LogCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		Report:        report,
		Args:          args,
		Level:         level,
	}
	p.logMutex.Lock()
	defer p.logMutex.Unlock()
	p.logCalls = append(p.logCalls, &callArgs)
}

func (p *StructuredOutputLogProvider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.saveLogCallArgs(providers.Error, ctx, report, args...)
	p.LogProvider.Error(ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.saveLogCallArgs(providers.Warn, ctx, report, args...)
	p.LogProvider.Warn(ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.saveLogCallArgs(providers.Info, ctx, report, args...)
	p.LogProvider.Info(ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.saveLogCallArgs(providers.Debug, ctx, report, args...)
	p.LogProvider.Debug(ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.LogProvider.Record(ctx, metrics)

	callArgs := RecordCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		Metrics:       metrics,
	}
	p.recordMutex.Lock()
	defer p.recordMutex.Unlock()
	p.recordCalls = append(p.recordCalls, &callArgs)
}

func (p *StructuredOutputLogProvider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.LogProvider.RecordEvent(ctx, eventName, metrics)

	callArgs := RecordCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		EventName:     eventName,
		Metrics:       metrics,
	}
	p.recordMutex.Lock()
	defer p.recordMutex.Unlock()
	p.recordCalls = append(p.recordCalls, &callArgs)
}
