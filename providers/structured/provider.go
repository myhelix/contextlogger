// © 2017 Helix OpCo LLC. All rights reserved.
// Initial Author: jpecknerhelix

package structured

import (
	"context"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
)

type RawLogCallArgs struct {
	ContextFields log.Fields
	Report        bool
	Args          []interface{}
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
	rawLogCalls      map[providers.RawLogCallType][]RawLogCallArgs
	recordCalls      []RecordCallArgs
	recordEventCalls []RecordEventCallArgs
}

func (p StructuredOutputLogProvider) GetRawLogCalls() map[providers.RawLogCallType][]RawLogCallArgs {
	return p.rawLogCalls
}

func (p StructuredOutputLogProvider) GetRecordCalls() []RecordCallArgs {
	return p.recordCalls
}

func (p StructuredOutputLogProvider) GetRecordEventCalls() []RecordEventCallArgs {
	return p.recordEventCalls
}

// NewStructuredOutputLogProvider returns a LogProvider which records all calls made to it.
func NewStructuredOutputLogProvider() *StructuredOutputLogProvider {
	rawLogCalls := map[providers.RawLogCallType][]RawLogCallArgs{}
	for _, callType := range providers.RawLogCallTypes() {
		rawLogCalls[callType] = []RawLogCallArgs{}
	}

	return &StructuredOutputLogProvider{
		rawLogCalls:      rawLogCalls,
		recordCalls:      []RecordCallArgs{},
		recordEventCalls: []RecordEventCallArgs{},
	}
}

func (p *StructuredOutputLogProvider) saveRawCallArgs(
	callType providers.RawLogCallType,
	ctx context.Context,
	report bool,
	args ...interface{},
) {
	callArgs := RawLogCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		Report:        report,
		Args:          args,
	}
	p.rawLogCalls[callType] = append(p.rawLogCalls[callType], callArgs)
}

func (p *StructuredOutputLogProvider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.saveRawCallArgs(providers.Error, ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.saveRawCallArgs(providers.Warn, ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.saveRawCallArgs(providers.Info, ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.saveRawCallArgs(providers.Debug, ctx, report, args...)
}

func (p *StructuredOutputLogProvider) Record(ctx context.Context, metrics log.Metrics) {
	callArgs := RecordCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		Metrics:       metrics,
	}
	p.recordCalls = append(p.recordCalls, callArgs)
}

func (p *StructuredOutputLogProvider) RecordEvent(ctx context.Context, eventName string, metrics log.Metrics) {
	callArgs := RecordEventCallArgs{
		ContextFields: log.FieldsFromContext(ctx),
		EventName:     eventName,
		Metrics:       metrics,
	}
	p.recordEventCalls = append(p.recordEventCalls, callArgs)
}

func (p *StructuredOutputLogProvider) Wait() {}
