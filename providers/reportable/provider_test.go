// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

package reportable

import (
	"context"
	"testing"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/structured"
	. "github.com/onsi/gomega"
)

func TestReportTrueInjectsField(t *testing.T) {
	RegisterTestingT(t)

	capturer := structured.LogProvider(nil)
	testProvider := LogProvider(capturer)
	ctx := context.Background()

	testProvider.Error(ctx, true, "test error")

	calls := capturer.LogCalls(providers.Error)
	Expect(calls).To(HaveLen(1))
	Expect(calls[0].ContextFields).To(HaveKeyWithValue(FieldName, true))
}

func TestReportFalseDoesNotInjectField(t *testing.T) {
	RegisterTestingT(t)

	capturer := structured.LogProvider(nil)
	testProvider := LogProvider(capturer)
	ctx := context.Background()

	testProvider.Error(ctx, false, "test error")

	calls := capturer.LogCalls(providers.Error)
	Expect(calls).To(HaveLen(1))
	Expect(calls[0].ContextFields).NotTo(HaveKey(FieldName))
}

// Error/Warn are the reportable levels: report=true tags both. Info/Debug are
// covered separately (TestInfoReportDoesNotInjectField,
// TestDebugReportDoesNotInjectField) since report=true is deliberately a
// no-op for them.
func TestErrorAndWarnLevelsWorkWithReportTrue(t *testing.T) {
	RegisterTestingT(t)

	capturer := structured.LogProvider(nil)
	testProvider := LogProvider(capturer)
	ctx := context.Background()

	testProvider.Error(ctx, true, "error msg")
	testProvider.Warn(ctx, true, "warn msg")

	errorCalls := capturer.LogCalls(providers.Error)
	Expect(errorCalls).To(HaveLen(1))
	Expect(errorCalls[0].ContextFields).To(HaveKeyWithValue(FieldName, true))

	warnCalls := capturer.LogCalls(providers.Warn)
	Expect(warnCalls).To(HaveLen(1))
	Expect(warnCalls[0].ContextFields).To(HaveKeyWithValue(FieldName, true))
}

func TestPreservesExistingFields(t *testing.T) {
	RegisterTestingT(t)

	capturer := structured.LogProvider(nil)
	testProvider := LogProvider(capturer)
	ctx := log.ContextWithFields(context.Background(), log.Fields{
		"userId":   "123",
		"requestId": "abc-def",
	})

	testProvider.Error(ctx, true, "test error")

	calls := capturer.LogCalls(providers.Error)
	Expect(calls).To(HaveLen(1))
	Expect(calls[0].ContextFields).To(HaveKeyWithValue("userId", "123"))
	Expect(calls[0].ContextFields).To(HaveKeyWithValue("requestId", "abc-def"))
	Expect(calls[0].ContextFields).To(HaveKeyWithValue(FieldName, true))
}

func TestRecordDoesNotInjectField(t *testing.T) {
	RegisterTestingT(t)

	capturer := structured.LogProvider(nil)
	testProvider := LogProvider(capturer)
	ctx := context.Background()

	testProvider.Record(ctx, map[string]interface{}{"metric": 42})

	recordCalls := capturer.RecordCalls()
	Expect(recordCalls).To(HaveLen(1))
	Expect(recordCalls[0].ContextFields).NotTo(HaveKey(FieldName))
}

func TestFieldNameConstantIsExported(t *testing.T) {
	RegisterTestingT(t)

	Expect(FieldName).To(Equal("reportableError"))
}

// Info/Debug must NOT inject reportableError, even on report=true
// (InfoReport/DebugReport) — mirrors datadog_errors' Info/Debug exemption so
// the two sibling providers stay consistent about what "reported" means.
func TestInfoReportDoesNotInjectField(t *testing.T) {
	RegisterTestingT(t)

	capturer := structured.LogProvider(nil)
	testProvider := LogProvider(capturer)
	ctx := context.Background()

	testProvider.Info(ctx, true, "info msg")

	calls := capturer.LogCalls(providers.Info)
	Expect(calls).To(HaveLen(1))
	Expect(calls[0].ContextFields).NotTo(HaveKey(FieldName))
}

func TestDebugReportDoesNotInjectField(t *testing.T) {
	RegisterTestingT(t)

	capturer := structured.LogProvider(nil)
	testProvider := LogProvider(capturer)
	ctx := context.Background()

	testProvider.Debug(ctx, true, "debug msg")

	calls := capturer.LogCalls(providers.Debug)
	Expect(calls).To(HaveLen(1))
	Expect(calls[0].ContextFields).NotTo(HaveKey(FieldName))
}
