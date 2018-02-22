// © 2017 Helix OpCo LLC. All rights reserved.
// Initial Author: jpecknerhelix

package structured

import (
	"os"

	"github.com/myhelix/contextlogger/providers/chaining"
	"github.com/myhelix/contextlogger/providers/dummy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
)

var _ = Describe("bufferedLogProvider", func() {

	var (
		fields = log.Fields{
			"myFieldOne":   "myValueOne",
			"myFieldTwo":   "myValueTwo",
			"myFieldThree": "myValueThree",
		}

		provider      *StructuredOutputLogProvider
		contextLogger log.ContextLogger
	)

	var verifyEmptyLogCalls = func(callTypes []providers.RawLogCallType) {
		for _, callType := range callTypes {
			Ω(provider.GetLogCallsByCallType(callType)).Should(BeEmpty())
		}
	}

	BeforeEach(func() {
		provider = NewStructuredOutputLogProvider()
		contextLogger = log.WithFields(fields)
	})

	It("Should log multiple Error calls", func() {
		// Call log method under test
		provider.Error(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Error(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		errorCalls := provider.GetLogCallsByCallType(providers.Error)
		Ω(errorCalls).Should(Equal([]LogCallArgs{
			{
				ContextFields: fields,
				Report:        true,
				Args:          []interface{}{"Message 1", "Some additional details about message 1"},
				CallType:      providers.Error,
			}, {
				ContextFields: fields,
				Report:        false,
				Args:          []interface{}{"Message 2", "Some additional details about message 2"},
				CallType:      providers.Error,
			},
		}))

		verifyEmptyLogCalls([]providers.RawLogCallType{providers.Info, providers.Warn, providers.Debug})
	})

	It("Should log multiple Warn calls", func() {
		// Call log method under test
		provider.Warn(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Warn(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		warnCalls := provider.GetLogCallsByCallType(providers.Warn)
		Ω(warnCalls).Should(Equal([]LogCallArgs{
			{
				ContextFields: fields,
				Report:        true,
				Args:          []interface{}{"Message 1", "Some additional details about message 1"},
				CallType:      providers.Warn,
			}, {
				ContextFields: fields,
				Report:        false,
				Args:          []interface{}{"Message 2", "Some additional details about message 2"},
				CallType:      providers.Warn,
			},
		}))

		verifyEmptyLogCalls([]providers.RawLogCallType{providers.Info, providers.Error, providers.Debug})
	})

	It("Should log multiple Info calls", func() {
		// Call log method under test
		provider.Info(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Info(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		infoCalls := provider.GetLogCallsByCallType(providers.Info)
		Ω(infoCalls).Should(Equal([]LogCallArgs{
			{
				ContextFields: fields,
				Report:        true,
				Args:          []interface{}{"Message 1", "Some additional details about message 1"},
				CallType:      providers.Info,
			}, {
				ContextFields: fields,
				Report:        false,
				Args:          []interface{}{"Message 2", "Some additional details about message 2"},
				CallType:      providers.Info,
			},
		}))

		verifyEmptyLogCalls([]providers.RawLogCallType{providers.Error, providers.Warn, providers.Debug})
	})

	It("Should log multiple Debug calls", func() {
		// Call log method under test
		provider.Debug(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Debug(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		debugCalls := provider.GetLogCallsByCallType(providers.Debug)
		Ω(debugCalls).Should(Equal([]LogCallArgs{
			{
				ContextFields: fields,
				Report:        true,
				Args:          []interface{}{"Message 1", "Some additional details about message 1"},
				CallType:      providers.Debug,
			}, {
				ContextFields: fields,
				Report:        false,
				Args:          []interface{}{"Message 2", "Some additional details about message 2"},
				CallType:      providers.Debug,
			},
		}))

		verifyEmptyLogCalls([]providers.RawLogCallType{providers.Info, providers.Warn, providers.Error})
	})

	It("Should log multiple Record calls", func() {
		// Call log method under test
		provider.Record(contextLogger, log.Metrics{
			"Metric 1": "Value 1",
			"Metric 2": "Value 2",
		})
		provider.Record(contextLogger, log.Metrics{
			"Metric 3": "Value 3",
			"Metric 4": "Value 4",
		})

		// Verify
		recordCalls := provider.GetRecordCalls()
		Ω(len(recordCalls)).Should(Equal(2))
		Ω(recordCalls[0]).Should(Equal(RecordCallArgs{
			ContextFields: fields,
			Metrics: log.Metrics{
				"Metric 1": "Value 1",
				"Metric 2": "Value 2",
			},
		}))
		Ω(recordCalls[1]).Should(Equal(RecordCallArgs{
			ContextFields: fields,
			Metrics: log.Metrics{
				"Metric 3": "Value 3",
				"Metric 4": "Value 4",
			},
		}))
	})

	It("Should log multiple RecordEvent calls", func() {
		// Call log method under test
		provider.RecordEvent(contextLogger, "Event1", log.Metrics{
			"Metric 1": "Value 1",
			"Metric 2": "Value 2",
		})
		provider.RecordEvent(contextLogger, "Event2", log.Metrics{
			"Metric 3": "Value 3",
			"Metric 4": "Value 4",
		})

		// Verify
		recordEventCalls := provider.GetRecordEventCalls()
		Ω(len(recordEventCalls)).Should(Equal(2))
		Ω(recordEventCalls[0]).Should(Equal(RecordEventCallArgs{
			ContextFields: fields,
			EventName:     "Event1",
			Metrics: log.Metrics{
				"Metric 1": "Value 1",
				"Metric 2": "Value 2",
			},
		}))
		Ω(recordEventCalls[1]).Should(Equal(RecordEventCallArgs{
			ContextFields: fields,
			EventName:     "Event2",
			Metrics: log.Metrics{
				"Metric 3": "Value 3",
				"Metric 4": "Value 4",
			},
		}))
	})

	It("Should construct a chaining StructuredOutputLogProvider that handles a nil next provider", func() {
		lp := LogProvider(nil)
		Ω(lp.LogProvider).ShouldNot(BeNil())
	})

	It("Should construct a chaining log provider with a dummy provider as next provider", func() {
		dummyProvider := dummy.LogProvider(os.Stdout)
		lp := LogProvider(dummyProvider)
		Ω(lp.LogProvider).Should(Equal(chaining.LogProvider(dummyProvider)))
	})

	It("Should call wait on the next provider", func() {
		ws := new(dummy.WaitState)
		Ω(ws.Get()).Should(BeFalse())
		dummyProvider := dummy.LogProviderWithWaitState(os.Stdout, ws)
		lp := LogProvider(dummyProvider)
		Ω(lp.LogProvider).Should(Equal(chaining.LogProvider(dummyProvider)))

		// Calling wait should call wait on the next provider too
		lp.Wait()
		Ω(ws.Get()).Should(BeTrue())
	})

	It("Should log and then pass on the logs to the next provider", func() {
		var (
			msg1     = "Message 1"
			msg2     = "Message 2"
			details1 = "Details about message 1"
			details2 = "Details about message 2"
		)

		// lp is `this` provider, and the `provider` is the next provider.
		// At the end of this, they should have the same log calls, metrics and events
		lp := LogProvider(provider)

		// put some logs
		lp.Error(contextLogger, true, msg1, details1)
		lp.Error(contextLogger, false, msg2, details2)
		lp.Info(contextLogger, true, msg1, details1)
		lp.Info(contextLogger, false, msg2, details2)
		lp.Debug(contextLogger, true, msg1, details1)
		lp.Debug(contextLogger, false, msg2, details2)
		lp.Warn(contextLogger, true, msg1, details1)
		lp.Warn(contextLogger, false, msg2, details2)

		// put some metrics
		lp.Record(contextLogger, log.Metrics{
			"Metric 1": "Value 1",
			"Metric 2": "Value 2",
		})
		lp.Record(contextLogger, log.Metrics{
			"Metric 3": "Value 3",
			"Metric 4": "Value 4",
		})

		// put some events
		lp.RecordEvent(contextLogger, "Event1", log.Metrics{
			"Metric 1": "Value 1",
			"Metric 2": "Value 2",
		})
		lp.RecordEvent(contextLogger, "Event2", log.Metrics{
			"Metric 3": "Value 3",
			"Metric 4": "Value 4",
		})

		// Verify logs
		Ω(lp.GetLogCallsByCallType(providers.Error)).Should(Equal(provider.GetLogCallsByCallType(providers.Error)))
		Ω(lp.GetLogCallsByCallType(providers.Info)).Should(Equal(provider.GetLogCallsByCallType(providers.Info)))
		Ω(lp.GetLogCallsByCallType(providers.Debug)).Should(Equal(provider.GetLogCallsByCallType(providers.Debug)))
		Ω(lp.GetLogCallsByCallType(providers.Warn)).Should(Equal(provider.GetLogCallsByCallType(providers.Warn)))

		// Verify metrics
		Ω(lp.GetRecordCalls()).Should(Equal(provider.GetRecordCalls()))

		// Verify events
		Ω(lp.GetRecordEventCalls()).Should(Equal(provider.GetRecordEventCalls()))
	})
})
