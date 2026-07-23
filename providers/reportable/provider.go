// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

package reportable

import (
	"context"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/chaining"
)

const FieldName = "reportableError"

type provider struct {
	providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider) providers.LogProvider {
	return provider{chaining.LogProvider(nextProvider)}
}

func (p provider) injectIfReport(ctx context.Context, report bool) context.Context {
	if report {
		return log.ContextWithFields(ctx, log.Fields{
			FieldName: true,
		})
	}
	return ctx
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Error(p.injectIfReport(ctx, report), report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Warn(p.injectIfReport(ctx, report), report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Info(p.injectIfReport(ctx, report), report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.LogProvider.Debug(p.injectIfReport(ctx, report), report, args...)
}
