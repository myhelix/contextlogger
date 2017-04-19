// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package provides a basic dumb logger to get bootstrapped with
*/
package dummy

import (
	"github.com/myhelix/contextlogger/providers"

	"context"
	"fmt"
	"io"
)

type provider struct {
	io.Writer
}

func LogProvider(writer io.Writer) providers.LogProvider {
	return provider{writer}
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	fmt.Fprintln(p, args...)
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	fmt.Fprintln(p, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	fmt.Fprintln(p, eventName, metrics)
}

func (p provider) Wait() {}
