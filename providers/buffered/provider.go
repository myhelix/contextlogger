// © 2017 Helix OpCo LLC. All rights reserved.
// Initial Author: jpecknerhelix

package buffered

import (
	"context"
	"io"

	"github.com/Sirupsen/logrus"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
)

type bufferedLogProvider struct {
	*logrus.Entry
}

// LogProvider returns a LogProvider which outputs to writer, including the value of the "report" parameter
// in methods which have it.
func LogProvider(writer io.Writer) providers.LogProvider {
	return bufferedLogProvider{
		logrus.NewEntry(&logrus.Logger{
			Out: writer,
			Formatter: &logrus.TextFormatter{
				DisableColors:   true,
				TimestampFormat: "sometime", // Omit timestamp to make output predictable
			},
			Hooks: make(logrus.LevelHooks),
			Level: logrus.DebugLevel,
		}),
	}
}

func (p bufferedLogProvider) entryFor(ctx context.Context) *logrus.Entry {
	return p.Entry.WithFields(logrus.Fields(log.FieldsFromContext(ctx)))
}

func (p bufferedLogProvider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).WithField("report", report).Error(args)
}

func (p bufferedLogProvider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).WithField("report", report).Warn(args)
}

func (p bufferedLogProvider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).WithField("report", report).Info(args)
}

func (p bufferedLogProvider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).WithField("report", report).Debug(args)
}

func (p bufferedLogProvider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.Entry.WithFields(metrics).Info("Reporting metrics")
}

func (p bufferedLogProvider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.Entry.WithField("eventName", eventName).WithFields(metrics).Info("Reporting metrics")
}

func (p bufferedLogProvider) Wait() {}
