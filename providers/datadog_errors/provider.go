// © 2016-2026 Helix OpCo LLC. All rights reserved.

/*
Package datadog_errors adds error.kind, error.message, and error.stack to
reported Error and Warn events that contain an error. Non-reported events are
unchanged.

WarnReport keeps warning severity by default. Set PromoteReportedWarnings to
make reported warnings containing an error eligible for Datadog Error Tracking.
Use with the reportable provider to tag reported events.
*/
package datadog_errors

import (
	"context"
	"reflect"

	"github.com/ansel1/merry"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
)

// Datadog Error Tracking reserved attributes.
const (
	FieldErrorKind    = "error.kind"
	FieldErrorMessage = "error.message"
	FieldErrorStack   = "error.stack"
)

type provider struct {
	providers.LogProvider
	options Options
}

// Options configures Datadog error logging.
type Options struct {
	// PromoteReportedWarnings emits WarnReport calls containing an error at error severity.
	PromoteReportedWarnings bool
}

// LogProvider preserves WarnReport severity.
func LogProvider(nextProvider providers.LogProvider) providers.LogProvider {
	return LogProviderWithOptions(nextProvider, Options{})
}

// LogProviderWithOptions configures reported warning behavior.
func LogProviderWithOptions(nextProvider providers.LogProvider, options Options) providers.LogProvider {
	return provider{
		LogProvider: chaining.LogProvider(nextProvider),
		options:     options,
	}
}

// errorKind returns the Go type used by Datadog for grouping. Merry's current
// version has no stable unwrap API, so wrapped errors group by wrapper type.
// See DSI-1454.
func errorKind(err error) string {
	t := reflect.TypeOf(err)
	if t == nil {
		return "error"
	}
	if t.Kind() == reflect.Ptr {
		return t.Elem().String()
	}
	return t.String()
}

// firstError returns the first non-nil error and ignores typed-nil values.
func firstError(args []interface{}) error {
	for _, a := range args {
		err, ok := a.(error)
		if !ok || err == nil {
			continue
		}
		// Ignore error interfaces containing a nil pointer.
		if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
			continue
		}
		return err
	}
	return nil
}

// injectIfReport adds Datadog fields from the first error in a reported event.
func (p provider) injectIfReport(ctx context.Context, report bool, args []interface{}) (context.Context, bool) {
	if !report {
		return ctx, false
	}
	err := firstError(args)
	if err == nil {
		return ctx, false
	}
	// Add a stack to plain errors and preserve existing Merry stacks.
	wrapped := merry.Wrap(err)
	return log.ContextWithFields(ctx, log.Fields{
		FieldErrorKind:    errorKind(err),
		FieldErrorMessage: err.Error(),
		FieldErrorStack:   merry.Stacktrace(wrapped),
	}), true
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	ctx, _ = p.injectIfReport(ctx, report, args)
	p.LogProvider.Error(ctx, report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	ctx, hasError := p.injectIfReport(ctx, report, args)
	if report && hasError && p.options.PromoteReportedWarnings {
		p.LogProvider.Error(ctx, report, args...)
		return
	}
	p.LogProvider.Warn(ctx, report, args...)
}

// Info and Debug never add Error Tracking fields.
func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Info(ctx, report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Debug(ctx, report, args...)
}
