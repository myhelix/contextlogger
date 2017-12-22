// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

package providers

import (
	"context"
)

type RawLogCallType int

const (
	Error RawLogCallType = iota
	Warn
	Info
	Debug
)

func RawLogCallTypes() []RawLogCallType {
	return []RawLogCallType{
		Error,
		Warn,
		Info,
		Debug,
	}
}

type LogProvider interface {
	Error(ctx context.Context, report bool, args ...interface{})
	Warn(ctx context.Context, report bool, args ...interface{})
	Info(ctx context.Context, report bool, args ...interface{})
	Debug(ctx context.Context, report bool, args ...interface{})

	// Record metrics, or events with metrics
	Record(ctx context.Context, metrics map[string]interface{})
	RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{})

	// Wait for any asynchronous logging processes to complete; good to call before exiting program
	Wait()
}
