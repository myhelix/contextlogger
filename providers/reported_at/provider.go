// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This adds a field to log reports showing where the logger was called from
*/
package reported_at

import (
	"github.com/calm/contextlogger/log"
	"github.com/calm/contextlogger/providers"
	"github.com/calm/contextlogger/providers/chaining"

	"context"
	"fmt"
	"regexp"
	"runtime"
)

type provider struct {
	config *Config
	providers.LogProvider
}

type Config struct {
	// Ignore these stack frames for purposes of reportedAt
	IgnoreStackFrames *regexp.Regexp
}

var RecommendedConfig = Config{
	IgnoreStackFrames: regexp.MustCompile("calm/contextlogger"),
}

var alwaysIgnore = regexp.MustCompile("<autogenerated>")

func LogProvider(nextProvider providers.LogProvider, config Config) providers.LogProvider {
	return provider{&config, chaining.LogProvider(nextProvider)}
}

func (p provider) reportedAt(ctx context.Context) context.Context {
	pc := make([]uintptr, 50)
	runtime.Callers(1, pc)
	frameData := runtime.FuncForPC(pc[0])
	thisFile, _ := frameData.FileLine(pc[0])
	for _, frame := range pc {
		frameData := runtime.FuncForPC(frame)
		file, line := frameData.FileLine(frame)
		if file != thisFile && !alwaysIgnore.MatchString(file) &&
			(p.config.IgnoreStackFrames == nil || !p.config.IgnoreStackFrames.MatchString(file)) {
			return log.ContextWithFields(ctx, log.Fields{
				"reportedAt": fmt.Sprintf("%s:%d", file, line),
			})
		}
	}
	// Just in case we don't find anything
	return ctx
}

// We always extract merry Values from an error, but only for Error level do we print a traceback
func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Error(p.reportedAt(ctx), report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Warn(p.reportedAt(ctx), report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Info(p.reportedAt(ctx), report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Debug(p.reportedAt(ctx), report, args...)
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.LogProvider.Record(p.reportedAt(ctx), metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.LogProvider.RecordEvent(p.reportedAt(ctx), eventName, metrics)
}
