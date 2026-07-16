# Tasks: API Adapter & Auth (F04)

**Input**: Design documents from `/specs/004-api-adapter-auth/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Фикстуры токенов только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Tasks grouped by user story. Labels: `[US1]` session, `[US1b]` config validate, `[US2]` happy path, `[US3]` non-2xx mapping, `[US4]` shared boundary.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]` / `[US1b]` / `[US2]` / `[US3]` / `[US4]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: каталог адаптера и предпосылки поверх F02/F03.

- [X] T001 Create package directory `internal/api/` with `doc.go` (package `api`, brief godoc: thin adapter over `apiclient`)
- [X] T002 [P] Confirm `internal/apiclient/client.gen.go` builds: `go build ./internal/apiclient/...`
- [X] T003 [P] Confirm F02 config types usable as factory input (`internal/config` Document / EffectiveSettings) — no code change, smoke `go test ./internal/config/...`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: зелёный baseline до story-работы; зафиксировать контрактные пути.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F04 changes
- [X] T011 [P] Skim contracts `specs/004-api-adapter-auth/contracts/api-session.md` and `specs/004-api-adapter-auth/contracts/cli-config-validate.md` for acceptance wording to mirror in tests

**Checkpoint**: Foundation ready — US1 can start.

---

## Phase 3: User Story 1 — Build authenticated API session from config (Priority: P1) 🎯 MVP

**Goal**: Фабрика сеанса из base URL / token / timeout (и `NewFromSettings`) с fail-fast на пустой token/URL и невалидный timeout; сеанс держит `*apiclient.ClientWithResponses` с Bearer editor и `http.Client.Timeout`.

**Independent Test**: Вызвать фабрику с валидными параметрами — Session ≠ nil; с пустым токеном — error, Session == nil, без сети. Override-токен из settings используется при последующих запросах (проверяется в US2).

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before `session.go` implementation.

- [X] T020 [P] [US1] Add failing tests: empty/whitespace token → factory error, no session in `internal/api/session_test.go`
- [X] T021 [P] [US1] Add failing tests: empty/whitespace base URL → factory error in `internal/api/session_test.go`
- [X] T022 [P] [US1] Add failing tests: invalid timeout string → factory error in `internal/api/session_test.go`
- [X] T023 [US1] Add failing test: valid `NewSession` / `NewFromSettings` returns non-nil Session with configured client in `internal/api/session_test.go`

### Implementation for User Story 1

- [X] T030 [US1] Implement `HTTPError`-free session factory: `NewSession`, `NewFromSettings`, Bearer `WithRequestEditorFn`, timeout `WithHTTPClient` in `internal/api/session.go` (per `contracts/api-session.md`, `research.md`)
- [X] T031 [US1] Export `Session` with access to `*apiclient.ClientWithResponses` (field or method) in `internal/api/session.go` without package-stutter names
- [X] T032 [US1] Make US1 tests green in `internal/api/session_test.go` (Russian error messages; no token leakage in error text)

### Coverage gate for US1

- [X] T033 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP — session factory works offline (no network required for fail-fast / construct).

---

## Phase 4: User Story 2 — Authenticated happy-path request (Priority: P1)

**Goal**: Через сеанс выполнить `ProjectControllerListWithResponse` против `httptest`; заголовок `Authorization: Bearer <test-token-…>`; 2xx → успех маппинга; timeout сессии соблюдается.

**Independent Test**: `httptest` отдаёт 200 на `GET /v2/project`; Session с URL мока и `test-token-happy` получает успех и проверяемый Bearer.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US2] Add failing happy-path test: mock 200 + JSON body for `/v2/project`, assert `Authorization: Bearer test-token-happy` in `internal/api/session_test.go` or `internal/api/validate_test.go`
- [X] T041 [P] [US2] Add failing test: bare token does not produce `Bearer Bearer` in `internal/api/session_test.go`
- [X] T042 [US2] Add failing test: short session timeout against hanging handler returns error without hang in `internal/api/session_test.go`

### Implementation for User Story 2

- [X] T050 [US2] Implement probe helper used by happy path / validate: `ValidateConnectivity` (or internal list helper) calling `ProjectControllerListWithResponse` in `internal/api/validate.go`
- [X] T051 [US2] Add minimal success mapping path (2xx → nil) via `EnsureSuccess` in `internal/api/errors.go` if not yet present; wire probe to use it
- [X] T052 [US2] Make US2 tests green in `internal/api/*_test.go`

### Coverage gate for US2

- [X] T053 Run `make test` after US2 (Makefile)

**Checkpoint**: Happy path + Bearer proven with mock HTTP.

---

## Phase 5: User Story 3 — Surface non-success responses without retry (Priority: P2)

**Goal**: Не-2xx → типизированная `*HTTPError` с программно доступным StatusCode (+ Body); ровно одна HTTP-попытка; без taxonomy/retry.

**Independent Test**: Мок 401/404/429 → `errors.As` → `*HTTPError` с ожидаемым status; счётчик запросов к моку == 1.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US3] Add failing unit tests for `EnsureSuccess`: 2xx → nil; 401/404/500 → `*HTTPError` with StatusCode in `internal/api/errors_test.go`
- [X] T061 [P] [US3] Add failing test: non-2xx from probe returns `*HTTPError` and mock saw exactly one request in `internal/api/validate_test.go`
- [X] T062 [US3] Add failing test: transport error (server closed) is not `*HTTPError` (or distinct path) in `internal/api/validate_test.go`

### Implementation for User Story 3

- [X] T070 [US3] Implement `HTTPError` (`StatusCode`, `Body`, `Error()`) and `EnsureSuccess` in `internal/api/errors.go` per `contracts/api-session.md`
- [X] T071 [US3] Ensure probe/validate maps non-2xx through `EnsureSuccess` without retry in `internal/api/validate.go`
- [X] T072 [US3] Make US3 tests green in `internal/api/errors_test.go`, `internal/api/validate_test.go`

### Coverage gate for US3

- [X] T073 Run `make test` after US3 (Makefile)

**Checkpoint**: F05-ready typed HTTP errors; 0 retry.

---

## Phase 6: User Story 1b — Remote config validate via adapter (Priority: P1)

**Goal**: `singctl config validate` при наличии токена вызывает адаптер (probe); успех — remote OK без текста заглушки; сбой — ≠0 без ложного OK; без токена — hint `set-token`.

**Independent Test**: CLI-тест с `httptest` + temp config (`api.base_url` = mock, `api.token` = `test-token-validate`) → exit 0 и remote wording; mock 401 → ≠0; no token → set-token hint.

### Tests for User Story 1b (REQUIRED — TDD) ⚠️

> Update/replace F02 stub assertions in existing file.

- [X] T080 [P] [US1b] Replace/extend `TestConfigValidateWithTokenReportsLocalStubSuccess` with failing remote-success test using `httptest` in `internal/cli/config_validate_test.go`
- [X] T081 [P] [US1b] Add failing test: mock non-2xx → validate nonzero, no «удалённо OK» / stub success wording in `internal/cli/config_validate_test.go`
- [X] T082 [US1b] Keep/adjust `TestConfigValidateWithoutTokenFailsWithHint` still passes (no network) in `internal/cli/config_validate_test.go`
- [X] T083 [US1b] Add failing test: `--token` override used for remote validate against mock in `internal/cli/config_validate_test.go`

### Implementation for User Story 1b

- [X] T090 [US1b] Wire `config validate` to `api.NewFromSettings` + `ValidateConnectivity` in `internal/cli/config_validate.go` (Russian messages per `contracts/cli-config-validate.md`)
- [X] T091 [US1b] Ensure validate never prints full token; remove stub-only success path in `internal/cli/config_validate.go`
- [X] T092 [US1b] Make US1b CLI tests green in `internal/cli/config_validate_test.go`

### Coverage gate for US1b

- [X] T093 Run `make test` after US1b (Makefile)

**Checkpoint**: DoD validate — remote check through adapter.

---

## Phase 7: User Story 4 — Shared adapter boundary for CLI and TUI (Priority: P2)

**Goal**: Auth/base URL/timeout живут в `internal/api`; тесты адаптера не импортируют `internal/cli` / cobra; CLI не собирает raw HTTP в обход адаптера.

**Independent Test**: `go list` / test compile: `internal/api` tests have no import of `github.com/.../internal/cli`; `config_validate.go` imports `internal/api`.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T100 [P] [US4] Add failing/arch test or compile-time assertion that `internal/api` test packages do not import `internal/cli` in `internal/api/boundary_test.go` (or extend existing test file)
- [X] T101 [US4] Add assertion/review test that validate uses adapter package (import present) documented via test comment or thin check in `internal/cli/config_validate_test.go`

### Implementation for User Story 4

- [X] T110 [US4] Fix any accidental CLI↔api dependency leaks; keep session factory as sole auth wiring entry in `internal/api/*.go`, `internal/cli/config_validate.go`
- [X] T111 [US4] Make US4 tests green in `internal/api/boundary_test.go` (and related)

### Coverage gate for US4

- [X] T112 Run `make test` after US4 (Makefile)

**Checkpoint**: Shared boundary ready for future CLI/TUI consumers.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: docs, godoc, final gate, quickstart.

- [X] T120 [P] Add/verify exported godoc on `Session`, `NewSession`, `NewFromSettings`, `HTTPError`, `EnsureSuccess`, `ValidateConnectivity` in `internal/api/*.go` (no package-stutter)
- [X] T121 [P] Sync any user-facing validate help/short text if needed in `internal/cli/config_validate.go` (`Short`/`Long`)
- [X] T122 Run full `make test` coverage gate (Makefile) — no regression
- [X] T123 [P] Walk `specs/004-api-adapter-auth/quickstart.md` checklist offline (`make test` path)
- [X] T124 Confirm no entity CRUD commands and no retry/backoff added (grep/review `internal/api`, `internal/cli`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: immediate
- **Foundational (Phase 2)**: after Setup — BLOCKS stories
- **US1 (Phase 3)**: after Foundational — MVP factory
- **US2 (Phase 4)**: after US1 (needs Session)
- **US3 (Phase 5)**: after US2 preferred (shares probe); can start after US1 if `EnsureSuccess` stub exists — recommend after US2
- **US1b (Phase 6)**: after US2 + US3 (needs probe + `HTTPError`)
- **US4 (Phase 7)**: after US1b (boundary around final wiring)
- **Polish (Phase 8)**: after desired stories

### User Story Dependencies

| Story | Depends on | Independently testable by |
|-------|------------|---------------------------|
| US1 Session | Foundation | factory unit tests only |
| US2 Happy path | US1 | httptest + Bearer assert |
| US3 Non-2xx | US1 (+ probe from US2) | httptest 401 + `errors.As` |
| US1b Validate | US1+US2+US3 | CLI + httptest |
| US4 Boundary | US1b | import/arch tests |

### Within Each User Story

- Tests MUST fail before implementation (TDD)
- `make test` at story checkpoint — no coverage drop
- Token fixtures: `test-token-…` only

### Parallel Opportunities

- T002/T003; T020–T022; T040/T041; T060/T061; T080/T081; T100; T120/T121/T123 after their phase prerequisites

---

## Parallel Example: User Story 1

```bash
# Tests first (parallel):
Task: "empty token factory fail in internal/api/session_test.go"
Task: "empty base URL factory fail in internal/api/session_test.go"
Task: "invalid timeout factory fail in internal/api/session_test.go"

# Then implement session.go; then make test
```

## Parallel Example: User Story 1b

```bash
# After adapter probe exists:
Task: "remote success CLI test in internal/cli/config_validate_test.go"
Task: "remote failure CLI test in internal/cli/config_validate_test.go"
# Then wire config_validate.go
```

---

## Implementation Strategy

### MVP First (US1 only)

1. Phase 1–2 Setup + Foundational
2. Phase 3 US1 session factory
3. **STOP**: validate factory fail-fast + construct via `go test ./internal/api/...`

### Incremental Delivery

1. US1 → factory
2. US2 → Bearer happy path
3. US3 → `HTTPError`
4. US1b → `config validate` remote (full DoD)
5. US4 → boundary
6. Polish → `make test` + quickstart

### Suggested MVP scope

**US1** (session factory) as first demoable increment; **full F04 DoD** = through **US1b** (remote validate) + US2/US3 tests green.

---

## Notes

- Probe operation: `ProjectController_list` / `GET /v2/project` only
- `api.base_url` = origin **without** `/v2`
- No retry; no entity facades; no new CRUD CLI commands
- Commit after each story checkpoint when implementing
- Next command: `/speckit-implement`
