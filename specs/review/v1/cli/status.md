# CLI Code Review v1 - Status

**Started**: 2026-03-21
**Last Updated**: 2026-03-21 (session 2 - security fixes)

---

## Phase Summary

| Phase | Description | Status | Progress |
|-------|-------------|--------|----------|
| 1 | Critical Fixes (error swallowing, CI security) | DONE | 2/2 |
| 2 | Error Handling Foundation | DONE | 5/5 |
| 3 | Test Coverage (highest-value gaps) | DONE | 4/5 |
| 4 | Code Quality + Refactoring | DONE | 4/6 |
| 5 | Observability | DEFERRED | 0/3 |
| 6 | CI/CD Improvements | DONE | 4/5 |
| 7 | Minor Improvements | DONE | 1/6 |
| 8 | Security - HIGH (token leak, HTTP MITM) | DONE | 2/2 |
| 9 | Security - MEDIUM (DoS, path traversal, perms) | DONE | 4/4 |
| 10 | Security - Future TODO | DOCUMENTED | - |

---

## Detailed Status

### Phase 1: Critical Fixes

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 1.1 | Fix error swallowing (auth.go, forge.go) | DONE | auth_test.go, forge_test.go | Fixed 3 `return nil` -> `return err`. Also fixed brittle `hasScheme()` with `strings.HasPrefix` |
| 1.2 | Fix Homebrew tap token exposure | DONE | N/A (CI) | Replaced git clone with token in URL with actions/checkout + token param |

### Phase 2: Error Handling Foundation

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 2.1 | Introduce typed errors (APIError, sentinels) | DONE | errors_test.go (8 tests) | APIError with Unwrap(), sentinels, errors.Is/As support. Client.do() returns *APIError. |
| 2.2 | Add error wrapping to catalog API methods | DONE | - | All 20 bare `return nil, err` wrapped with fmt.Errorf context |
| 2.3 | Fix silent error swallowing in owned.go | DONE | - | Now prints warnings to stderr for failed team fetches |
| 2.4 | Add nil-safety for type assertions after spinner | DEFERRED | - | Deferred to Phase 4 (requires refactor) |
| 2.5 | Fix ignored errors in builder_test.go | DONE | - | Fixed 2 ignored errors in metamorphic test |

### Phase 3: Test Coverage

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 3.1 | Test pkg/config/config.go | DONE | config_test.go (24 tests) | Table-driven + metamorphic for auth/expiry logic |
| 3.2 | Test pkg/api/client.go | DONE | catalog_test.go (5 tests) | httptest for auth headers, typed errors, status codes |
| 3.3 | Test pkg/api/catalog.go (pure logic) | DONE | catalog_test.go (13 tests) | parseOwner, formatLastSeen + metamorphic tests |
| 3.4 | Test pkg/api/manifests.go | DEFERRED | - | Lower priority, deferred |
| 3.5 | Test pkg/ui/detect.go | DONE | detect_test.go (13 tests) | Table-driven + metamorphic for mode detection |

### Phase 4: Code Quality + Refactoring

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 4.1 | Add URL path encoding | DONE | - | url.PathEscape added to all 10 user-provided path segments in catalog.go |
| 4.2 | Remove dead code | DEFERRED | - | Lower priority |
| 4.3 | Fix duplicate formatBytes/FormatBuildSize | DONE | - | Removed duplicate, using addon.FormatBuildSize |
| 4.4 | Refactor concurrent entity fetch to errgroup | DEFERRED | - | Medium risk, deferred |
| 4.5 | Split catalog.go into domain files | DEFERRED | - | Mechanical, deferred |
| 4.6 | Fix hasScheme URL check | DONE | auth_test.go | Fixed in Phase 1.1 with strings.HasPrefix |

### Phase 5: Observability

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 5.1 | Add --debug flag + SHOEHORN_DEBUG | PENDING | - | |
| 5.2 | Add request timing to API client | PENDING | - | |
| 5.3 | Add token sanitization helper | PENDING | - | |

### Phase 6: CI/CD Improvements

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 6.1 | Add concurrency control to workflows | DONE | - | cancel-in-progress for CI, no-cancel for release |
| 6.2 | Parallelize CI jobs | DEFERRED | - | Current CI is fast enough |
| 6.3 | Add go test -race to CI | DONE | - | Already present in ci.yml |
| 6.4 | Add govulncheck to CI | DEFERRED | - | |
| 6.5 | Add dependabot for actions + gomod | DONE | - | .github/dependabot.yml created |

### Phase 7: Minor Improvements

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 7.1 | Use atomic.Bool for plainMode | DONE | - | Changed to sync/atomic.Bool in spinner.go |
| 7.2 | Add signal.Stop cleanup | PENDING | - | |
| 7.3 | Use exec.CommandContext | PENDING | - | |
| 7.4 | Fix shared flag variables | PENDING | - | |
| 7.5 | Standardize error message style | PENDING | - | |
| 7.6 | Add cancellable contexts to commands | PENDING | - | |

### Phase 8: Security - HIGH Priority

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 8.1 | SHOEHORN_TOKEN env var + stdin + --token warning | DONE | auth_test.go (3 tests) | resolveToken() with flag/env/none sources. Warning on --token flag |
| 8.2 | Warn/block HTTP with non-localhost servers | DONE | auth_test.go (7 tests) | validateServerSecurity() blocks HTTP to remote hosts |

### Phase 9: Security - MEDIUM Priority

| ID | Task | Status | Tests | Notes |
|----|------|--------|-------|-------|
| 9.1 | Response body size limit (io.LimitReader) | DONE | catalog_test.go (1 test) | 10MB cap on client.go + addons.go (3 locations) |
| 9.2 | URL-escape addon slugs in 6 API paths | DONE | - | url.PathEscape on all 6 addon slug paths in addons.go |
| 9.3 | URL-encode marketplace kind parameter | DONE | - | Replaced string concat with url.Values |
| 9.4 | Warn on loose config file permissions | DONE | - | Platform-aware: POSIX warns on >0600, Windows no-op |

### Phase 10: Future Security Enhancements (TODO)

| ID | Task | Status | Notes |
|----|------|--------|-------|
| 10.1 | OS keychain integration | TODO | macOS Keychain, Windows Cred Mgr, Linux keyring |
| 10.2 | Custom CA cert support (--ca-cert) | TODO | Enterprise TLS proxy compat |
| 10.3 | Client-side rate limiting + backoff | TODO | 429/5xx retry with exponential backoff |
| 10.4 | Server-side token revocation on logout | TODO | Requires API support |
| 10.5 | Token encryption at rest | TODO | Encrypt tokens in config.yaml |

---

## Test Runs

| Date | Command | Result | Notes |
|------|---------|--------|-------|
| 2026-03-21 | go test ./... -count=1 | ALL PASS | Baseline + Phase 1.1 verified |
| 2026-03-21 | go test ./... -count=1 | ALL PASS | After Phase 2 complete |
| 2026-03-21 | go test ./... -count=1 | ALL PASS | After Phase 3 - 7 packages now have tests |
| 2026-03-21 | go test ./... -count=1 | ALL PASS | After Phase 4 - URL encoding, dedup, atomic.Bool |
| 2026-03-21 | go test ./... -count=1 | ALL PASS | Session 1 final - phases 1-7 |
| 2026-03-21 | go test ./... -count=1 | ALL PASS | Session 2 - after Phase 8 (HIGH security) |
| 2026-03-21 | go test ./... -count=1 | ALL PASS | Session 2 - after Phase 9 (MEDIUM security) |

---

## Regression Tracking

| Date | Phase | Regression Found | Resolution |
|------|-------|-----------------|------------|
| | | | |
