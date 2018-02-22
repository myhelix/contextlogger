## 1.4.0 (2018-02-22)
Features:
- Thread safety for StructuredOutputLogProvider (@emre.colak)

Breaking changes:
- Renamed type RawLogCalls inside providers/structured to LogCalls. (@emre.colak)
- Removed GetRawLogCalls() method from StructuredOutputLogProvider. Clients should now use GetLogCallsByCallType() (@emre.colak)

## 1.3.0 (2018-02-14)
Features:
- Added chaining support to StructuredOutputLogProvider (@emre.colak)

## 1.2.1 (2018-01-05)
Fixes:
- Fixes StructuredOutputLogProvider so that it conforms to the LogProvider interface. (@jpecknerhelix)

## 1.2.0 (2017-12-20)
Features:
- Added StructuredOutputLogProvider so that the args passed to LogProvider methods can be verified. (@jpecknerhelix)

## 1.1.2 (2017-08-01)
Fixes:
- Correctly wire up Wait() for rollbar (@chriswhelix)

## 1.1.1 (2017-07-25)
Features:
- Added JSON formatter and test for logrus provider. (@mentat)

## 1.1.0 (2017-04-21)
Features:
- Added chaining provider to DRY up provider chaining and allow any provider to act as the base provider. (@chriswhelix)

## 1.0.0
Initial import from internal Helix hss project.
