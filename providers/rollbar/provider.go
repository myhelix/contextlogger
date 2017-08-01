// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This logger reports errors to Rollbar if they came in via {Error,Warn,Info}Report, then passes
through for logging by the base logger, if any. It uses the stack stored by merry, if the error
is a merry error; otherwise it generates a new one based on the reporting callstack.
*/
package rollbar

import (
	goerr "github.com/go-errors/errors"
	"github.com/myhelix/rollbar"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"

	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

type provider struct {
	providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider) (providers.LogProvider, error) {
	if rollbar.Token == "" {
		return nil, errors.New("Rollbar is not configured (no token)")
	}
	return provider{chaining.LogProvider(nextProvider)}, nil
}

type contextRequestKey struct{}

func WithRequest(ctx context.Context, req *http.Request) log.ContextLogger {
	return log.FromContext(context.WithValue(ctx, contextRequestKey{}, req))
}

func requestFrom(ctx context.Context) *http.Request {
	if req, ok := ctx.Value(contextRequestKey{}).(*http.Request); ok {
		return req
	}
	return nil
}

// If our list of arbitrary things is actually one error, return that error
func listOfOneError(errs []interface{}) error {
	if len(errs) == 1 {
		if err, ok := errs[0].(error); ok {
			return err
		}
	}
	return nil
}

func consolidateErrs(ctx context.Context, errs []interface{}) (err error, stack rollbar.Stack) {
	err = listOfOneError(errs)
	if err == nil {
		// What was passed in wasn't an error, but we need an error
		err = errors.New(fmt.Sprint(errs...))
	}
	goStack := log.StackFromContext(ctx)
	// If a stack wasn't saved in the context (e.g. by merry provider), then generate one
	if goStack == nil {
		goStack = make([]uintptr, 50)
		runtime.Callers(1, goStack)
	}
	for _, f := range goStack {
		sf := goerr.NewStackFrame(f)
		stack = append(stack, rollbar.Frame{
			Filename: rollbar.ShortenFilePath(sf.File),
			Method:   sf.Name,
			Line:     sf.LineNumber,
		})
	}
	return
}

func (p provider) reportToRollbar(ctx context.Context, level string, errs ...interface{}) {
	err, stack := consolidateErrs(ctx, errs)

	logFields := &rollbar.Field{
		Name: "logFields",
		Data: log.FieldsFromContext(ctx),
	}

	if req := requestFrom(ctx); req != nil {
		rollbar.RequestErrorWithStack(level, req, err, stack, logFields)
	} else {
		rollbar.ErrorWithStack(level, err, stack, logFields)
	}
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.ERR, args...)
	}
	p.LogProvider.Error(ctx, report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.WARN, args...)
	}
	p.LogProvider.Warn(ctx, report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.INFO, args...)
	}
	p.LogProvider.Info(ctx, report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.DEBUG, args...)
	}
	p.LogProvider.Debug(ctx, report, args...)
}

func (p provider) Wait() {
	rollbar.Wait()
	p.LogProvider.Wait()
}
