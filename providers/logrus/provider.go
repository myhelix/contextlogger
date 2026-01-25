// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package provides a concrete implementation of LogProvider using Logrus
*/
package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"

	"context"
	"encoding/json"
	"io"
	"time"
)

type provider struct {
	*logrus.Entry
	providers.LogProvider
}

type Config struct {
	Output               io.Writer
	Level                string
	Formatter            logrus.Formatter
	FlattenDataDogFields bool // If true, promote DD fields to root level for DataDog Lambda Extension compatibility
}

var RecommendedFormatter = &logrus.TextFormatter{
	FullTimestamp:   true,
	DisableColors:   true,
	TimestampFormat: time.RFC3339Nano,
}

var JSONFormatter = &logrus.JSONFormatter{
	DisableTimestamp: false,
	TimestampFormat:  time.RFC3339Nano,
}

// dataDogJSONFormatter wraps a logrus formatter to promote DataDog-specific fields
// to the root level of JSON output for DataDog Lambda Extension compatibility
type dataDogJSONFormatter struct {
	wrapped logrus.Formatter
}

// DataDog fields that should be promoted to root level
var dataDogFields = map[string]bool{
	"dd.trace_id":       true,
	"dd.span_id":        true,
	"lambda.request_id": true,
}

func (f *dataDogJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// First, format using the wrapped formatter to get standard output
	data, err := f.wrapped.Format(entry)
	if err != nil {
		return nil, err
	}

	// If there are no fields, return as-is
	if len(entry.Data) == 0 {
		return data, nil
	}

	// Parse the JSON output
	var output map[string]interface{}
	if err := json.Unmarshal(data, &output); err != nil {
		return data, nil // If we can't parse, return original
	}

	// Extract DD fields and regular context fields
	ddFieldsFound := make(map[string]interface{})
	contextFields := make(map[string]interface{})

	for key, value := range entry.Data {
		if dataDogFields[key] {
			ddFieldsFound[key] = value
		} else {
			contextFields[key] = value
		}
	}

	// If no DD fields found, return original output
	if len(ddFieldsFound) == 0 {
		return data, nil
	}

	// Remove all entry.Data fields from output (they were added by logrus formatter)
	for key := range entry.Data {
		delete(output, key)
	}

	// Add DD fields at root level
	for key, value := range ddFieldsFound {
		output[key] = value
	}

	// Add remaining fields under "context" key if any exist
	if len(contextFields) > 0 {
		output["context"] = contextFields
	}

	// Re-serialize to JSON
	result, err := json.Marshal(output)
	if err != nil {
		return data, nil // If we can't marshal, return original
	}

	// Add newline to match logrus behavior
	return append(result, '\n'), nil
}

func LogProvider(nextProvider providers.LogProvider, config Config) (l providers.LogProvider, err error) {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return
	}

	formatter := config.Formatter
	// Wrap formatter with DataDog field flattening if enabled
	if config.FlattenDataDogFields && formatter != nil {
		formatter = &dataDogJSONFormatter{wrapped: formatter}
	}

	l = provider{logrus.NewEntry(&logrus.Logger{
		Out:       config.Output,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}), chaining.LogProvider(nextProvider)}
	return
}

func (p provider) entryFor(ctx context.Context) *logrus.Entry {
	return p.Entry.WithFields(logrus.Fields(log.FieldsFromContext(ctx)))
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Error(args...)
	p.LogProvider.Error(ctx, report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Warn(args...)
	p.LogProvider.Warn(ctx, report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Info(args...)
	p.LogProvider.Info(ctx, report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Debug(args...)
	p.LogProvider.Debug(ctx, report, args...)
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.Entry.WithFields(metrics).Info("Reporting metrics")
	p.LogProvider.Record(ctx, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.Entry.WithField("eventName", eventName).WithFields(metrics).Info("Reporting metrics")
	p.LogProvider.RecordEvent(ctx, eventName, metrics)
}

func (p provider) Wait() {
	p.LogProvider.Wait()
}
