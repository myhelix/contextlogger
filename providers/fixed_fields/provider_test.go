// © 2026 Helix OpCo LLC. All rights reserved.

package fixed_fields

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/myhelix/contextlogger/log"
	cl_logrus "github.com/myhelix/contextlogger/providers/logrus"
	. "github.com/onsi/gomega"
)

func TestConfiguredFieldsOverrideContextFields(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	outputProvider, err := cl_logrus.LogProvider(nil, cl_logrus.Config{
		Output:    output,
		Level:     "debug",
		Formatter: cl_logrus.JSONFormatter,
	})
	Expect(err).To(BeNil())

	provider := LogProvider(outputProvider, log.Fields{"service": "fulfillment"})
	ctx := log.ContextWithFields(context.Background(), log.Fields{"service": "calling-service"})
	provider.Info(ctx, false, "test message")

	var entry map[string]interface{}
	Expect(json.Unmarshal(output.Bytes(), &entry)).To(Succeed())
	Expect(entry["service"]).To(Equal("fulfillment"))
}

func TestConfiguredFieldsOverrideMetricFields(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	outputProvider, err := cl_logrus.LogProvider(nil, cl_logrus.Config{
		Output:    output,
		Level:     "debug",
		Formatter: cl_logrus.JSONFormatter,
	})
	Expect(err).To(BeNil())

	provider := LogProvider(outputProvider, log.Fields{"service": "fulfillment"})
	provider.Record(context.Background(), map[string]interface{}{"service": "calling-service"})

	var entry map[string]interface{}
	Expect(json.Unmarshal(output.Bytes(), &entry)).To(Succeed())
	Expect(entry["service"]).To(Equal("fulfillment"))
}

func TestConfiguredFieldsAreCopied(t *testing.T) {
	RegisterTestingT(t)

	output := new(bytes.Buffer)
	outputProvider, err := cl_logrus.LogProvider(nil, cl_logrus.Config{
		Output:    output,
		Level:     "debug",
		Formatter: cl_logrus.JSONFormatter,
	})
	Expect(err).To(BeNil())

	fields := log.Fields{"service": "fulfillment"}
	provider := LogProvider(outputProvider, fields)
	fields["service"] = "mutated-after-configuration"
	provider.Info(context.Background(), false, "test message")

	var entry map[string]interface{}
	Expect(json.Unmarshal(output.Bytes(), &entry)).To(Succeed())
	Expect(entry["service"]).To(Equal("fulfillment"))
}
