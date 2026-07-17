# Tasks: Project Sections (F12)

**Input**: Design documents from `/specs/012-project-sections/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Токены только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Labels: `[US5]` API section facade (все TaskGroupController_*), `[US1]` list/get CLI, `[US2]` create/update CLI, `[US3]` delete CLI, `[US4]` help discoverability. Фасад (US5) — до CLI-историй: без него list/get не закрывают DoD. `SingleObject` уже есть (F06/F08) — reuse only. Depends F11 (`project` group exists).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]`…`[US5]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: baseline и контракты F12 перед TDD.

- [x] T001 Confirm module builds / existing tests green: `go test ./internal/api/ ./internal/output/ ./internal/cli/ ./cmd/singctl/` (or `make test`) before F12 edits
- [x] T002 [P] Skim `specs/012-project-sections/contracts/api-section-facade.md` and `specs/012-project-sections/data-model.md`
- [x] T003 [P] Skim `specs/012-project-sections/contracts/cli-section.md` and `specs/012-project-sections/contracts/section-output.md`
- [x] T004 [P] Skim `specs/012-project-sections/research.md` (Section vs TaskGroup naming, required parent on list, write surface, columns)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: типы Section/query/write — блокируют фасад и CLI. Output SingleObject — reuse only.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [x] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F12 changes
- [x] T011 Define `Section`, `SectionListQuery`, `SectionWriteInput` types (per data-model) in `internal/api/section.go` (types only / stubs OK)
- [x] T012 Update package doc in `internal/api/doc.go` to mention TaskGroupController / section facade
- [x] T013 Run `make test` after foundation (Makefile)

**Checkpoint**: Foundation ready — US5 (facade) can start.

---

## Phase 3: User Story 5 — Adapter coverage with unit tests (Priority: P1)

**Goal**: Фасад `ListSections` / `GetSection` / `CreateSection` / `UpdateSection` / `DeleteSection` поверх codegen; unit-тесты httptest для всех TaskGroupController_*; 404 → KindNotFound; list unwraps `taskGroups`.

**Independent Test**: `go test ./internal/api/ -count=1` — happy path на list/create/get/update/delete; get 404 → `KindNotFound`; list requires parent; без cobra.

### Tests for User Story 5 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before facade implementation.

- [x] T020 [P] [US5] Add failing httptest tests: `ListSections` maps `parent` + filters (`includeRemoved`, `maxCount`, `offset`) → query + returns sections from `taskGroups` in `internal/api/section_test.go`
- [x] T021 [P] [US5] Add failing httptest tests: `GetSection` / `CreateSection` / `UpdateSection` / `DeleteSection` happy paths in `internal/api/section_test.go`
- [x] T022 [P] [US5] Add failing test: get (or update/delete) HTTP 404 → `Classify` KindNotFound with `WithEntityID` in `internal/api/section_test.go`
- [x] T023 [US5] Add failing test: `ListSections` with empty `Parent` → usage-style error before network (or assert CLI contract; facade rejects empty parent) in `internal/api/section_test.go`
- [x] T024 [P] [US5] Add failing tests: `CreateSection` body has `title`+`parent`; `UpdateSection` partial body (title only / parent only / both) in `internal/api/section_test.go`

### Implementation for User Story 5

- [x] T030 [US5] Implement `ListSections` / `GetSection` in `internal/api/section.go` via `TaskGroupController*WithResponse` + `EnsureSuccess` + `Classify`
- [x] T031 [US5] Implement `CreateSection` and `UpdateSection` (partial DTO) in `internal/api/section.go`
- [x] T032 [US5] Implement `DeleteSection` in `internal/api/section.go` (204 success)
- [x] T033 [US5] Make US5 tests green in `internal/api/section_test.go` (Authorization `Bearer test-token-…`)

### Coverage gate for US5

- [x] T034 Run `make test` after US5 (Makefile)

**Checkpoint**: All TaskGroupController_* reachable via facade + mocked unit tests (SC-002/003).

---

## Phase 4: User Story 1 — List and inspect project sections (Priority: P1) 🎯 MVP

**Goal**: `singctl project section list <PROJECT_ID>` (обязательный project id; `--removed`/`--limit`/`--offset` + валидация) и `project section get <SECTION_ID>`; рендер F06; json list=массив, get=объект; exit/streams F07.

**Independent Test**: httptest + `executeForTest`: list requires PROJECT_ID; list filters → query; list `-o json` → array; get `-o json` → object; 404 → ExitCode 3, empty stdout; bad `--limit` → ExitCode 1 without network.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

- [x] T040 [P] [US1] Add failing tests: `project section list` happy path + empty list json `[]` + filter flags mapping in `internal/cli/project_section_list_test.go` (httptest + temp config like `project_list_test.go`)
- [x] T041 [P] [US1] Add failing tests: list without `<PROJECT_ID>` / empty trim → ExitCode 1, no HTTP hit in `internal/cli/project_section_list_test.go`
- [x] T042 [P] [US1] Add failing tests: `--limit` out of 1…1000 / negative `--offset` → ExitCode 1, no HTTP hit in `internal/cli/project_section_list_test.go`
- [x] T043 [P] [US1] Add failing tests: `project section get` json single object; 404 → ExitCode 3, empty stdout in `internal/cli/project_section_get_test.go`
- [x] T044 [US1] Add failing tests: no token → ExitCode 2 for `project section list`/`get` in `internal/cli/project_section_list_test.go` or `internal/cli/project_section_auth_test.go`

### Implementation for User Story 1

- [x] T050 [US1] Add `newProjectSectionCmd()` group and register on `newProjectCmd()` in `internal/cli/project_section_cmd.go` + `internal/cli/project_cmd.go` (update Long to mention `section`)
- [x] T051 [US1] Implement Section→`output.RecordSet` mapper (stable columns per `contracts/section-output.md`) in `internal/cli/project_section_render.go`
- [x] T052 [US1] Implement `project section list` in `internal/cli/project_section_list.go` (required PROJECT_ID, flags, validation, session, `ListSections`, `Render` SingleObject=false)
- [x] T053 [US1] Implement `project section get` in `internal/cli/project_section_get.go` (`Render` SingleObject=true)
- [x] T054 [US1] Make US1 tests green; assert success stderr empty / error stdout empty where applicable

### Coverage gate for US1

- [x] T055 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP CLI — list/get usable for pipe scripts (SC-006).

---

## Phase 5: User Story 2 — Create and update sections (Priority: P1)

**Goal**: `project section create <PROJECT_ID> --title` / `project section update <SECTION_ID>` с `--title` и/или `--parent` (перенос); пустой/whitespace title → exit 1; update без флагов → exit 1; create без `--parent` flag; stdout = полная секция.

**Independent Test**: create `--title` → json object; create without title / empty title → ExitCode 1 no network; update `--parent` only → PATCH parent; update no flags → ExitCode 1; empty `--parent` when set → ExitCode 1.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [x] T060 [P] [US2] Add failing tests: `project section create` happy path + required PROJECT_ID/title in `internal/cli/project_section_create_test.go`
- [x] T061 [P] [US2] Add failing tests: create without `--title` / empty whitespace title / missing PROJECT_ID → ExitCode 1 no network in `internal/cli/project_section_create_test.go`
- [x] T062 [P] [US2] Add failing tests: `project section update` partial flags (`--title`, `--parent`, both); update with no write flags → ExitCode 1; empty title/parent when flag set → ExitCode 1; 404 → ExitCode 3 in `internal/cli/project_section_update_test.go`

### Implementation for User Story 2

- [x] T070 [US2] Implement `project section create` in `internal/cli/project_section_create.go` (required PROJECT_ID + non-empty trim title; no `--parent` flag; parent from positional arg)
- [x] T071 [US2] Implement `project section update` in `internal/cli/project_section_update.go` (require ≥1 of `--title`/`--parent` Changed; partial `SectionWriteInput`; non-empty trim when flag set)
- [x] T072 [US2] Make US2 tests green; stdout SingleObject section; streams/exit per F07

### Coverage gate for US2

- [x] T073 Run `make test` after US2 (Makefile)

**Checkpoint**: create/update closed (FR-004/005/005a + clarify).

---

## Phase 6: User Story 3 — Delete section (Priority: P1)

**Goal**: `project section delete <SECTION_ID>` → пустой stdout; 404 → exit 3; empty ID → exit 1; без confirm.

**Independent Test**: delete 204 + empty stdout; 404 → ExitCode 3; whitespace ID → ExitCode 1.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [x] T080 [P] [US3] Add failing tests: `project section delete` success empty stdout in `internal/cli/project_section_delete_test.go`
- [x] T081 [P] [US3] Add failing tests: delete 404 → ExitCode 3; empty/whitespace SECTION_ID → ExitCode 1 no network in `internal/cli/project_section_delete_test.go`

### Implementation for User Story 3

- [x] T090 [US3] Implement `project section delete` in `internal/cli/project_section_delete.go` (no render on success)
- [x] T091 [US3] Make US3 tests green

### Coverage gate for US3

- [x] T092 Run `make test` after US3 (Makefile)

**Checkpoint**: delete closed (FR-006); full CRUD lifecycle.

---

## Phase 7: User Story 4 — Discoverable CLI help (Priority: P2)

**Goal**: `project section --help` и help каждой из пяти подкоманд документируют scope F12; `project --help` упоминает `section`; `--parent` на update описан как перенос; нет `column`; термин «секция».

**Independent Test**: `executeForTest` / `--help` substring tests; unknown subcommand → ExitCode 1.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [x] T100 [P] [US4] Add failing test: `project --help` mentions `section`; `project section --help` lists `list|get|create|update|delete` in `internal/cli/project_section_help_test.go`
- [x] T101 [P] [US4] Add failing tests: each subcommand `--help` mentions key args/flags (required PROJECT_ID on list/create; `--parent` on update; `--removed`/`--limit`/`--offset` on list); help MUST NOT claim `column` as available in `internal/cli/project_section_help_test.go`
- [x] T102 [US4] Add failing test: unknown `project section` subcommand → ExitCode 1, empty stdout in `internal/cli/project_section_help_test.go`

### Implementation for User Story 4

- [x] T110 [US4] Flesh out Short/Long/Example and flag usage strings across `internal/cli/project_section_cmd.go`, `project_section_list.go`, `project_section_get.go`, `project_section_create.go`, `project_section_update.go`, `project_section_delete.go`; update `internal/cli/project_cmd.go` Long
- [x] T111 [US4] Make US4 help tests green

### Coverage gate for US4

- [x] T112 Run `make test` after US4 (Makefile)

**Checkpoint**: SC-004 help discoverability.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: сквозные проверки DoD, docs pointers, coverage.

- [x] T120 [P] Add/adjust cross-cutting exit+stream matrix sample for at least one mutate + list error path if gaps remain in `internal/cli/project_section_*_test.go`
- [x] T121 [P] Optional: note in `docs/api/coverage.md` that TaskGroupController_* are implemented via `project section` (F12) — only if project tracks coverage updates in feature PRs
- [x] T122 Confirm `internal/api/doc.go` mentions section facade (if not fully done in T012)
- [x] T123 Run quickstart checks from `specs/012-project-sections/quickstart.md` (`make test` + optional `go build` help smoke)
- [x] T124 Coverage gate: final `make test` with no regression vs Phase 2 baseline (Makefile)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: depends on Setup — BLOCKS all user stories
- **US5 (Phase 3)**: depends on Foundational — BLOCKS CLI stories that need facade (US1–US3)
- **US1 (Phase 4)**: depends on US5 — MVP
- **US2 (Phase 5)**: depends on US5 + US1 group registration (`project_section_cmd` / render); can start after T050–T051
- **US3 (Phase 6)**: depends on US5 + US1 registration
- **US4 (Phase 7)**: depends on all five subcommands existing (after US1–US3)
- **Polish (Phase 8)**: after desired stories complete

### User Story Dependencies

- **US5 (P1)**: after Foundational — no other story deps
- **US1 (P1)**: after US5 — MVP CLI
- **US2 (P1)**: after US5 (+ shared `project_section_cmd`/`project_section_render` from US1)
- **US3 (P1)**: after US5 (+ shared cmd/render from US1)
- **US4 (P2)**: after US1–US3 command files exist

### Within Each User Story

- Tests MUST fail before implementation (TDD)
- `make test` coverage MUST NOT drop at story checkpoints

### Parallel Opportunities

- T002–T004 skim in parallel
- T020–T022, T024 facade tests in parallel (same file — serialize writes carefully; or one author)
- T040–T043 CLI list/get tests in parallel (different files)
- T060–T062 create/update tests in parallel (different files)
- T080–T081 delete tests in parallel
- T100–T101 help tests in parallel
- After US5 + T050/T051: US2 and US3 can proceed in parallel by different authors

---

## Parallel Example: User Story 5

```bash
# Facade failing tests (coordinate if same section_test.go):
Task: "ListSections httptest tests in internal/api/section_test.go"
Task: "Get/Create/Update/Delete happy paths in internal/api/section_test.go"
Task: "Update partial body tests in internal/api/section_test.go"
```

---

## Parallel Example: User Story 1

```bash
Task: "project section list tests in internal/cli/project_section_list_test.go"
Task: "project section get tests in internal/cli/project_section_get_test.go"
```

---

## Implementation Strategy

### MVP First (US5 + US1)

1. Phase 1 Setup → Phase 2 Foundation (types)
2. Phase 3 US5 facade
3. Phase 4 US1 list/get
4. **STOP and VALIDATE**: `project section list` / `project section get` on mocks

### Incremental Delivery

1. US5 → API coverage DoD
2. US1 → MVP pipe read path
3. US2 → write path + `--parent` move
4. US3 → delete
5. US4 → help polish
6. Phase 8 → quickstart + final `make test`

### Suggested MVP scope

**US5 + US1** (facade + list/get) — минимальный полезный инкремент для скриптов.

---

## Notes

- [P] = different files / no incomplete deps
- Reuse `internal/output` SingleObject — do not reimplement
- No `project column` in F12
- No `--order` / `externalId` / `fake` on CLI
- Create: parent only via positional `<PROJECT_ID>`; update: optional `--parent` for move
- List: `<PROJECT_ID>` required (clarify)
- Empty/whitespace `--title` → exit 1 before network (clarify)
- Verify tests fail before implementing (TDD)
- Confirm `make test` coverage does not drop at checkpoints
- Commit after each task or logical group when asked
