# Tasks: Error Handling & Retry (F05)

**Input**: Design documents from `/specs/005-error-retry/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Фикстуры токенов только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Tasks grouped by user story. Labels: `[US1]` taxonomy messages, `[US2]` 429 retry, `[US3]` client validation (token/date), `[US4]` exit codes + `config validate`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]` / `[US2]` / `[US3]` / `[US4]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: подтвердить F04 baseline и контракты F05 перед TDD.

- [X] T001 Confirm F04 package builds: `go test ./internal/api/...` (session, `HTTPError`, validate probe present)
- [X] T002 [P] Skim `specs/005-error-retry/contracts/api-errors-retry.md` and `specs/005-error-retry/contracts/cli-exit-codes.md` for acceptance strings to mirror in tests
- [X] T003 [P] Skim message catalog in `specs/005-error-retry/data-model.md` (401/403/404/422/429/5xx + exit map)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: зелёный coverage baseline до изменений F05.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F05 changes

**Checkpoint**: Foundation ready — US1 can start.

---

## Phase 3: User Story 1 — Predictable HTTP error messages (Priority: P1) 🎯 MVP

**Goal**: `Classify` поверх `*HTTPError` → `ClassifiedError` с Kind + стабильными EN-сообщениями по ТЗ §8.1 / catalog; 422 из тела; 404 с опциональным EntityID; без retry.

**Independent Test**: Unit-таблица: `EnsureSuccess`/`HTTPError` → `Classify` даёт ожидаемые Message для 401/403/404±ID/422/5xx/429; токен не утекает.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before Classify implementation.

- [X] T020 [P] [US1] Add failing table tests: Classify 401/403/5xx → exact catalog messages in `internal/api/errors_test.go`
- [X] T021 [P] [US1] Add failing tests: Classify 404 without ID → `Error: entity not found`; with `WithEntityID` → `Error: entity not found: <ID>` in `internal/api/errors_test.go`
- [X] T022 [P] [US1] Add failing tests: Classify 422 extracts JSON `message`/`error`/`detail` or plain text; empty/binary → `Error: validation failed` in `internal/api/errors_test.go`
- [X] T023 [US1] Add failing test: Classify 429 → `Error: rate limited. Retry later` in `internal/api/errors_test.go`
- [X] T024 [US1] Add failing test: Classify nil → nil; unwrap/`errors.As` to `*HTTPError` / `*ClassifiedError` in `internal/api/errors_test.go`

### Implementation for User Story 1

- [X] T030 [US1] Implement `Kind`, `ClassifiedError`, `Classify`, `WithEntityID` (and 422 body extract helpers) in `internal/api/errors.go` per `contracts/api-errors-retry.md` and `data-model.md`
- [X] T031 [US1] Make US1 tests green in `internal/api/errors_test.go` (no token in Message; godoc without package-stutter)

### Coverage gate for US1

- [X] T032 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP — taxonomy messages ready offline (no network).

---

## Phase 4: User Story 2 — Automatic retry on rate limit (Priority: P1)

**Goal**: Session HTTP client retries only HTTP 429 up to 3 attempts; exponential 1s/2s or `Retry-After` (cap 30s); injectable sleeper; non-429 = 1 attempt.

**Independent Test**: `httptest` 429→429→200 → 3 requests + success; 429×3 → 3 requests + Classified rate limited; 404 → 1 request; sleeper records durations for exponential and `Retry-After`.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US2] Add failing test: 429 then 429 then 200 → exactly 3 HTTP hits and success via session/probe in `internal/api/retry_test.go`
- [X] T041 [P] [US2] Add failing test: 429×3 → exactly 3 hits and Classify message `Error: rate limited. Retry later` in `internal/api/retry_test.go`
- [X] T042 [P] [US2] Add failing test: 401/404/422/5xx → exactly 1 hit (no retry) in `internal/api/retry_test.go`
- [X] T043 [US2] Add failing test: without `Retry-After`, sleeper sees 1s then 2s (injectable clock) in `internal/api/retry_test.go`
- [X] T044 [US2] Add failing test: `Retry-After: 1` respected; oversized value capped at 30s in `internal/api/retry_test.go`
- [X] T045 [US2] Add failing test: context cancel during backoff stops retries in `internal/api/retry_test.go`

### Implementation for User Story 2

- [X] T050 [US2] Implement retry RoundTripper + policy constants (max 3, 1s/2s, cap 30s, injectable sleeper) in `internal/api/retry.go`
- [X] T051 [US2] Wire retry transport into session `http.Client` in `internal/api/session.go`
- [X] T052 [US2] Update `ValidateConnectivity` docs/comments (logical call may issue up to 3 HTTP on 429); return `Classify(err)` on failure in `internal/api/validate.go`
- [X] T053 [US2] Make US2 tests green in `internal/api/retry_test.go` (suite stays fast — no real multi-second sleeps)

### Coverage gate for US2

- [X] T054 Run `make test` after US2 (Makefile)

**Checkpoint**: 429 retry policy proven with mock HTTP.

---

## Phase 5: User Story 3 — Client-side validation errors (Priority: P2)

**Goal**: `ParseDate` + `DateError` with `Expected: YYYY-MM-DD`; Classify/config path for missing token as Kind Config (exit semantics later in US4).

**Independent Test**: Unit `ParseDate` rejects bad input with hint; empty token factory/Classify → Config kind (no network).

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US3] Add failing tests: `ParseDate` accepts `YYYY-MM-DD`; rejects empty/wrong layout/invalid calendar with `Expected: YYYY-MM-DD` in `internal/api/date_test.go`
- [X] T061 [P] [US3] Add failing test: Classify / map missing-token / factory empty-token error → Kind Config in `internal/api/errors_test.go` (or `session_test.go`)

### Implementation for User Story 3

- [X] T070 [US3] Implement `ParseDate` and `DateError` in `internal/api/date.go`
- [X] T071 [US3] Ensure Classify maps DateError and config/missing-token errors to correct Kind in `internal/api/errors.go`
- [X] T072 [US3] Make US3 tests green in `internal/api/date_test.go`, `internal/api/errors_test.go`

### Coverage gate for US3

- [X] T073 Run `make test` after US3 (Makefile)

**Checkpoint**: Client validation contracts closed without CLI date command.

---

## Phase 6: User Story 4 — Stable exit semantics for scripts (Priority: P2)

**Goal**: `cli.ExitCode` 0/1/2/3; `main` uses it; `config validate` surfaces Classified messages and correct exit codes (404→3, no token→2, API/transport→1).

**Independent Test**: Table `ExitCode`; CLI tests for validate with httptest 401/404/429×3/5xx + no-token → expected codes/substrings.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US4] Add failing table tests for `ExitCode`: nil→0, NotFound→3, Config→2, other kinds/transport/raw HTTP→1 in `internal/cli/exit_test.go`
- [X] T081 [P] [US4] Add failing CLI test: `config validate` without token → exit semantics 2 + set-token hint in `internal/cli/config_validate_test.go`
- [X] T082 [P] [US4] Add failing CLI test: mock 401 → message catalog + ExitCode 1 in `internal/cli/config_validate_test.go`
- [X] T083 [P] [US4] Add failing CLI test: mock 404 → catalog not-found + ExitCode 3 in `internal/cli/config_validate_test.go`
- [X] T084 [US4] Add failing CLI test: mock persistent 429 → 3 mock hits, rate-limited message, ExitCode 1 in `internal/cli/config_validate_test.go`
- [X] T085 [US4] Add failing CLI test: mock 5xx → server error message, ExitCode 1 in `internal/cli/config_validate_test.go`

### Implementation for User Story 4

- [X] T090 [US4] Implement `ExitCode(err error) int` in `internal/cli/exit.go` per `contracts/cli-exit-codes.md`
- [X] T091 [US4] Update `cmd/singctl/main.go` to `os.Exit(cli.ExitCode(err))` instead of always `1`
- [X] T092 [US4] Wire `config validate` to return Classified errors (messages from catalog) in `internal/cli/config_validate.go`
- [X] T093 [US4] Make US4 tests green in `internal/cli/exit_test.go`, `internal/cli/config_validate_test.go` (assert via `ExitCode(err)` and/or execute harness)

### Coverage gate for US4

- [X] T094 Run `make test` after US4 (Makefile)

**Checkpoint**: Scriptable exit codes + validate UX complete for F05 DoD.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: godoc, edge polish, quickstart DoD, final coverage.

- [X] T100 [P] Add/refresh godoc on exported Classify/ClassifiedError/ParseDate/ExitCode/retry types in `internal/api/*.go`, `internal/cli/exit.go` (no package-stutter)
- [X] T101 [P] Align any remaining F04 comments that claim «exactly one HTTP attempt» always in `internal/api/validate.go` / tests
- [X] T102 Run full `make test` coverage gate (Makefile) — no regression vs T010 baseline
- [X] T103 [P] Walk `specs/005-error-retry/quickstart.md` DoD checklist and mark items done when verified
- [X] T104 Confirm no TUI banners and no entity CRUD added (diff review)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 (Phase 3)**: After Foundational — MVP taxonomy
- **US2 (Phase 4)**: After US1 preferred (Classify on exhausted 429); can start RoundTripper tests in parallel with US1 impl if stubs exist, but wire+Classify after US1
- **US3 (Phase 5)**: After Foundational; Classify Kind mapping after US1
- **US4 (Phase 6)**: After US1 (messages) + US2 (429 validate) + US3 (Config kind); ExitCode can start after US1
- **Polish (Phase 7)**: After desired stories complete

### User Story Dependencies

- **US1 (P1)**: No story deps — MVP
- **US2 (P1)**: Uses Classify for final 429 message; session wire independent of CLI
- **US3 (P2)**: ParseDate independent; Config Kind needs Classify from US1
- **US4 (P2)**: Depends on US1–US3 for full validate matrix; ExitCode table needs Classified kinds

### Within Each User Story

- Tests FAIL before implementation (TDD)
- Implementation → green tests → `make test` coverage gate
- Story checkpoint before next priority when sequential

### Parallel Opportunities

- T002/T003 skim in parallel
- US1 test tasks T020–T022 in parallel
- US2 test tasks T040–T042 in parallel
- US3 T060/T061 in parallel after US1 Classify API shape known
- US4 T080–T083 in parallel after kinds exist
- T100/T101/T103 polish in parallel

---

## Parallel Example: User Story 1

```bash
# Tests first (parallel):
Task: "Classify 401/403/5xx table in internal/api/errors_test.go"
Task: "Classify 404 ± EntityID in internal/api/errors_test.go"
Task: "Classify 422 body extract in internal/api/errors_test.go"

# Then implementation:
Task: "Implement Classify/ClassifiedError in internal/api/errors.go"
Task: "make test"
```

---

## Parallel Example: User Story 2

```bash
# Tests first (parallel):
Task: "429→429→200 three hits in internal/api/retry_test.go"
Task: "429×3 rate limited in internal/api/retry_test.go"
Task: "non-429 single hit in internal/api/retry_test.go"

# Then implementation:
Task: "retry.go RoundTripper + session.go wire"
Task: "make test"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1–2 Setup + baseline
2. Phase 3 US1 Classify catalog
3. **STOP and VALIDATE**: `go test ./internal/api/ -run Classify` / `make test`
4. Demo: table of code → message

### Incremental Delivery

1. US1 → taxonomy messages
2. US2 → 429 retry
3. US3 → ParseDate + Config kind
4. US4 → ExitCode + validate + main
5. Polish → quickstart DoD

### Parallel Team Strategy

1. Together: Setup + Foundational + US1
2. Then: Dev A US2, Dev B US3 (ParseDate), Dev C ExitCode stubs
3. Integrate US4 validate matrix last

---

## Notes

- [P] = different files, no incomplete deps
- Retry only 429; sleeper injectable — CI must stay fast
- Exact catalog strings from `data-model.md` / clarify
- Transport → ExitCode 1; missing token → 2; NotFound → 3
- No TUI; no CRUD commands; no new Make targets required
- Suggested MVP: US1 only; full F05 DoD needs US1–US4
