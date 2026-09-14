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
	testProvider = LogProvider(outputProvider)
}

// A report of an error should get error.kind / error.message / error.stack.
func TestReportInjectsErrorFields(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, errors.New("it broke"))
	out := output.String()
	Expect(out).To(MatchRegexp(`error\.kind=`))
	Expect(out).To(MatchRegexp(`error\.message="it broke"`))
	Expect(out).To(MatchRegexp(`error\.stack=`))
}

// A non-report (report=false) must NOT get any error.* fields, so ordinary
// Error() calls don't become Datadog Error Tracking issues.
func TestNonReportDoesNotInject(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), false, errors.New("it broke"))
	out := output.String()
	Expect(out).NotTo(ContainSubstring("error.kind"))
	Expect(out).NotTo(ContainSubstring("error.message"))
	Expect(out).NotTo(ContainSubstring("error.stack"))
}

// error.kind should reflect the Go type of the error value.
func TestErrorKindFromType(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, errors.New("boom"))
	// stdlib errors.New yields *errors.errorString
	Expect(output.String()).To(MatchRegexp(`error\.kind=errors\.errorString`))
}

// Merry errors are supported (message + stack extracted).
func TestMerryError(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, merry.New("merry broke"))
	out := output.String()
	Expect(out).To(MatchRegexp(`error\.message="merry broke"`))
	Expect(out).To(MatchRegexp(`error\.stack=`))
}

// Non-error args are ignored (no error.* fields, no panic).
func TestNonErrorArgIgnored(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, "just a string")
	Expect(output.String()).NotTo(ContainSubstring("error.kind"))
}

func TestWarnReportInjectsAndUsesErrorSeverity(t *testing.T) {
	setup(t)

	testProvider.Warn(context.Background(), true, errors.New("warn broke"))
	out := output.String()
	Expect(out).To(MatchRegexp(`level=error`))
	Expect(out).To(MatchRegexp(`error\.message="warn broke"`))
}

func TestNonReportWarnUsesWarningSeverity(t *testing.T) {
	setup(t)

	testProvider.Warn(context.Background(), false, errors.New("warn broke"))
	Expect(output.String()).To(MatchRegexp(`level=warning`))
}

// An error passed alongside extra args (the common ErrorReport(err, "context")
// shape) must still be enriched — we scan all args, not just a lone one.
func TestReportInjectsWithExtraArgs(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, errors.New("it broke"), "extra context", 42)
	out := output.String()
	Expect(out).To(MatchRegexp(`error\.kind=`))
	Expect(out).To(MatchRegexp(`error\.message="it broke"`))
	Expect(out).To(MatchRegexp(`error\.stack=`))
}

// A typed-nil error (nil pointer stored in an error interface) satisfies the
// error type assertion but would panic on err.Error(); it must be treated as
// absent, not crash the report path.
type typedNilErr struct{}

func (*typedNilErr) Error() string { return "should never be called" }

func TestTypedNilErrorDoesNotPanic(t *testing.T) {
	setup(t)

	var e *typedNilErr // nil pointer, non-nil error interface
	Expect(func() {
		testProvider.Error(context.Background(), true, error(e))
	}).NotTo(Panic())
	Expect(output.String()).NotTo(ContainSubstring("error.kind"))
}

// When several args are errors, the first non-nil error wins.
func TestFirstNonNilErrorWins(t *testing.T) {
	setup(t)

	testProvider.Error(context.Background(), true, "leading string", errors.New("the real error"), errors.New("second"))
	Expect(output.String()).To(MatchRegexp(`error\.message="the real error"`))
}

// Info/Debug must NOT inject error.* fields even on report=true — this
// provider's contract is ErrorReport/WarnReport only.
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
