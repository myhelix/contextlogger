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
