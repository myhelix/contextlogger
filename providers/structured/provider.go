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
	CallType      providers.LogCallType
}

type RecordCallArgs struct {
	ContextFields log.Fields
	Metrics       log.Metrics
}

type RecordEventCallArgs struct {
	ContextFields log.Fields
	EventName     string
	Metrics       log.Metrics
}

type StructuredOutputLogProvider struct {
	providers.LogProvider

	logCalls         []LogCallArgs
	recordCalls      []RecordCallArgs
	recordEventCalls []RecordEventCallArgs
	logMutex         sync.RWMutex
	metricMutex      sync.Mutex
	eventMutex       sync.Mutex
}

func (p StructuredOutputLogProvider) GetLogCallsByCallType(callType providers.LogCallType) []LogCallArgs {
	var result []LogCallArgs

	p.logMutex.RLock()
	defer p.logMutex.RUnlock()

	for _, log := range p.logCalls {
		if log.CallType == callType {
			result = append(result, log)
		}
	}
	return result
}

func (p StructuredOutputLogProvider) GetRecordCalls() []RecordCallArgs {
	return p.recordCalls
}

func (p StructuredOutputLogProvider) GetRecordEventCalls() []RecordEventCallArgs {
	return p.recordEventCalls
}

// NewStructuredOutputLogProvider returns a LogProvider which records all calls made to it.
// Deprecated: Clients should call LogProvider instead
func NewStructuredOutputLogProvider() *StructuredOutputLogProvider {
	return LogProvider(nil)
}

func LogProvider(nextProvider providers.LogProvider) *StructuredOutputLogProvider {
	return &StructuredOutputLogProvider{
		LogProvider:      chaining.LogProvider(nextProvider),
		logCalls:         []LogCallArgs{},
		recordCalls:      []RecordCallArgs{},
		recordEventCalls: []RecordEventCallArgs{},
	}
}

func (p *StructuredOutputLogProvider) saveLogCallArgs(
	callType providers.LogCallType,
	ctx context.Context,
	report bool,
	args ...interface{},
) {
	lc := LogCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		Report:        report,
		Args:          args,
		CallType:      callType,
	}
	p.logMutex.Lock()
	defer p.logMutex.Unlock()
	p.logCalls = append(p.logCalls, lc)
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
	callArgs := RecordCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		Metrics:       metrics,
	}
	p.metricMutex.Lock()
	defer p.metricMutex.Unlock()
	p.recordCalls = append(p.recordCalls, callArgs)
	p.LogProvider.Record(ctx, metrics)
}

func (p *StructuredOutputLogProvider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	callArgs := RecordEventCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		EventName:     eventName,
		Metrics:       metrics,
	}
	p.eventMutex.Lock()
	defer p.eventMutex.Unlock()
	p.recordEventCalls = append(p.recordEventCalls, callArgs)
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}
