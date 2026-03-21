# CLI Code Review v1 - Implementation Plan

**Created**: 2026-03-21
**Go Version**: 1.26.1
**Module**: github.com/shoehorn-dev/cli
**Approach**: TDD (write failing tests first, then fix, verify pass)
**Constraint**: Zero regressions - all existing tests must pass throughout

---

## Phase 1: Critical Fixes (exit code correctness + security)

### 1.1 Fix Error Swallowing (exit code 0 on failure)

**Problem**: 3 locations return `nil` after printing error boxes, causing exit code 0 on failure.
**Files**: `auth.go:87`, `forge.go:269`, `forge.go:452`
**Risk**: Low - changes `return nil` to `return err/spinErr`
**Flagged by**: logging-master, code-reviewer, architect, exception-handling

**TDD approach**:
1. Write tests that verify non-nil error returns on failure paths
2. Fix the three return statements
3. Verify tests pass + existing tests still pass

### 1.2 Fix Homebrew Tap Token Exposure (CI/CD)

**Problem**: TAP_TOKEN secret embedded in git clone URL in release.yml
**File**: `.github/workflows/release.yml:511`
**Risk**: Low - uses actions/checkout instead of manual git clone
**Flagged by**: build-master

---

## Phase 2: Error Handling Foundation

### 2.1 Introduce Typed Errors (APIError, sentinel errors)

**Problem**: Zero custom error types. Exit code classification uses fragile string matching.
**Files**: New `pkg/api/errors.go`, update `pkg/ui/exit_codes.go`
**Risk**: Low - additive change, existing error strings preserved
**Flagged by**: exception-handling

**TDD approach**:
1. Write tests for APIError type, sentinel errors, errors.Is/errors.As
2. Create error types
3. Update exit_codes.go to use typed errors instead of string matching
4. Regression tests for exit code behavior

### 2.2 Add Error Context Wrapping to Catalog API Methods

**Problem**: ~10 methods in catalog.go return bare `err` without context
**File**: `pkg/api/catalog.go`
**Risk**: Low - only adds wrapping context
**Flagged by**: logging-master, exception-handling

**TDD approach**:
1. Write tests verifying error messages include operation context
2. Add fmt.Errorf wrapping to all bare returns

### 2.3 Fix Silent Error Swallowing in owned.go

**Problem**: `owned.go:55-56` silently swallows errors, shows 0 results
**File**: `cmd/shoehorn/commands/get/owned.go`
**Risk**: Low - adds warning output on error
**Flagged by**: exception-handling

### 2.4 Add Nil-Safety for Type Assertions After Spinner

**Problem**: Unchecked type assertions after RunSpinner could panic
**Files**: Throughout command handlers
**Risk**: Low - adds nil checks before type assertions
**Flagged by**: exception-handling

---

## Phase 3: Test Coverage (highest-value gaps)

### 3.1 Test pkg/config/config.go

**Priority**: Critical - handles auth credentials
**Functions**: Load, Save, GetCurrentProfile, SetProfile, IsAuthenticated, IsPATAuth, IsTokenExpired
**Test types**: Unit, table-driven, property-based, metamorphic

### 3.2 Test pkg/api/client.go

**Priority**: Critical - core HTTP transport
**Functions**: do, doIgnoreStatus, NewClientFromConfig, SetToken
**Test types**: Unit with httptest, error paths, timeout handling

### 3.3 Test pkg/api/catalog.go (pure logic functions)

**Priority**: High - 890 lines, most business logic, zero tests
**Functions**: GetMe, parseOwner, parseMoldInputs, formatLastSeen
**Test types**: Table-driven, property-based, metamorphic

### 3.4 Test pkg/api/manifests.go

**Priority**: Medium - non-standard status code handling
**Functions**: ValidateManifest, ConvertManifest
**Test types**: Unit with httptest

### 3.5 Test pkg/ui/detect.go

**Priority**: Medium - output mode selection
**Functions**: DetectMode, IsInteractive, ShouldUseColor
**Test types**: Table-driven with all flag/env combinations

---

## Phase 4: Code Quality + Refactoring

### 4.1 Add URL Path Encoding for User-Provided Path Segments

**Problem**: Entity IDs, team slugs injected directly into URL paths without encoding
**File**: `pkg/api/catalog.go` (multiple lines)
**Risk**: Low - adds url.PathEscape calls
**Flagged by**: code-reviewer

**TDD approach**:
1. Write tests with IDs containing special chars (/, ?, #, spaces)
2. Add url.PathEscape to all user-provided path segments

### 4.2 Remove Dead Code

**Problem**: Unused functions and flags
**Files**: `pkg/ui/exit_codes.go` (ExitWithError/ExitWithMessage), `pkg/ui/detect.go` (IsInteractive, ShouldUseColor), duplicate `formatBytes` in addon.go
**Risk**: Low - removing unused code
**Flagged by**: code-reviewer, architect

### 4.3 Fix Duplicate formatBytes / FormatBuildSize

**Problem**: Same function implemented twice
**Files**: `pkg/addon/builder.go:62` and `cmd/shoehorn/commands/addon.go:405`
**Risk**: Low - remove duplicate, use existing
**Flagged by**: code-reviewer

### 4.4 Refactor Concurrent Entity Fetch to Use errgroup

**Problem**: Fragile ad-hoc channel pattern in entity detail view
**File**: `cmd/shoehorn/commands/get/entities.go:150-196`
**Risk**: Medium - behavioral change in concurrent code
**Flagged by**: concurrency-expert, code-reviewer

**TDD approach**:
1. Write tests for the concurrent fetch behavior (success, partial failure, all fail)
2. Refactor to use golang.org/x/sync/errgroup
3. Verify identical behavior

### 4.5 Split catalog.go into Domain-Specific Files

**Problem**: 890-line file covering 7 domains violates SRP
**File**: `pkg/api/catalog.go` -> `entities.go`, `teams.go`, `users.go`, `groups.go`, `search.go`, `k8s.go`, `forge_api.go`
**Risk**: Low - mechanical file split, no logic changes
**Flagged by**: architect

### 4.6 Fix hasScheme URL Check

**Problem**: Brittle manual string slicing instead of strings.HasPrefix
**File**: `cmd/shoehorn/commands/auth.go:241-243`
**Risk**: Low
**Flagged by**: code-reviewer

---

## Phase 5: Observability

### 5.1 Add --debug Flag and SHOEHORN_DEBUG Env Var

**Problem**: Zero diagnostic visibility for users
**Files**: `root.go`, new `pkg/logger/` or use slog
**Risk**: Low - additive, no behavior change without flag
**Flagged by**: logging-master

### 5.2 Add Request Timing to API Client

**Problem**: No visibility into slow requests
**File**: `pkg/api/client.go`
**Risk**: Low - only logs when debug enabled
**Flagged by**: logging-master

### 5.3 Add Token Sanitization Helper

**Problem**: No guardrails against accidental token logging
**Risk**: Low - additive utility
**Flagged by**: logging-master

---

## Phase 6: CI/CD Improvements

### 6.1 Add Concurrency Control to Workflows

**Problem**: No cancellation of superseded CI runs
**Files**: `.github/workflows/ci.yml`, `.github/workflows/release.yml`
**Flagged by**: build-master

### 6.2 Parallelize CI Jobs

**Problem**: Lint, test, build run sequentially
**File**: `.github/workflows/ci.yml`
**Flagged by**: build-master

### 6.3 Add go test -race to CI

**Problem**: Race detector not in CI pipeline
**File**: `.github/workflows/ci.yml`
**Flagged by**: concurrency-expert

### 6.4 Add govulncheck to CI

**Problem**: No vulnerability scanning
**File**: `.github/workflows/ci.yml`
**Flagged by**: build-master

### 6.5 Add Dependabot for GitHub Actions

**Problem**: No automated action update mechanism
**File**: New `.github/dependabot.yml`
**Flagged by**: build-master

---

## Phase 7: Minor Improvements

### 7.1 Use atomic.Bool for plainMode

**File**: `pkg/tui/spinner.go:16`
**Flagged by**: concurrency-expert

### 7.2 Add signal.Stop Cleanup

**File**: `cmd/shoehorn/commands/addon_dev.go:47-48`
**Flagged by**: concurrency-expert

### 7.3 Use exec.CommandContext

**File**: `cmd/shoehorn/commands/addon_dev.go:41`
**Flagged by**: concurrency-expert

### 7.4 Fix Shared Flag Variables Between Commands

**File**: `cmd/shoehorn/commands/forge.go:103-112`
**Flagged by**: code-reviewer, architect

### 7.5 Standardize Error Message Style (drop "failed to" prefix)

**Files**: Throughout codebase
**Flagged by**: exception-handling

### 7.6 Add Cancellable Contexts to Commands

**Problem**: All commands use context.Background() with no cancellation
**Flagged by**: exception-handling

---

## Validation Protocol

Before each phase completion:
```bash
go test ./... -race -count=1     # All tests pass with race detector
go vet ./...                     # No vet warnings
go build ./cmd/shoehorn          # Clean build
```

## Phase 8: Security Fixes - Immediate (HIGH severity)

### 8.1 Support SHOEHORN_TOKEN Env Var + Stdin for Token Input

**Problem**: `--token` flag exposes PAT in process lists (`ps aux`)
**File**: `cmd/shoehorn/commands/auth.go:55-70`
**Severity**: HIGH
**Flagged by**: security-master (Finding 1)

**TDD approach**:
1. Write test: env var token takes effect when --token empty
2. Write test: --token still works (backward compat)
3. Write test: warning printed when --token used
4. Implement SHOEHORN_TOKEN env var support
5. Add stderr warning when --token flag used

### 8.2 Warn/Block HTTP with Non-Localhost Servers

**Problem**: Users can send Bearer tokens over plaintext HTTP to remote servers
**Files**: `cmd/shoehorn/commands/auth.go`, `config.go:80`
**Severity**: HIGH
**Flagged by**: security-master (Finding 2)

**TDD approach**:
1. Write test: HTTP + localhost allowed without warning
2. Write test: HTTP + remote host returns error
3. Write test: HTTPS + remote host works fine
4. Add URL validation in runLogin before API call

---

## Phase 9: Security Fixes - Medium Priority

### 9.1 Add Response Body Size Limit (DoS Protection)

**Problem**: `io.ReadAll` without size limit, malicious server can OOM
**Files**: `pkg/api/client.go:77,130`, `pkg/api/addons.go:212`
**Severity**: MEDIUM
**Flagged by**: security-master (Finding 4)

**TDD approach**:
1. Write test: normal response (<10MB) works
2. Write test: oversized response returns error
3. Replace io.ReadAll with io.LimitReader

### 9.2 URL-Escape Addon Slugs in API Paths

**Problem**: 6 addon API paths missing url.PathEscape (path traversal risk)
**File**: `pkg/api/addons.go:89,107,115,123,137,197`
**Severity**: MEDIUM
**Flagged by**: security-master (Finding 6)

**TDD approach**:
1. Write test: slug with special chars is properly escaped in URL
2. Add url.PathEscape to all 6 addon path constructions

### 9.3 URL-Encode Marketplace Kind Parameter

**Problem**: `kind` parameter string-concatenated into query (injection risk)
**File**: `pkg/api/addons.go:151-153`
**Severity**: LOW
**Flagged by**: security-master (Finding 7)

### 9.4 Warn on Loose Config File Permissions

**Problem**: Config loaded without checking permissions could be world-readable
**File**: `pkg/config/config.go:66-97`
**Severity**: LOW
**Flagged by**: security-master (Finding 12)

**TDD approach**:
1. Write test: Load() on 0600 file produces no warning
2. Implement permission check on POSIX systems

---

## Phase 10: Future Security Enhancements (TODO)

These are documented for future implementation:

### 10.1 OS Keychain Integration for Token Storage
- macOS Keychain, Windows Credential Manager, Linux secret-service
- Eliminates plaintext token storage on disk
- **Priority**: Future sprint

### 10.2 Custom CA Certificate Support
- `--ca-cert` flag and `SHOEHORN_CA_CERT` env var
- Enterprise TLS proxy compatibility
- **Priority**: Future sprint

### 10.3 Client-Side Rate Limiting with Backoff
- Exponential backoff on HTTP 429 and 5xx
- Prevent thundering herd in scripted usage
- **Priority**: Future sprint

### 10.4 Server-Side Token Revocation on Logout
- Call revocation endpoint if API supports it
- Currently logout only clears local credentials
- **Priority**: Future sprint (requires API support)

### 10.5 Token Encryption at Rest
- Encrypt tokens in config.yaml using OS-derived key
- Defense in depth for backup/sync exposure
- **Priority**: Future sprint

---

## Validation Protocol

Before each phase completion:
```bash
go test ./... -count=1              # All tests pass
go vet ./...                        # No vet warnings
go build ./cmd/shoehorn             # Clean build
```

## Scope Limits

- Max 400 lines per new file
- Each phase is independently shippable
- No behavioral changes without regression tests
- Preserve all existing public API signatures
