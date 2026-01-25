// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

package logrus

import (
	"bytes"
	"testing"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var output *bytes.Buffer
var testProvider providers.LogProvider

func setupJSON(t *testing.T) {
	RegisterTestingT(t)

	output = new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		Output: output,
		Level:  "debug",
		Formatter: &logrus.JSONFormatter{
			DisableTimestamp: true,
			TimestampFormat:  "sometime", // Omit timestamp to make output predictable
		},
		FlattenDataDogFields: false,
	})
	Expect(err).To(BeNil())
	testProvider = provider
}

func TestJSONLogs(t *testing.T) {
	setupJSON(t)

	log.SetDefaultProvider(testProvider)

	log.WithFields(log.Fields{
		"first_name": "Sam",
		"last_name":  "Gamgee",
	}).Info("Hi there.")
	Expect(output.String()).To(MatchRegexp(`{"first_name":"Sam","last_name":"Gamgee","level":"info","msg":"Hi there."}`))
}

func TestJSONLogsMultiline(t *testing.T) {
	setupJSON(t)

	log.SetDefaultProvider(testProvider)

	multiLine1 := `This
	is
	a
	multiline
	string`

	multiLine2 := `Another
	great
	multiline
	string`

	log.WithFields(log.Fields{
		"long_string": multiLine1,
	}).Error(multiLine2)

	expected := `{"level":"error","long_string":"This\n\tis\n\ta\n\tmultiline\n\tstring","msg":"Another\n\tgreat\n\tmultiline\n\tstring"}`

	Expect(output.String()).To(MatchJSON(expected))
}

func TestDataDogFieldFlattening(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		Output: output,
		Level:  "info",
		Formatter: &logrus.JSONFormatter{
			DisableTimestamp: true,
		},
		FlattenDataDogFields: true,
	})
	Expect(err).To(BeNil())

	log.SetDefaultProvider(provider)

	// Test Case 1: All DataDog fields present with additional fields
	log.WithFields(log.Fields{
		"dd.trace_id":       "1234567890",
		"dd.span_id":        "9876543210",
		"lambda.request_id": "abc-123-def",
		"user_id":           "12345",
		"other_field":       "value",
	}).Info("Processing request")

	expected := `{
		"level":"info",
		"msg":"Processing request",
		"dd.trace_id":"1234567890",
		"dd.span_id":"9876543210",
		"lambda.request_id":"abc-123-def",
		"context":{
			"user_id":"12345",
			"other_field":"value"
		}
	}`

	Expect(output.String()).To(MatchJSON(expected))
}

func TestDataDogFieldFlatteningPartial(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		Output: output,
		Level:  "info",
		Formatter: &logrus.JSONFormatter{
			DisableTimestamp: true,
		},
		FlattenDataDogFields: true,
	})
	Expect(err).To(BeNil())

	log.SetDefaultProvider(provider)

	// Test Case 2: Only some DataDog fields present
	log.WithFields(log.Fields{
		"dd.trace_id": "1234567890",
		"user_id":     "12345",
		"action":      "login",
	}).Info("User action")

	expected := `{
		"level":"info",
		"msg":"User action",
		"dd.trace_id":"1234567890",
		"context":{
			"user_id":"12345",
			"action":"login"
		}
	}`

	Expect(output.String()).To(MatchJSON(expected))
}

func TestDataDogFieldFlatteningNoDataDogFields(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		Output: output,
		Level:  "info",
		Formatter: &logrus.JSONFormatter{
			DisableTimestamp: true,
		},
		FlattenDataDogFields: true,
	})
	Expect(err).To(BeNil())

	log.SetDefaultProvider(provider)

	// Test Case 3: No DataDog fields, should keep original behavior
	log.WithFields(log.Fields{
		"user_id": "12345",
		"action":  "logout",
	}).Info("User logout")

	expected := `{
		"level":"info",
		"msg":"User logout",
		"action":"logout",
		"user_id":"12345"
	}`

	Expect(output.String()).To(MatchJSON(expected))
}

func TestDataDogFieldFlatteningOnlyDataDogFields(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		Output: output,
		Level:  "info",
		Formatter: &logrus.JSONFormatter{
			DisableTimestamp: true,
		},
		FlattenDataDogFields: true,
	})
	Expect(err).To(BeNil())

	log.SetDefaultProvider(provider)

	// Test Case 4: Only DataDog fields, no context object should be created
	log.WithFields(log.Fields{
		"dd.trace_id":       "1234567890",
		"dd.span_id":        "9876543210",
		"lambda.request_id": "req-789",
	}).Info("Traced request")

	expected := `{
		"level":"info",
		"msg":"Traced request",
		"dd.trace_id":"1234567890",
		"dd.span_id":"9876543210",
		"lambda.request_id":"req-789"
	}`

	Expect(output.String()).To(MatchJSON(expected))
}

func TestBackwardCompatibilityWithoutFlattening(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		Output: output,
		Level:  "info",
		Formatter: &logrus.JSONFormatter{
			DisableTimestamp: true,
		},
		FlattenDataDogFields: false, // Explicitly disabled
	})
	Expect(err).To(BeNil())

	log.SetDefaultProvider(provider)

	// Test Case 5: Backward compatibility - DD fields should remain in flat structure
	log.WithFields(log.Fields{
		"dd.trace_id":       "1234567890",
		"dd.span_id":        "9876543210",
		"lambda.request_id": "req-789",
		"user_id":           "12345",
	}).Info("Legacy format")

	expected := `{
		"level":"info",
		"msg":"Legacy format",
		"dd.trace_id":"1234567890",
		"dd.span_id":"9876543210",
		"lambda.request_id":"req-789",
		"user_id":"12345"
	}`

	Expect(output.String()).To(MatchJSON(expected))
}

func TestDataDogFieldFlatteningWithNumericTraceIds(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		Output: output,
		Level:  "info",
		Formatter: &logrus.JSONFormatter{
			DisableTimestamp: true,
		},
		FlattenDataDogFields: true,
	})
	Expect(err).To(BeNil())

	log.SetDefaultProvider(provider)

	// Test Case 6: Numeric trace IDs (DataDog can use either string or numeric)
	log.WithFields(log.Fields{
		"dd.trace_id": 1234567890,
		"dd.span_id":  9876543210,
		"user_id":     "12345",
	}).Info("Numeric trace IDs")

	expected := `{
		"level":"info",
		"msg":"Numeric trace IDs",
		"dd.trace_id":1234567890,
		"dd.span_id":9876543210,
		"context":{
			"user_id":"12345"
		}
	}`

	Expect(output.String()).To(MatchJSON(expected))
}
