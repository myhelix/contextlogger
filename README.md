# ContextLogger
This project provides a pluggable, context-based logging, error reporting, and performance metrics interface for golang.

The "log" package is largely compatible with [Logrus](https://github.com/sirupsen/logrus), and Logrus can be used for output.

## Included Log Providers

The following packages are provided:

- **logrus**: Log output using the [Logrus](https://github.com/sirupsen/logrus) logger
- **rollbar**: Error reporting via [Rollbar](https://rollbar.com)
- **newrelic**: Performance and custom metrics via [NewRelic](https://newrelic.com)
- **merry**: Log structured error data and tracebacks to where an error was actually generated, using [Merry](https://github.com/ansel1/merry) errors
- **reported_at**: Include the file and line number responsible for each log message

Log providers are chained together in whatever combination you desire. New log providers can be easily implemented by following the simple LogProvider interface.

## Logging Interface

The main interface you interact with in using ContextLogger is log.ContextLogger, which includes standard library context.Context as well as the following logging methods:

```go
ErrorReport(    args ...interface{})
Error(          args ...interface{})
WarnReport(     args ...interface{})
Warn(           args ...interface{})
InfoReport(     args ...interface{})
Info(           args ...interface{})
DebugReport(    args ...interface{})
Debug(          args ...interface{})
```

The distinction between, e.g., "Error" and "ErrorReport" is up to you to define in your environment. At Helix, we use it to distinguish between "this broke, and a human needs to look at it" (ErrorReport) and "this broke, but just make a note of it, don't wake anyone up" (Error). Having does-someone-get-notified be an explicit dimension independent from severity has worked out well for managing our on-call quality of life, but YMMV; if you don't like the *Report methods, just ignore them.

## Adding Log Fields

ContextLogger allows you to build up metadata on a Context as it passed through your program, all of which will be output alongside any log message produced with that Context. This data survives even when the ContextLogger is passed as a standard context.Context. The interface for adding fields is Logrus-compatible:

```go
func A(ctx log.ContextLogger) {
    ctx = ctx.WithFields(log.Fields{
        "A": 1,
    })
    B(ctx)
}
func B(ctx log.ContextLogger) {
    C(ctx.WithField("B", 2))
}
// Accepts standard context.Context
func C(ctx context.Context) {
    // Recover log.ContextLogger from context.Context
    // Will generate default ContextLogger if ctx never had one before
    ctxLog := log.FromContext(ctx)
    ctxLog.Info("Made it to C")
    // Output: msg="Made it to C" A=1 B=2
}
```

## Metrics

ContextLogger also provides two methods for logging metrics:

```go
Record(metrics Metrics)
RecordEvent(eventName string, metrics Metrics)
```

Metrics is just another name for map[string]interface{}, same as log.Fields; and you might wonder what the difference between a metric with an event name and a log field with a log message is -- similar to ErrorReport vs. Error, this is really to provide a way to selectively send information to a different destination. The NewRelic log provider will take data from Record and add it to a newrelic.Transaction in the Context, and will put data from RecordEvent into a NewRelic Custom Event. But you could easily write a provider to send these anywhere you want to track some sort of metrics.

## Setting up the Default Provider

Here's an example config which chains together all the built-in providers (except for dummy, which is just for startup and testing):

```go
// This assumes there's a "config" struct in this package where certain project config data is coming from.

import (
	"github.com/myhelix/contextlogger/log"
	cl_logrus "github.com/myhelix/contextlogger/providers/logrus"
	cl_merry "github.com/myhelix/contextlogger/providers/merry"
	cl_newrelic "github.com/myhelix/contextlogger/providers/newrelic"
	"github.com/myhelix/contextlogger/providers/reported_at"
	cl_rollbar "github.com/myhelix/contextlogger/providers/rollbar"
	"github.com/myhelix/rollbar"
)

func configureLogging() error {
	// Keep track for reporting at the end
	var rollbarEnabled, newRelicEnabled bool

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	// Logrus provides the base log output
	logProvider, err := cl_logrus.LogProvider(nil, cl_logrus.Config{
		Output:    os.Stderr,
		Level:     config.LogLevel,
		Formatter: cl_logrus.RecommendedFormatter,
	})
	if err != nil {
		return err
	}

	// Rollbar error reporting
	if config.RollbarToken != "" {
		codeRevBytes, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {
			return err
		}
		codeRev := strings.Trim(string(codeRevBytes), " \n")

		rollbar.Token = config.RollbarToken
		rollbar.Environment = config.Env
		rollbar.CodeVersion = codeRev  // Git hash/branch/tag (required for GitHub integration)
		rollbar.ServerRoot = config.Package // path of project (required for GitHub integration and non-project stacktrace collapsing)
		rollbar.FilterFields = regexp.MustCompile("(?i)password|secret|token|auth")

		// Rollbar config is all at package level, so no config to pass in here
		logProvider, err = cl_rollbar.LogProvider(logProvider)
		if err != nil {
			return err
		}
		rollbarEnabled = true
	}

	// NewRelicApp is a newrelic.Application
	if config.NewRelicApp != nil {
		logProvider, err = cl_newrelic.LogProvider(logProvider, config.NewRelicApp)
		if err != nil {
			return err
		}
		newRelicEnabled = true
	}

	// Parse merry.Error values into log fields; log merry tracebacks with Error/ErrorReport
	logProvider, err = cl_merry.LogProvider(logProvider)
	if err != nil {
		return err
	}

	// Note the file and line number where each log message was reported from
	logProvider, err = reported_at.LogProvider(logProvider, reported_at.RecommendedConfig)
	if err != nil {
		return err
	}

	log.SetDefaultProvider(logProvider)

	log.WithFields(log.Fields{
		"level":           config.LogLevel,
		"rollbarEnabled":  rollbarEnabled,
		"newRelicEnabled": newRelicEnabled,
	}).Info("Configured logging")

	return nil
}
```

