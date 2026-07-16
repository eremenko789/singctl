# Tasks: OpenAPI Codegen Pipeline (F03)

**Input**: Design documents from `/specs/003-openapi-codegen/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story, которая меняет executable production-код (Make-скрипты, хелперы подсчёта ops, generate wiring), сначала добавляются падающие тесты/контракты, затем минимальная реализация. Сгенерированный `*.gen.go` MAY исключаться из coverage gate. Покрытие ручного кода MUST NOT регрессировать (`make test`).

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: `[US1]`…`[US4]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: каталоги и предпосылки tooling под F03 (поверх F01 Makefile).

- [X] T001 Create directories `api/` and `internal/apiclient/` (placeholder `.gitkeep` only if needed so empty dirs are tracked) in `api/`, `internal/apiclient/`
- [X] T002 [P] Confirm `oapi-codegen` install instructions are present/accurate in `.env.example` and align with `docs/openapi-codegen.md` (`.env.example`, `docs/openapi-codegen.md`)
- [X] T003 [P] Verify committed OpenAPI snapshot and matrix exist: `docs/api/openapi.json`, `docs/api/openapi.yaml`, `docs/api/coverage.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: зафиксировать независимость Make-таргетов и зелёный baseline тестов до story-работы.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Confirm Makefile targets `openapi-fetch`, `api-coverage-check`, and `generate` have **no** Make prerequisites on each other in `Makefile`
- [X] T011 [P] Confirm `make help` lists `openapi-fetch`, `api-coverage-check`, and `generate` with RU descriptions in `Makefile`
- [X] T012 Run coverage-gated baseline via `make test` (Makefile) and ensure current suite is green before F03 changes

**Checkpoint**: Foundation ready — user stories can proceed (US1–US4 независимо по контракту; US4 DoD требует артефакты US1).

---

## Phase 3: User Story 1 — Generate typed API client from snapshot (Priority: P1) 🎯 MVP

**Goal**: `make generate` по `api/oapi-codegen.yaml` + `docs/api/openapi.yaml` создаёт `internal/apiclient/client.gen.go` (models+client, package `apiclient`); без конфига — ненулевой exit и понятная ошибка.

**Independent Test**: При установленном `oapi-codegen` выполнить `make generate` и `go build ./internal/apiclient/...`; убедиться, что `client.gen.go` на месте. Без `api/oapi-codegen.yaml` — `make generate` ≠ 0.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

> Write these FIRST; ensure they FAIL before adding config/running generate successfully.

- [X] T020 [P] [US1] Add failing contract test that `internal/apiclient/client.gen.go` must exist after pipeline readiness in `internal/openapipipeline/generate_contract_test.go` (add `internal/openapipipeline/doc.go` for package)
- [X] T021 [P] [US1] Add failing contract test that package `apiclient` builds (`go/build` or `exec` of `go build ./internal/apiclient/...`) in `internal/openapipipeline/generate_contract_test.go`
- [X] T022 [US1] Add failing/asserting contract test that `make generate` exits non-zero when `api/oapi-codegen.yaml` is absent (temp rename) in `internal/openapipipeline/generate_missing_config_test.go`

### Implementation for User Story 1

- [X] T030 [US1] Create codegen config `api/oapi-codegen.yaml` (`package: apiclient`, `models: true`, `client: true`, `output: internal/apiclient/client.gen.go`)
- [X] T031 [US1] Ensure `make generate` creates output directory if needed and invokes `oapi-codegen -config api/oapi-codegen.yaml docs/api/openapi.yaml` in `Makefile`
- [X] T032 [US1] Run `make generate` to produce `internal/apiclient/client.gen.go` and run `go mod tidy` for generator runtime deps in `go.mod` / `go.sum`
- [X] T033 [US1] Exclude `internal/apiclient/*.gen.go` from coverage collection/gate in `Makefile` `test` target (constitution IX MAY)
- [X] T034 [US1] Turn US1 contract tests green without weakening assertions in `internal/openapipipeline/*_test.go`

### Coverage gate for US1

- [X] T035 Run coverage-gated tests via `make test` after US1 (Makefile)

**Checkpoint**: MVP — offline `make generate` works; gen client present and builds.

---

## Phase 4: User Story 2 — Refresh OpenAPI snapshot from upstream (Priority: P1)

**Goal**: `make openapi-fetch` атомарно обновляет пару `docs/api/openapi.json` + `docs/api/openapi.yaml`; при сбое второго скачивания целевые файлы не остаются рассинхронизированными.

**Independent Test**: Успешный `make openapi-fetch` (сеть) обновляет оба файла; контрактный тест атомарности на temp-файлах проходит без реального upstream (если вынесен в скрипт).

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US2] Add failing tests for atomic dual-file replace (both temps OK → both finals updated; second download fail → finals unchanged) in `scripts/openapi_fetch_atomic_test.sh` (or `scripts/openapi_fetch_test.py`)
- [X] T041 [US2] Add failing contract asserting `openapi-fetch` recipe does not `curl -o` directly onto finals before both downloads succeed (review harness) in `scripts/openapi_fetch_atomic_test.sh`

### Implementation for User Story 2

- [X] T050 [US2] Implement atomic fetch helper (download JSON+YAML to temps, then `mv` both) in `scripts/openapi_fetch.sh` (or inline equivalent kept testable)
- [X] T051 [US2] Wire `openapi-fetch` target to call the helper with `OPENAPI_JSON_URL` / `OPENAPI_YAML_URL` in `Makefile`
- [X] T052 [US2] Ensure failure paths exit non-zero and leave prior snapshot pair consistent in `scripts/openapi_fetch.sh` / `Makefile`

### Coverage gate for US2

- [X] T053 Run coverage-gated tests via `make test` after US2 (Makefile); also run `scripts/openapi_fetch_atomic_test.sh` if not hooked into `make test`

**Checkpoint**: Fetch обновляет снимок безопасно; таргет независим от `generate` / `api-coverage-check`.

---

## Phase 5: User Story 3 — Verify API operation coverage count (Priority: P1)

**Goal**: `make api-coverage-check` сверяет ops в `docs/api/openapi.json` с `EXPECTED_API_OPS` (51) и наличие `docs/api/coverage.md`; без парсинга строк матрицы.

**Independent Test**: `make api-coverage-check` → 0 на каноне; `EXPECTED_API_OPS=0 make api-coverage-check` → ≠ 0; удаление/rename `coverage.md` → ≠ 0.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US3] Add failing unit tests for operation counting (51 ops on current `docs/api/openapi.json`; ignore `x-*` keys) in `scripts/count_openapi_ops_test.py`
- [X] T061 [P] [US3] Add failing unit tests for mismatch expected vs actual → non-zero semantics in `scripts/count_openapi_ops_test.py`
- [X] T062 [US3] Add failing test that missing `docs/api/coverage.md` is detected by check flow in `scripts/count_openapi_ops_test.py` or `internal/openapipipeline/coverage_check_test.go`

### Implementation for User Story 3

- [X] T070 [US3] Implement `scripts/count_openapi_ops.py` (count ops; compare to expected; exit codes per `contracts/make-openapi.md`)
- [X] T071 [US3] Wire `api-coverage-check` in `Makefile` to `scripts/count_openapi_ops.py` + `test -f docs/api/coverage.md` (no matrix row/operationId parsing)
- [X] T072 [US3] Keep `EXPECTED_API_OPS ?= 51` overridable via `.env` in `Makefile`
- [X] T073 [US3] Make US3 unit/contract tests green in `scripts/count_openapi_ops_test.py`

### Coverage gate for US3

- [X] T074 Hook `python3 -m unittest scripts/count_openapi_ops_test.py` (or equivalent) into `make test` / dedicated target and run gate in `Makefile`
- [X] T075 Run `make api-coverage-check` and `make test` after US3 (Makefile)

**Checkpoint**: Coverage-check соответствует clarify (ops + file exists only).

---

## Phase 6: User Story 4 — Keep generated client and config in version control (Priority: P1)

**Goal**: DoD F03 — в git попадают конфиг, снимок, матрица и `*.gen.go`; `.env`/токены — нет; офлайн-перегенерация со снимка работает.

**Independent Test**: Список путей из spec FR-010 присутствует в working tree и не игнорируется git; `make generate` офлайн успешен.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US4] Add failing check that critical paths are not ignored by git (`api/oapi-codegen.yaml`, `internal/apiclient/*.gen.go`, `docs/api/openapi.*`, `docs/api/coverage.md`) via `git check-ignore` assertions in `internal/openapipipeline/vcs_dod_test.go`
- [X] T081 [US4] Add failing/asserting offline regenerate contract: with network blocked or without fetch, `make generate` still succeeds when snapshot+config exist in `internal/openapipipeline/generate_contract_test.go`

### Implementation for User Story 4

- [X] T090 [US4] Audit `.gitignore` / `.cursorignore` so DoD artifacts are trackable (fix ignores if any) in `.gitignore`
- [X] T091 [US4] Ensure DoD files exist on disk ready to commit: `api/oapi-codegen.yaml`, `docs/api/openapi.json`, `docs/api/openapi.yaml`, `docs/api/coverage.md`, `internal/apiclient/client.gen.go`
- [X] T092 [US4] Turn US4 VCS/offline tests green in `internal/openapipipeline/vcs_dod_test.go` and related tests

### Coverage gate for US4

- [X] T093 Run coverage-gated tests via `make test` after US4 (Makefile)

**Checkpoint**: DoD артефакты готовы к коммиту (сам commit — по запросу пользователя / отдельный шаг вне tasks checklist, если политика репо требует).

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: docs sync, quickstart, финальные gates.

- [X] T100 [P] Sync `docs/openapi-codegen.md` with actual targets (independent targets, recommended order, atomic fetch, check strictness, commit list)
- [X] T101 [P] Sync `docs/makefile.md` with actual `Makefile` target behavior
- [X] T102 [P] Ensure `make help` text mentions recommended order `openapi-fetch` → `api-coverage-check` → `generate` in `Makefile` and/or docs
- [X] T103 Run full quickstart validation from `specs/003-openapi-codegen/quickstart.md` (Steps 1–2 required; Step 3 optional network)
- [X] T104 Confirm no hand-written HTTP CRUD/DTO and no new entity CLI commands were added (spot-check `internal/`, `internal/cli/`)
- [X] T105 Final coverage gate via `make test` (Makefile)
- [X] T106 Final acceptance: `make api-coverage-check` && `make generate` (Makefile)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: start immediately
- **Foundational (Phase 2)**: after Setup — BLOCKS all user stories
- **US1 (Phase 3)**: after Foundational — 🎯 MVP
- **US2 (Phase 4)**: after Foundational — parallelizable with US1/US3 (different files: `scripts/openapi_fetch*`, `Makefile` fetch recipe)
- **US3 (Phase 5)**: after Foundational — parallelizable with US1/US2 (careful: shared `Makefile` edits → serialize Makefile touches or merge carefully)
- **US4 (Phase 6)**: after US1 (needs `client.gen.go` + config); benefits from US2/US3 completeness for full DoD snapshot freshness
- **Polish (Phase 7)**: after desired stories complete (ideally all)

### User Story Dependencies

- **US1**: no dependency on US2/US3
- **US2**: no dependency on US1/US3
- **US3**: no dependency on US1/US2
- **US4**: depends on US1 artifacts; snapshot/matrix assumed present (Setup T003)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Implementation before coverage gate
- `make test` MUST pass without coverage drop at story checkpoint
- Do not add Make prerequisite linking `generate` → `api-coverage-check`

### Parallel Opportunities

- T002/T003 in Setup
- T020/T021 in US1 tests
- T040 parallel with US1 if different owners (watch `Makefile`)
- T060/T061 in US3 tests
- T100/T101/T102 in Polish
- **Note**: concurrent edits to `Makefile` are NOT [P] — serialize or single owner

---

## Parallel Example: User Story 1

```bash
# Tests first (must fail):
Task: "T020 contract test client.gen.go exists in internal/openapipipeline/generate_contract_test.go"
Task: "T021 contract test go build ./internal/apiclient/... in internal/openapipipeline/generate_contract_test.go"

# Then implementation:
Task: "T030 Create api/oapi-codegen.yaml"
Task: "T031/T032 make generate + go mod tidy"
Task: "T033 Exclude *.gen.go from coverage in Makefile"
Task: "T035 make test"
```

---

## Parallel Example: User Story 3

```bash
Task: "T060 count ops unit tests in scripts/count_openapi_ops_test.py"
Task: "T061 mismatch expected tests in scripts/count_openapi_ops_test.py"
# After fail:
Task: "T070 Implement scripts/count_openapi_ops.py"
Task: "T071 Wire Makefile api-coverage-check"
Task: "T075 make api-coverage-check && make test"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 Setup
2. Phase 2 Foundational
3. Phase 3 US1 (`api/oapi-codegen.yaml` + `make generate` + committed-ready `client.gen.go`)
4. **STOP and VALIDATE** via US1 Independent Test + `make test`
5. Optionally continue US2/US3/US4 for full F03 DoD

### Incremental Delivery

1. Setup + Foundational
2. US1 → MVP codegen
3. US3 → hardened coverage-check (часто уже почти готов)
4. US2 → atomic fetch
5. US4 → VCS DoD assertions
6. Polish → docs + quickstart

### Parallel Team Strategy

1. Together: Setup + Foundational
2. Then:
   - Dev A: US1 (+ US4)
   - Dev B: US3 (scripts + Makefile check section)
   - Dev C: US2 (fetch script)
3. Merge Makefile changes carefully; one integrator for Polish

---

## Notes

- [P] = different files, no incomplete-task dependencies
- Contract source of truth: `specs/003-openapi-codegen/contracts/make-openapi.md`
- No drift/no-diff gate in F03 (spec FR-013)
- No hand-written HTTP CRUD; no entity CLI commands
- Prefer not inventing `internal/api` adapter (F04)
- Commit after each story checkpoint when asked by user
- Validate against `specs/003-openapi-codegen/quickstart.md` before calling F03 done
