package log_test

import (
	"context"
	"testing"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/merry"
	"github.com/myhelix/contextlogger/providers/reported_at"
	"github.com/myhelix/contextlogger/providers/structured"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "log Acceptance")
}

var _ = Describe("ReportableError Field Injection", func() {
	var (
		capturer *structured.StructuredOutputLogProvider
		ctx      context.Context
	)

	BeforeEach(func() {
		capturer = structured.LogProvider(nil)
		log.SetDefaultProvider(capturer)
		ctx = context.Background()
	})

	Describe("Positive Tests (Report methods SHOULD have the field)", func() {
		It("TestErrorReport_InjectsReportableErrorField", func() {
			log.FromContext(ctx).ErrorReport("test error")
			calls := capturer.LogCalls(providers.Error)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestWarnReport_InjectsReportableErrorField", func() {
			log.FromContext(ctx).WarnReport("test warn")
			calls := capturer.LogCalls(providers.Warn)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestInfoReport_InjectsReportableErrorField", func() {
			log.FromContext(ctx).InfoReport("test info")
			calls := capturer.LogCalls(providers.Info)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestDebugReport_InjectsReportableErrorField", func() {
			log.FromContext(ctx).DebugReport("test debug")
			calls := capturer.LogCalls(providers.Debug)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestPackageLevelErrorReport_InjectsReportableErrorField", func() {
			log.ErrorReport("test error")
			calls := capturer.LogCalls(providers.Error)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestPackageLevelWarnReport_InjectsReportableErrorField", func() {
			log.WarnReport("test warn")
			calls := capturer.LogCalls(providers.Warn)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestPackageLevelInfoReport_InjectsReportableErrorField", func() {
			log.InfoReport("test info")
			calls := capturer.LogCalls(providers.Info)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestPackageLevelDebugReport_InjectsReportableErrorField", func() {
			log.DebugReport("test debug")
			calls := capturer.LogCalls(providers.Debug)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})
	})

	Describe("Negative Tests (Non-report methods should NOT have the field)", func() {
		It("TestError_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).Error("test error")
			calls := capturer.LogCalls(providers.Error)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].Report).To(BeFalse())
		})

		It("TestWarn_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).Warn("test warn")
			calls := capturer.LogCalls(providers.Warn)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].Report).To(BeFalse())
		})

		It("TestInfo_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).Info("test info")
			calls := capturer.LogCalls(providers.Info)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].Report).To(BeFalse())
		})

		It("TestDebug_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).Debug("test debug")
			calls := capturer.LogCalls(providers.Debug)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].Report).To(BeFalse())
		})
	})

	Describe("Integration Test", func() {
		It("TestReportableError_ThroughFullProviderChain", func() {
			// Chain: merry -> reported_at -> structured
			capturer = structured.LogProvider(nil)
			reportedAt := reported_at.LogProvider(capturer, reported_at.RecommendedConfig)
			merryProvider := merry.LogProvider(reportedAt)
			log.SetDefaultProvider(merryProvider)

			log.FromContext(ctx).ErrorReport("test error")

			calls := capturer.LogCalls(providers.Error)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].Report).To(BeTrue())
		})
	})

	Describe("Extensibility Test", func() {
		It("TestReportFields_ExtensibilityDesign", func() {
			// Verify that if we add fields to the context, they are preserved alongside reportableError
			ctxWithFields := log.ContextWithFields(ctx, log.Fields{"existingField": "existingValue"})
			log.FromContext(ctxWithFields).ErrorReport("test error")

			calls := capturer.LogCalls(providers.Error)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("existingField", "existingValue"))
		})
	})
})
