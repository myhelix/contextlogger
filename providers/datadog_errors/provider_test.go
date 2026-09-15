// © 2016-2026 Helix OpCo LLC. All rights reserved.

package datadog_errors

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/ansel1/merry"
	"github.com/myhelix/contextlogger/providers"
	cl_logrus "github.com/myhelix/contextlogger/providers/logrus"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var output *bytes.Buffer
var testProvider providers.LogProvider

func setup(t *testing.T) {
	setupWithOptions(t, Options{})
}

func setupWithOptions(t *testing.T, options Options) {
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
	testProvider = LogProviderWithOptions(outputProvider, options)
}

func TestReportInjectsErrorFields(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, errors.New("it broke"))
	out := output.String()
	Expect(out).To(MatchRegexp(`error\.kind=`))
	Expect(out).To(MatchRegexp(`error\.message="it broke"`))
	Expect(out).To(MatchRegexp(`error\.stack=`))
}

func TestNonReportDoesNotInject(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), false, errors.New("it broke"))
	out := output.String()
	Expect(out).NotTo(ContainSubstring("error.kind"))
	Expect(out).NotTo(ContainSubstring("error.message"))
	Expect(out).NotTo(ContainSubstring("error.stack"))
}

func TestErrorKindFromType(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, errors.New("boom"))
	Expect(output.String()).To(MatchRegexp(`error\.kind=errors\.errorString`))
}

func TestMerryError(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, merry.New("merry broke"))
	out := output.String()
	Expect(out).To(MatchRegexp(`error\.message="merry broke"`))
	Expect(out).To(MatchRegexp(`error\.stack=`))
}

func TestStringOnlyErrorReportUsesSyntheticFields(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, "just a string")
	out := output.String()
	Expect(out).To(MatchRegexp(`error\.kind=ReportedError`))
	Expect(out).To(MatchRegexp(`error\.message="just a string"`))
}

func TestWarnReportInjectsAndPreservesWarningSeverityByDefault(t *testing.T) {
	setup(t)

	testProvider.Warn(context.Background(), true, errors.New("warn broke"))
	out := output.String()
	Expect(out).To(MatchRegexp(`level=warning`))
	Expect(out).To(MatchRegexp(`error\.message="warn broke"`))
}

func TestWarnReportCanPromoteErrorSeverity(t *testing.T) {
	setupWithOptions(t, Options{PromoteReportedWarnings: true})

	testProvider.Warn(context.Background(), true, errors.New("warn broke"))
	out := output.String()
	Expect(out).To(MatchRegexp(`level=error`))
	Expect(out).To(MatchRegexp(`error\.message="warn broke"`))
}

func TestStringOnlyWarnReportIsPromotedWithSyntheticFields(t *testing.T) {
	setupWithOptions(t, Options{PromoteReportedWarnings: true})

	testProvider.Warn(context.Background(), true, "warning without an error")
	out := output.String()
	Expect(out).To(MatchRegexp(`level=error`))
	Expect(out).To(MatchRegexp(`error\.kind=ReportedWarning`))
	Expect(out).To(MatchRegexp(`error\.message="warning without an error"`))
}

func TestStringOnlyWarnReportPreservesWarningSeverityByDefault(t *testing.T) {
	setup(t)

	testProvider.Warn(context.Background(), true, "warning without an error")
	out := output.String()
	Expect(out).To(MatchRegexp(`level=warning`))
	Expect(out).To(MatchRegexp(`error\.kind=ReportedWarning`))
}

func TestNonReportWarnUsesWarningSeverity(t *testing.T) {
	setup(t)

	testProvider.Warn(context.Background(), false, errors.New("warn broke"))
	Expect(output.String()).To(MatchRegexp(`level=warning`))
}

func TestReportInjectsWithExtraArgs(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, errors.New("it broke"), "extra context", 42)
	out := output.String()
	Expect(out).To(MatchRegexp(`error\.kind=`))
	Expect(out).To(MatchRegexp(`error\.message="it broke"`))
	Expect(out).To(MatchRegexp(`error\.stack=`))
}

// A typed-nil error must be ignored instead of panicking on Error().
type typedNilErr struct{}

func (*typedNilErr) Error() string { return "should never be called" }

func TestTypedNilErrorDoesNotPanic(t *testing.T) {
	setup(t)

	var e *typedNilErr
	Expect(func() {
		testProvider.Error(context.Background(), true, error(e))
	}).NotTo(Panic())
	Expect(output.String()).NotTo(ContainSubstring("error.kind"))
}

func TestFirstNonNilErrorWins(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, "leading string", errors.New("the real error"), errors.New("second"))
	Expect(output.String()).To(MatchRegexp(`error\.message="the real error"`))
}

func TestInfoReportDoesNotInject(t *testing.T) {
	setup(t)

	testProvider.Info(context.Background(), true, errors.New("info with error"))
	Expect(output.String()).NotTo(ContainSubstring("error.kind"))
	Expect(output.String()).NotTo(ContainSubstring("error.stack"))
}

func TestDebugReportDoesNotInject(t *testing.T) {
	setup(t)

	testProvider.Debug(context.Background(), true, errors.New("debug with error"))
	Expect(output.String()).NotTo(ContainSubstring("error.kind"))
	Expect(output.String()).NotTo(ContainSubstring("error.stack"))
}
