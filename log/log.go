// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package provides the generic logging interface; it is broadly compatible with logrus, which is
itself generally compatible with the standard library logger.
*/
package log

import (
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/dummy"

	"context"
	"os"
	"time"
)

type Metrics map[string]interface{}
type Fields map[string]interface{}

var defaultProvider providers.LogProvider

func SetDefaultProvider(provider providers.LogProvider) {
	defaultProvider = provider
}

func DefaultProvider() providers.LogProvider {
	return defaultProvider
}

// Start with something, to avoid crashing before we're configured
func init() {
	defaultProvider = dummy.LogProvider(os.Stderr)
}

var (
	errorReportFields = Fields{
		"reportableError": true,
		"reportedLevel":   "error",
	}
	warnReportFields = Fields{
		"reportableError": true,
		"reportedLevel":   "warning",
	}
)

/* Keys for Context Values */
type contextLogProviderKey struct{}
type contextLogFieldsKey struct{}
type contextStackKey struct{}

/*
ContextLoggers are designed to be passed around for convenience within a given project; APIs
which are meant to be used from other projects should use the standard golang context.Context
interface instead, for broad compatibility. You can always recover a ContextLogger from a Context
using FromContext below.
*/
type ContextLogger interface {
	context.Context
	ComposableLogger
}

/*
This can be combined with an object that already supplies context.Context, without conflict,
to create a ContextLogger; it's not really intended to be used standalone, and in most cases
you probably want to include ContextLogger rather than this in other interfaces/structs.

In particular, if you are composing this with another supplier of Context, you need to keep in mind
that a ContextLogger using the implementation in this file won't have access to your Context that
lives next to it. So usually what you would want to do instead is use an interface to mask out the
Context part of the ContextLogger's neighbor, then derive the ContextLogger instance from the
source object that you've masked the context out of for composition.

TODO: show example of what I'm talking about
*/
type ComposableLogger interface {
	LogProvider() providers.LogProvider

	/* Methods passed through to LogProvider with added context */
	ErrorReport(args ...interface{})
	Error(args ...interface{})
	WarnReport(args ...interface{})
	Warn(args ...interface{})
	InfoReport(args ...interface{})
	Info(args ...interface{})
	DebugReport(args ...interface{})
	Debug(args ...interface{})
	Record(metrics Metrics)
	RecordEvent(eventName string, metrics Metrics)

	// Add log data to a context to be used with future log messages
	WithField(key string, val interface{}) ContextLogger
	WithFields(fields Fields) ContextLogger
}

type contextLogger struct {
	context.Context
	provider providers.LogProvider
}

func (c contextLogger) LogProvider() providers.LogProvider {
	return c.provider
}
func (c contextLogger) ErrorReport(args ...interface{}) {
	c.provider.Error(contextWithReportFields(c.Context, errorReportFields), true, args...)
}
func (c contextLogger) Error(args ...interface{}) {
	c.provider.Error(c.Context, false, args...)
}
func (c contextLogger) WarnReport(args ...interface{}) {
	c.provider.Warn(contextWithReportFields(c.Context, warnReportFields), true, args...)
}
func (c contextLogger) Warn(args ...interface{}) {
	c.provider.Warn(c.Context, false, args...)
}

// InfoReport does not inject report metadata into context: only Error/Warn
// reports are treated as reportable. report=true is still passed through to
// the provider so providers can distinguish *Report calls if they choose to.
func (c contextLogger) InfoReport(args ...interface{}) {
	c.provider.Info(c.Context, true, args...)
}
func (c contextLogger) Info(args ...interface{}) {
	c.provider.Info(c.Context, false, args...)
}

// DebugReport does not inject reportFields into context: only Error/Warn
// reports are treated as reportable. report=true is still passed through to
// the provider so providers can distinguish *Report calls if they choose to.
func (c contextLogger) DebugReport(args ...interface{}) {
	c.provider.Debug(c.Context, true, args...)
}
func (c contextLogger) Debug(args ...interface{}) {
	c.provider.Debug(c.Context, false, args...)
}
func (c contextLogger) Record(metrics Metrics) {
	c.provider.Record(c.Context, metrics)
}
func (c contextLogger) RecordEvent(eventName string, metrics Metrics) {
	c.provider.RecordEvent(c.Context, eventName, metrics)
}
func (c contextLogger) WithField(key string, val interface{}) ContextLogger {
	fields := make(Fields)
	fields[key] = val
	return c.WithFields(fields)
}
func (c contextLogger) WithFields(fields Fields) ContextLogger {
	return contextLogger{ContextWithFields(c.Context, fields), c.provider}
}

// This is mostly for use by LogProviders; adds fields to a raw context.Context
// If you're looking to derive from the default ContextLogger, you want log.WithFields
func ContextWithFields(ctx context.Context, fields Fields) context.Context {
	var combinedFields = make(Fields)
	if existingFields, ok := ctx.Value(contextLogFieldsKey{}).(Fields); ok {
		for k, v := range existingFields {
			combinedFields[k] = v
		}
	}
	for k, v := range fields {
		combinedFields[k] = v
	}
	return context.WithValue(ctx, contextLogFieldsKey{}, combinedFields)
}

func FieldsFromContext(ctx context.Context) Fields {
	if fields, ok := ctx.Value(contextLogFieldsKey{}).(Fields); ok {
		return fields
	}
	return make(Fields)
}

// detachedContext embeds an internally-owned context for Deadline/Done/Err
// (so those retain correct stdlib semantics, e.g. DeadlineExceeded vs
// Canceled) and overrides only Value, which never reads from the context
// Detach was built from.
type detachedContext struct {
	context.Context
	fields   Fields
	provider providers.LogProvider
}

func (d *detachedContext) Value(key interface{}) interface{} {
	switch key {
	case contextLogFieldsKey{}:
		return d.fields
	case contextLogProviderKey{}:
		return d.provider
	}
	return nil
}

// Detach returns a context equivalent to ctx that never reads from ctx again.
// Use it before handing a context to a goroutine or outbound call that must
// outlive the current request — unlike WithFields/DeriveBackgroundJob, this
// doesn't just layer a value on top of ctx, since Value() on that kind of
// child still delegates back to ctx, which is unsafe if ctx wraps something
// reused across requests (e.g. a pooled *gin.Context).
func Detach(ctx context.Context) context.Context {
	var derived context.Context
	var cancel context.CancelFunc
	deadline, hasDL := ctx.Deadline()
	if hasDL {
		derived, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		derived, cancel = context.WithCancel(context.Background())
	}

	// derived already has the same deadline as ctx, so its own timer will
	// fire DeadlineExceeded on its own. Only forward ctx's Done() as an
	// explicit cancel when it closes before that deadline — otherwise this
	// goroutine could race derived's timer and overwrite a correct
	// DeadlineExceeded with Canceled.
	if done := ctx.Done(); done != nil {
		go func() {
			select {
			case <-done:
				if !hasDL || time.Now().Before(deadline) {
					cancel()
				}
			case <-derived.Done():
			}
		}()
	}

	provider, _ := ctx.Value(contextLogProviderKey{}).(providers.LogProvider)
	return &detachedContext{
		Context:  derived,
		fields:   FieldsFromContext(ctx),
		provider: provider,
	}
}

func contextWithReportFields(ctx context.Context, fields Fields) context.Context {
	return ContextWithFields(ctx, fields)
}

func ContextWithStack(ctx context.Context, stack []uintptr) context.Context {
	return context.WithValue(ctx, contextStackKey{}, stack)
}

func StackFromContext(ctx context.Context) []uintptr {
	if stack, ok := ctx.Value(contextStackKey{}).([]uintptr); ok {
		return stack
	}
	return nil
}

func FromContext(ctx context.Context) ContextLogger {
	if provider, ok := ctx.Value(contextLogProviderKey{}).(providers.LogProvider); ok {
		return contextLogger{ctx, provider}
	} else {
		return contextLogger{ctx, defaultProvider}
	}
}

func FromContextAndProvider(ctx context.Context, provider providers.LogProvider) ContextLogger {
	newContext := context.WithValue(ctx, contextLogProviderKey{}, provider)
	return contextLogger{newContext, provider}
}

func BackgroundContext() ContextLogger {
	return FromContext(context.Background())
}

/* package versions of functions, operate on default log provider and background context */

func ErrorReport(args ...interface{}) {
	BackgroundContext().ErrorReport(args...)
}

func WarnReport(args ...interface{}) {
	BackgroundContext().WarnReport(args...)
}

func InfoReport(args ...interface{}) {
	BackgroundContext().InfoReport(args...)
}

func DebugReport(args ...interface{}) {
	BackgroundContext().DebugReport(args...)
}

func Error(args ...interface{}) {
	BackgroundContext().Error(args...)
}

func Warn(args ...interface{}) {
	BackgroundContext().Warn(args...)
}

func Info(args ...interface{}) {
	BackgroundContext().Info(args...)
}

func Debug(args ...interface{}) {
	BackgroundContext().Debug(args...)
}

func Record(metrics Metrics) {
	BackgroundContext().Record(metrics)
}

func RecordEvent(eventName string, metrics Metrics) {
	BackgroundContext().RecordEvent(eventName, metrics)
}

func WithField(key string, val interface{}) ContextLogger {
	return BackgroundContext().WithField(key, val)
}

func WithFields(fields Fields) ContextLogger {
	return BackgroundContext().WithFields(fields)
}

func Wait() {
	BackgroundContext().LogProvider().Wait()
}
