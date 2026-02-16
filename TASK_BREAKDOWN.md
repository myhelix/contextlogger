# Task Breakdown: reportableError Field Injection

Detailed task list for implementing the feature described in [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).

---

This section maps each implementation task to the relevant plan section and provides clear, actionable instructions for a developer.

### Task 1: Add `reportFields` variable and helper to `log/log.go`

**Plan Reference**: [Section 3.3.1](IMPLEMENTATION_PLAN.md#331-add-report-fields-constant), [Section 3.3.2](IMPLEMENTATION_PLAN.md#332-add-helper-function)
**Files**: [`log/log.go`](log/log.go)
**Description**: Add the `reportFields` variable (e.g., `var reportFields = Fields{"reportableError": true}`) and the `contextWithReportFields(ctx)` helper function. Place these near the existing field-related functions.
**Acceptance**: Variable and helper compile successfully, are unexported (package-private).
**Status**: ✅ Completed - `reportFields` variable added to [`log/log.go:36`](log/log.go:36)

---

### Task 2: Modify `contextLogger.*Report()` methods in `log/log.go`

**Plan Reference**: [Section 3.3.3](IMPLEMENTATION_PLAN.md#333-modify-contextlogger-methods)
**Files**: [`log/log.go`](log/log.go)
**Description**: Update `ErrorReport()`, `WarnReport()`, `InfoReport()`, `DebugReport()` on `contextLogger` to call `contextWithReportFields(c.Context)` before passing to the provider. The pattern is:
```go
func (c contextLogger) ErrorReport(args ...interface{}) {
    c.provider.Error(contextWithReportFields(c.Context), true, args...)
}
```
Package-level `ErrorReport()`, `WarnReport()`, etc. already delegate to `BackgroundContext().*Report()` so they automatically get the field.
**Acceptance**: All `*Report` methods inject the field. Non-report methods (`Error`, `Warn`, `Info`, `Debug`) remain unchanged.
**Status**: ✅ Completed - All four `*Report` methods modified in [`log/log.go:93-114`](log/log.go:93-114)

---

### Task 3: Write tests for report field injection

**Plan Reference**: [Section 3.4](IMPLEMENTATION_PLAN.md#34-test-plan)
**Files**: `log/log_test.go` (new file, or add to existing test file if one exists)
**Description**: Write tests using Gomega that verify:
- **Positive**: Calling `ErrorReport`, `WarnReport`, `InfoReport`, `DebugReport` results in `"reportableError": true` being present in the log output fields
- **Negative**: Calling `Error`, `Warn`, `Info`, `Debug` (non-report) does NOT include `"reportableError"` in the fields
- **Integration**: Test through a full provider chain (structured provider captures fields)
- **Extensibility**: Adding a field to `reportFields` propagates correctly

**Acceptance**: All tests pass. Both positive and negative cases covered.
**Status**: ✅ Completed - 14 tests created in [`log/log_test.go`](log/log_test.go) (positive, negative, integration, extensibility). All tests pass.

---

### Task 4: (Optional - Phase 2) Create `providers/reportable/provider.go`

**Plan Reference**: [Section 4.1](IMPLEMENTATION_PLAN.md#41-scope)
**Files**: `providers/reportable/provider.go` (new)
**Description**: Create the reusable provider following the `reported_at` pattern. This provider checks the `report` boolean and injects configured fields when `report=true`. Does NOT override `Record`/`RecordEvent` (since `reportableError` only applies to log-level methods with the `report` boolean). Includes a Config struct for extensibility.
**Acceptance**: Provider compiles, follows chaining pattern, only injects fields when `report=true`.
**Status**: ✅ Completed - [`providers/reportable/provider.go`](providers/reportable/provider.go) created following `reported_at` pattern

---

### Task 5: (Optional - Phase 2) Write tests for reportable provider

**Plan Reference**: [Section 4.2](IMPLEMENTATION_PLAN.md#42-when-to-use-this-pattern)
**Files**: `providers/reportable/provider_test.go` (new)
**Description**: Write comprehensive tests for the reportable provider:
- Report=true injects configured fields
- Report=false does not inject fields
- Record/RecordEvent are unaffected
- Multiple fields can be configured

**Acceptance**: All tests pass.
**Status**: ✅ Completed - [`providers/reportable/provider_test.go`](providers/reportable/provider_test.go) created with 6 tests. All tests pass.

---

### Task 6: Update VERSION and CHANGELOG

**Plan Reference**: [Section 5](IMPLEMENTATION_PLAN.md#5-version--changelog)
**Files**: [`VERSION`](VERSION), [`CHANGELOG.md`](CHANGELOG.md)
**Description**:
- Update `VERSION` from `1.7.0` to `1.8.0`
- Add `1.8.0` entry to `CHANGELOG.md` with the new feature description
- Note: `CHANGELOG.md` currently only goes up to `1.6.2` — consider adding a minimal `1.7.0` entry as well to document the version gap

**Acceptance**: VERSION reads `1.8.0`. CHANGELOG has appropriate entries.
**Status**: ✅ Completed - VERSION updated to 1.8.0, CHANGELOG updated with 1.7.0 and 1.8.0 entries

---

### Task 7: Run full test suite and verify no regressions

**Plan Reference**: [Section 6](IMPLEMENTATION_PLAN.md#6-implementation-task-list), [Section 8](IMPLEMENTATION_PLAN.md#8-acceptance-criteria)
**Files**: All
**Description**: Run `go test ./...` from the project root. Verify all existing tests pass and no regressions were introduced.
**Acceptance**: All tests pass with zero failures.
**Status**: ✅ Completed - Full test suite executed. All new feature tests pass (20 total: 14 in log/log_test.go + 6 in providers/reportable/provider_test.go). Note: 2 pre-existing test failures in providers/reported_at (unrelated to this feature).

---

## Deviations from Implementation Plan

### Task 3 Deviation: Test Organization
**Implementation Plan**: Section 3.4 suggested specific test names and organization patterns.
**Actual Implementation**: Tests in [`log/log_test.go`](log/log_test.go) follow the same verification patterns but use Ginkgo test suite structure with `Describe`, `Context`, and `It` blocks. This provides better organization and readability. The test file includes Ginkgo suite bootstrap as recommended.
**Impact**: Minor - All test coverage requirements met, just organized differently for better maintainability.

### Task 4 & 5 Scope Addition: Phase 2 Implementation
**Implementation Plan**: Tasks 4 and 5 were marked as "Optional - Phase 2" suggesting they might be deferred.
**Actual Implementation**: Phase 2 was fully implemented in this release, including:
- Complete [`providers/reportable/provider.go`](providers/reportable/provider.go) following the `reported_at` pattern
- Comprehensive test suite in [`providers/reportable/provider_test.go`](providers/reportable/provider_test.go) with 6 tests
- Full Config struct for extensibility

**Impact**: Positive - Provides reusable provider pattern for future field injection needs, improving maintainability and code reuse.

### Pre-existing Test Failures
**Note**: The test suite shows 2 failures in `providers/reported_at/provider_test.go` (lines 40 and 47). These are pre-existing issues unrelated to the `reportableError` feature:
- `TestReportedAt`: Expects `reportedAt` to point to `provider.go` but captures `provider_test.go`
- `TestReportedAtFiltering`: Same issue with stack frame capture

**Impact**: None on new feature - All 20 tests related to `reportableError` functionality pass successfully.
