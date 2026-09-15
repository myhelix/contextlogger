// © 2016-2026 Helix OpCo LLC. All rights reserved.

/*
Package datadog_errors injects the attributes that Datadog Error Tracking needs
to create and group an error issue from a log event.

Datadog Error Tracking ingests an error only when the log carries:
  - error.kind    : the error type (used for grouping)
  - error.message : the human-readable message
  - error.stack   : the stack trace

This provider derives those from the first error argument and adds them to the
log context, but ONLY when the call is a report (ErrorReport / WarnReport, i.e.
report == true). Non-reported logs are left untouched so that ordinary Error()
calls do not become Error Tracking issues. Reported warnings retain warning
severity by default; callers may opt in to promoting reported warnings that
contain an actual error to error severity.

Pairs with the `reportable` provider (PR #34), which adds the `reportableError`
tag used to filter/monitor the reported subset. Wire both into the chain.
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

// Options controls behavior that is not safe to enable for every existing
// consumer of this provider.
type Options struct {
	// PromoteReportedWarnings emits WarnReport calls at error severity, but only
	// when the arguments contain an actual error. This makes those events
	// eligible for Datadog Error Tracking without turning string-only warnings
	// into errors.
	PromoteReportedWarnings bool
}

// LogProvider preserves the historical severity of WarnReport calls.
func LogProvider(nextProvider providers.LogProvider) providers.LogProvider {
	return LogProviderWithOptions(nextProvider, Options{})
}

// LogProviderWithOptions constructs a provider with explicit behavior for
// reported warnings.
func LogProviderWithOptions(nextProvider providers.LogProvider, options Options) providers.LogProvider {
	return provider{
		LogProvider: chaining.LogProvider(nextProvider),
		options:     options,
	}
}

// errorKind returns the Go type name of err for use as error.kind (the Datadog
// grouping key).
//
// NOTE: for merry-wrapped errors this reports merry's wrapper type (e.g.
// merry.errImpl) rather than the underlying cause. Reporting the real
// underlying type would improve grouping, but the merry version pinned by this
// library does not expose a stable unwrap helper; revisit once merry is
// upgraded. See DSI-1454.
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

// firstError returns the first non-nil error among args, or nil if there is
// none. It guards against typed-nil interfaces (e.g. a (*myErr)(nil) stored in
// an error interface), which satisfy the error type assertion but panic on
// err.Error(); such values are treated as absent.
func firstError(args []interface{}) error {
	for _, a := range args {
		err, ok := a.(error)
		if !ok || err == nil {
			continue
		}
		// Catch typed-nil: a non-nil interface wrapping a nil pointer.
		if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
			continue
		}
		return err
	}
	return nil
}

// injectIfReport adds Datadog error.* fields when this is a report and one of
// the args is a non-nil error. It scans all args (not just the first) so that
// calls like ErrorReport(err, "context", kv...) still enrich the event.
func (p provider) injectIfReport(ctx context.Context, report bool, args []interface{}) (context.Context, bool) {
	if !report {
		return ctx, false
	}
	err := firstError(args)
	if err == nil {
		return ctx, false
	}
	// merry.Wrap generates a stack for non-merry errors and preserves an
	// existing one for merry errors.
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

// Info and Debug deliberately do NOT inject error.* fields, even on report=true
// (InfoReport/DebugReport). This provider's contract is ErrorReport/WarnReport
// only. In practice Datadog Error Tracking wouldn't ingest info/debug anyway
// (its status gate is ERROR/CRITICAL/ALERT/EMERGENCY), but we skip injection so
// the fields don't appear as noise on sub-error logs.
func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Info(ctx, report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Debug(ctx, report, args...)
}
