// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

package reported_at

import (
	"bytes"
	"context"
	"testing"

	"github.com/calm/contextlogger/v2/providers"
	cl_logrus "github.com/calm/contextlogger/v2/providers/logrus"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var output *bytes.Buffer
var testProvider providers.LogProvider

func setup(t *testing.T, config Config) {
	RegisterTestingT(t)

	output = new(bytes.Buffer)
	outputProvider, err := cl_logrus.LogProvider(nil, cl_logrus.Config{
		output,
		"debug",
		&logrus.TextFormatter{
			DisableColors:   true,
			TimestampFormat: "sometime", // Omit timestamp to make output predictable
		},
	})
	Expect(err).To(BeNil())
	testProvider = LogProvider(outputProvider, config)
}

func TestReportedAt(t *testing.T) {
	setup(t, Config{})

	testProvider.Info(context.Background(), false, "foo")
	Expect(output.String()).To(MatchRegexp(`time=sometime level=info msg=foo reportedAt=".*/reported_at/provider_test.go:\d+`))
}

func TestReportedAtFiltering(t *testing.T) {
	setup(t, RecommendedConfig)

	testProvider.Info(context.Background(), false, "foo")
	Expect(output.String()).To(MatchRegexp(`time=sometime level=info msg=foo reportedAt=".*src/testing/testing.go:\d+`))
}
