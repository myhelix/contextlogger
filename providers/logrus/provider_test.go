// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

package logrus

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
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
		output,
		"debug",
		&logrus.JSONFormatter{
			DisableTimestamp: true,
			TimestampFormat:  "sometime", // Omit timestamp to make output predictable
		},
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

func setupProviderWithJSONOutput(t *testing.T) (*bytes.Buffer, providers.LogProvider) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	provider, err := LogProvider(nil, Config{
		output,
		"info",
		&logrus.JSONFormatter{
			DisableTimestamp: true,
		},
	})
	Expect(err).To(BeNil())
	return output, provider
}

func startTestTracer() {
	tracer.Start(
		tracer.WithService("test-service"),
		tracer.WithEnv("test-env"),
		tracer.WithServiceVersion("v1.0.0"),
	)
}

func parseJSONLog(output *bytes.Buffer) map[string]interface{} {
	var logEntry map[string]interface{}
	err := json.Unmarshal(output.Bytes(), &logEntry)
	Expect(err).To(BeNil())
	return logEntry
}

func TestDatadogTraceCorrelation_WithSpan(t *testing.T) {
	output, provider := setupProviderWithJSONOutput(t)

	startTestTracer()
	defer tracer.Stop()

	span, ctx := tracer.StartSpanFromContext(context.Background(), "test.operation")
	defer span.Finish()

	provider.Info(ctx, false, "test message")

	logEntry := parseJSONLog(output)

	Expect(logEntry).To(HaveKey("dd.trace_id"))
	Expect(logEntry).To(HaveKey("dd.span_id"))
	Expect(logEntry["dd.trace_id"]).ToNot(BeEmpty())
	Expect(logEntry["dd.span_id"]).ToNot(BeEmpty())
	Expect(logEntry["msg"]).To(Equal("test message"))
	Expect(logEntry["level"]).To(Equal("info"))
}

func TestDatadogTraceCorrelation_WithoutSpan(t *testing.T) {
	output, provider := setupProviderWithJSONOutput(t)

	provider.Info(context.Background(), false, "test message without span")

	logEntry := parseJSONLog(output)

	Expect(logEntry).ToNot(HaveKey("dd.trace_id"))
	Expect(logEntry).ToNot(HaveKey("dd.span_id"))
	Expect(logEntry).ToNot(HaveKey("dd.service"))
	Expect(logEntry).ToNot(HaveKey("dd.env"))
	Expect(logEntry).ToNot(HaveKey("dd.version"))
	Expect(logEntry["msg"]).To(Equal("test message without span"))
	Expect(logEntry["level"]).To(Equal("info"))
}
