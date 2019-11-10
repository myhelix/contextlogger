// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package provides a basic dumb logger to get bootstrapped with
*/
package dummy

import (
	"github.com/calm/contextlogger/v2/providers"

	"context"
	"fmt"
	"io"
)

type provider struct {
	io.Writer
	waitState *WaitState
}

type WaitState struct {
	waiting bool
}

func (ws *WaitState) Set(waiting bool) {
	ws.waiting = waiting
}

func (ws *WaitState) Get() bool {
	return ws.waiting
}

func LogProvider(writer io.Writer) providers.LogProvider {
	return provider{writer, new(WaitState)}
}

func LogProviderWithWaitState(writer io.Writer, waitState *WaitState) providers.LogProvider {
	return provider{writer, waitState}
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

func (p provider) Wait() {
	p.waitState.Set(true)
}
