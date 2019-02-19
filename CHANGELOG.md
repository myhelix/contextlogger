## 1.5.2 (2019-02-19)
- .gitignore changes for GoCD builds

## 1.5.1 (2019-02-19)
- Remove build config. Dev builds will use working copy version, gocd uses 4.9.2 based on pipeline template.

## 1.5.0 (2019-02-11)
- Fix dependency versions, no more range versions

## 1.4.1 (2018-03-05)
Fixes:
- Update the casing of "Sirupsen" to "sirupsen" in imports to logrus (https://github.com/sirupsen/logrus/issues/570)

## 1.4.0 (2018-02-22)
Features:
- Thread safety for StructuredOutputLogProvider (@emre.colak)

Breaking changes:
- Renamed type RawLogCallType inside providers to LogLevel. (@emre.colak)
- Renamed type RawLogCalls inside providers/structured to LogCalls.
- Renamed method GetRecordCallArgs to RecordCalls in StructuredOutputLogProvider. This method also returns a slice of pointers instead of values now.
- Removed type RecordEventCallArgs from providers/structured.
- Removed method GetRecordEventCalls from StructuredOutputLogProvider.
- Removed method GetRawLogCalls from StructuredOutputLogProvider. Clients should now use LogCalls by passing appropriate log levels. This method also returns a slice of pointers instead of values now.
- Removed the deprecated constructor NewStructuredOutputLogProvider from providers/structured

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
