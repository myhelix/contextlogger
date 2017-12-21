// © 2017 Helix OpCo LLC. All rights reserved.
// Initial Author: jpecknerhelix

package structured

import (
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

	BeforeEach(func() {
		provider = NewStructuredOutputLogProvider()
		contextLogger = log.WithFields(fields)
	})

	It("Should log multiple Error calls", func() {
		// Call log method under test
		provider.Error(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Error(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		logCalls := provider.GetRawLogCalls()[providers.Error]
		Ω(len(logCalls)).Should(Equal(2))
		Ω(logCalls[0]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        true,
			Args:          []interface{}{"Message 1", "Some additional details about message 1"},
		}))
		Ω(logCalls[1]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        false,
			Args:          []interface{}{"Message 2", "Some additional details about message 2"},
		}))
	})

	It("Should log multiple Warn calls", func() {
		// Call log method under test
		provider.Warn(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Warn(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		logCalls := provider.GetRawLogCalls()[providers.Warn]
		Ω(len(logCalls)).Should(Equal(2))
		Ω(logCalls[0]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        true,
			Args:          []interface{}{"Message 1", "Some additional details about message 1"},
		}))
		Ω(logCalls[1]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        false,
			Args:          []interface{}{"Message 2", "Some additional details about message 2"},
		}))
	})

	It("Should log multiple Info calls", func() {
		// Call log method under test
		provider.Info(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Info(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		logCalls := provider.GetRawLogCalls()[providers.Info]
		Ω(len(logCalls)).Should(Equal(2))
		Ω(logCalls[0]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        true,
			Args:          []interface{}{"Message 1", "Some additional details about message 1"},
		}))
		Ω(logCalls[1]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        false,
			Args:          []interface{}{"Message 2", "Some additional details about message 2"},
		}))
	})

	It("Should log multiple Debug calls", func() {
		// Call log method under test
		provider.Debug(contextLogger, true, "Message 1", "Some additional details about message 1")
		provider.Debug(contextLogger, false, "Message 2", "Some additional details about message 2")

		// Verify
		logCalls := provider.GetRawLogCalls()[providers.Debug]
		Ω(len(logCalls)).Should(Equal(2))
		Ω(logCalls[0]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        true,
			Args:          []interface{}{"Message 1", "Some additional details about message 1"},
		}))
		Ω(logCalls[1]).Should(Equal(RawLogCallArgs{
			ContextFields: fields,
			Report:        false,
			Args:          []interface{}{"Message 2", "Some additional details about message 2"},
		}))
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

})
