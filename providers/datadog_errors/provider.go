// © 2016-2026 Helix OpCo LLC. All rights reserved.

/*
Package datadog_errors adds Datadog Error Tracking fields to reported Error and
Warn events. Error values retain their type and stack; string-only reports use
a synthetic kind and message. Non-reported events are unchanged.

WarnReport keeps warning severity by default. Set PromoteReportedWarnings to
make reported warnings eligible for Datadog Error Tracking. Use with the
reportable provider to retain their original reported level.
*/
package datadog_errors

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ansel1/merry"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
)

// Datadog Error Tracking reserved attributes.
const (
	FieldErrorKind      = "error.kind"
	FieldErrorMessage   = "error.message"
	FieldErrorStack     = "error.stack"
	KindReportedError   = "ReportedError"
	KindReportedWarning = "ReportedWarning"
)

type provider struct {
	providers.LogProvider
	options Options
}

// Options configures Datadog error logging.
type Options struct {
	// PromoteReportedWarnings emits WarnReport calls at error severity.
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
		if !ok || isNilError(err) {
			continue
		}
		return err
	}
	return nil
}

func isNilError(err error) bool {
	if err == nil {
		return true
	}
	v := reflect.ValueOf(err)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

func reportMessage(args []interface{}) string {
	filtered := make([]interface{}, 0, len(args))
	for _, a := range args {
		if err, ok := a.(error); ok && isNilError(err) {
			continue
		}
		filtered = append(filtered, a)
	}
	return fmt.Sprint(filtered...)
}

// injectIfReport adds Datadog fields from the first error in a reported event,
// or synthesizes the minimum grouping fields for a string-only report.
func (p provider) injectIfReport(ctx context.Context, report bool, syntheticKind string, args []interface{}) (context.Context, bool) {
	if !report {
		return ctx, false
	}
	err := firstError(args)
	if err == nil {
		message := reportMessage(args)
		if message == "" {
			return ctx, false
		}
		return log.ContextWithFields(ctx, log.Fields{
			FieldErrorKind:    syntheticKind,
			FieldErrorMessage: message,
		}), true
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
	ctx, _ = p.injectIfReport(ctx, report, KindReportedError, args)
	p.LogProvider.Error(ctx, report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	ctx, hasErrorFields := p.injectIfReport(ctx, report, KindReportedWarning, args)
	if report && hasErrorFields && p.options.PromoteReportedWarnings {
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
