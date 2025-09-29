// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This extracts merry Values into logger Fields, then passes along to the base logger
*/
package merry

import (
	"github.com/ansel1/merry/v2"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"

	"context"
)

type provider struct {
	providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider) providers.LogProvider {
	return provider{chaining.LogProvider(nextProvider)}
}

// Extract fields from merry error values if input was exactly one error
func (p provider) extractContext(ctx context.Context, args []interface{}, includeTrace bool) context.Context {
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			fields := make(log.Fields)
			for key, val := range merry.Values(err) {
				// Allow all string values into log message
				if key, ok := key.(string); ok {
					fields[key] = val
				}
				// Add userMessage
				if usrMsg := merry.UserMessage(err); usrMsg != "" {
					fields["userMessage"] = usrMsg
				}
			}
			// Call merry.Wrap to generate trace for non-merry errors; that trace will be to
			// here, not to where the error was generated, but better than nothing.
			wrapped := merry.Wrap(err)
			// Put stack into context, for providers that might need it (e.g. Rollbar)
			ctx = log.ContextWithStack(ctx, merry.Stack(wrapped))
			if includeTrace {
				// Use tilde to sort stacktrace last, which at least for logrus is more readable
				fields["~stackTrace"] = merry.Stacktrace(wrapped)
			}
			return log.ContextWithFields(ctx, fields)
		}
	}
	// No error found
	return ctx
}

// We always extract merry Values from an error, but only for Error level do we print a traceback
func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Error(p.extractContext(ctx, args, true), report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Warn(p.extractContext(ctx, args, false), report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Info(p.extractContext(ctx, args, false), report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Debug(p.extractContext(ctx, args, false), report, args...)
}
