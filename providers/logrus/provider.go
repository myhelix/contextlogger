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
	"io"
	"time"

	dd_logrus "github.com/DataDog/dd-trace-go/contrib/sirupsen/logrus/v2"
)

type provider struct {
	*logrus.Entry
	providers.LogProvider
}

type Config struct {
	Output    io.Writer
	Level     string
	Formatter logrus.Formatter
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

func LogProvider(nextProvider providers.LogProvider, config Config) (l providers.LogProvider, err error) {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return
	}

	logger := &logrus.Logger{
		Out:       config.Output,
		Formatter: config.Formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	// Register Datadog context hook for trace correlation
	// This automatically injects dd.trace_id, dd.span_id, dd.service, dd.env, dd.version
	// when logging with a context that has an active tracer span
	logger.AddHook(&dd_logrus.DDContextLogHook{})

	l = provider{logrus.NewEntry(logger), chaining.LogProvider(nextProvider)}
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
