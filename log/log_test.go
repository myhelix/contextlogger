package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	"github.com/myhelix/contextlogger/providers/logrus"
	"github.com/myhelix/contextlogger/providers/merry"
	"github.com/myhelix/contextlogger/providers/reported_at"
	"github.com/myhelix/contextlogger/providers/structured"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	logrusLib "github.com/sirupsen/logrus"
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

	Describe("Positive Tests (Error/Warn Report methods SHOULD have the field)", func() {
		It("TestErrorReport_InjectsReportableErrorField", func() {
			log.FromContext(ctx).ErrorReport("test error")
			calls := capturer.LogCalls(providers.Error)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportedLevel", "error"))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestWarnReport_InjectsReportableErrorField", func() {
			log.FromContext(ctx).WarnReport("test warn")
			calls := capturer.LogCalls(providers.Warn)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportableError", true))
			Expect(calls[0].ContextFields).To(HaveKeyWithValue("reportedLevel", "warning"))
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
	})

	Describe("Negative Tests (Info/Debug Report methods and non-report methods should NOT have the field)", func() {
		It("TestError_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).Error("test error")
			calls := capturer.LogCalls(providers.Error)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportedLevel"))
			Expect(calls[0].Report).To(BeFalse())
		})

		It("TestWarn_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).Warn("test warn")
			calls := capturer.LogCalls(providers.Warn)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportedLevel"))
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

		// InfoReport/DebugReport are *Report methods (report=true is passed
		// through to the provider), but reportableError must NOT be injected
		// into context for them — only Error/Warn reports are reportable.
		// This exercises the real log.go entry point (contextLogger method
		// form), not an inline-built provider chain, so it actually catches
		// the class of bug where log.go injects the field before any
		// provider runs.
		It("TestInfoReport_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).InfoReport("test info")
			calls := capturer.LogCalls(providers.Info)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportedLevel"))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestDebugReport_DoesNotInjectReportableErrorField", func() {
			log.FromContext(ctx).DebugReport("test debug")
			calls := capturer.LogCalls(providers.Debug)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].Report).To(BeTrue())
		})

		// Package-level form: exercises log.go's InfoReport()/DebugReport()
		// free functions, which delegate to BackgroundContext(), the other
		// real entry point named in plan-amendment-02.md.
		It("TestPackageLevelInfoReport_DoesNotInjectReportableErrorField", func() {
			log.InfoReport("test info")
			calls := capturer.LogCalls(providers.Info)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].Report).To(BeTrue())
		})

		It("TestPackageLevelDebugReport_DoesNotInjectReportableErrorField", func() {
			log.DebugReport("test debug")
			calls := capturer.LogCalls(providers.Debug)
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].ContextFields).NotTo(HaveKey("reportableError"))
			Expect(calls[0].Report).To(BeTrue())
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

	Describe("End-to-End Test with Logrus", func() {
		var (
			buf            *bytes.Buffer
			logrusProvider providers.LogProvider
		)

		BeforeEach(func() {
			buf = new(bytes.Buffer)
			var err error
			logrusProvider, err = logrus.LogProvider(nil, logrus.Config{
				Output: buf,
				Level:  "debug",
				Formatter: &logrusLib.JSONFormatter{
					DisableTimestamp: true,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			log.SetDefaultProvider(logrusProvider)
		})

		It("TestErrorReport_WithLogrus_OutputContainsReportableError", func() {
			log.ErrorReport("test error with logrus")

			var output map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(HaveKeyWithValue("reportableError", true))
			Expect(output).To(HaveKeyWithValue("reportedLevel", "error"))
			Expect(output).To(HaveKeyWithValue("level", "error"))
			Expect(output).To(HaveKeyWithValue("msg", "test error with logrus"))
		})

		It("TestError_WithLogrus_OutputDoesNotContainReportableError", func() {
			log.Error("test error with logrus")

			var output map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(HaveKey("reportableError"))
			Expect(output).To(HaveKeyWithValue("level", "error"))
			Expect(output).To(HaveKeyWithValue("msg", "test error with logrus"))
		})
	})
})
