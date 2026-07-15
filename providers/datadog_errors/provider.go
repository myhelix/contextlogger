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
calls do not become Error Tracking issues.

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
}

func LogProvider(nextProvider providers.LogProvider) providers.LogProvider {
	return provider{chaining.LogProvider(nextProvider)}
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

// injectIfReport adds Datadog error.* fields when this is a report and the
// first argument is an error.
func (p provider) injectIfReport(ctx context.Context, report bool, args []interface{}) context.Context {
	if !report || len(args) != 1 {
		return ctx
	}
	err, ok := args[0].(error)
	if !ok {
		return ctx
	}
	// merry.Wrap generates a stack for non-merry errors and preserves an
	// existing one for merry errors.
	wrapped := merry.Wrap(err)
	return log.ContextWithFields(ctx, log.Fields{
		FieldErrorKind:    errorKind(err),
		FieldErrorMessage: err.Error(),
		FieldErrorStack:   merry.Stacktrace(wrapped),
	})
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Error(p.injectIfReport(ctx, report, args), report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Warn(p.injectIfReport(ctx, report, args), report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Info(p.injectIfReport(ctx, report, args), report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Debug(p.injectIfReport(ctx, report, args), report, args...)
}
