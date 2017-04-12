// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This logger reports errors to Rollbar if they came in via {Error,Warn,Info}Report, then passes
through for logging by the base logger, if any. It uses the stack stored by merry, if the error
is a merry error; otherwise it generates a new one based on the reporting callstack.
*/
package rollbar

import (
	"github.com/ansel1/merry"
	goerr "github.com/go-errors/errors"
	"github.com/myhelix/rollbar"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"

	"context"
	"errors"
	"fmt"
	"net/http"
)

type provider struct {
	nextProvider providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider) (providers.LogProvider, error) {
	if nextProvider == nil {
		return nil, errors.New("Rollbar log provider requires a base provider")
	}
	if rollbar.Token == "" {
		return nil, errors.New("Rollbar is not configured (no token)")
	}
	return provider{nextProvider}, nil
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

func consolidateErrsForRollbar(errs []interface{}) (err error, stack rollbar.Stack, fields []*rollbar.Field) {
	err = listOfOneError(errs)
	if err == nil {
		// What was passed in wasn't an error, but we need an error
		err = merry.New(fmt.Sprint(errs...))
	}
	// We need a stack here; so wrap the error if it wasn't already
	mStack := merry.Stack(merry.Wrap(err))
	for _, f := range mStack {
		sf := goerr.NewStackFrame(f)
		stack = append(stack, rollbar.Frame{
			Filename: rollbar.ShortenFilePath(sf.File),
			Method:   sf.Name,
			Line:     sf.LineNumber,
		})
	}

	vals := merry.Values(err)
	stringMap := make(map[string]interface{})
	for k, v := range vals {
		if ks, ok := k.(string); ok {
			if ks != "message" {
				stringMap[ks] = v
			}
		}
	}
	fields = append(fields, &rollbar.Field{Name: "errorValues", Data: stringMap})
	return
}

func (p provider) reportToRollbar(ctx context.Context, level string, errs ...interface{}) {
	err, stack, fields := consolidateErrsForRollbar(errs)

	// Copy fields from the provider onto the error for reporting
	logValues := make(map[string]interface{})
	fields = append(fields, &rollbar.Field{Name: "logValues", Data: logValues})
	for k, v := range log.FieldsFromContext(ctx) {
		logValues[k] = v
	}

	if req := requestFrom(ctx); req != nil {
		rollbar.RequestErrorWithStack(level, req, err, stack, fields...)
	} else {
		rollbar.ErrorWithStack(level, err, stack, fields...)
	}
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.ERR, args...)
	}
	p.nextProvider.Error(ctx, report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.WARN, args...)
	}
	p.nextProvider.Warn(ctx, report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.INFO, args...)
	}
	p.nextProvider.Info(ctx, report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	if report {
		p.reportToRollbar(ctx, rollbar.DEBUG, args...)
	}
	p.nextProvider.Debug(ctx, report, args...)
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.nextProvider.Record(ctx, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.nextProvider.RecordEvent(ctx, eventName, metrics)
}

func (p provider) Wait() {
	rollbar.Wait()
	p.nextProvider.Wait()
}
