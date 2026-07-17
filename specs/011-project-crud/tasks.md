# Tasks: Project CRUD (F11)

**Input**: Design documents from `/specs/011-project-crud/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Токены только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Labels: `[US5]` API project facade (все ProjectController_*), `[US1]` list/get CLI, `[US2]` create/update CLI (+ emoji), `[US3]` archive/trash/delete CLI, `[US4]` help discoverability. Фасад (US5) — до CLI-историй: без него list/get не закрывают DoD. `SingleObject` и `TodayCalendarDate` уже есть (F08) — в foundation не пересоздавать.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]`…`[US5]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: baseline и контракты F11 перед TDD.

- [X] T001 Confirm module builds / existing tests green: `go test ./internal/api/ ./internal/output/ ./internal/cli/ ./cmd/singctl/` (or `make test`) before F11 edits
- [X] T002 [P] Skim `specs/011-project-crud/contracts/api-project-facade.md` and `specs/011-project-crud/data-model.md`
- [X] T003 [P] Skim `specs/011-project-crud/contracts/cli-project.md` and `specs/011-project-crud/contracts/project-output.md`
- [X] T004 [P] Skim `specs/011-project-crud/research.md` (create unwrap project, emoji normalize, archive/trash, columns)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: emoji helper + типы Project/query/write — блокируют фасад и CLI. Output SingleObject / TodayCalendarDate — reuse only.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F11 changes
- [X] T011 [P] Add failing tests for `NormalizeProjectEmoji` (hex pass-through lowercase; unicode→hex e.g. 💞→1f49e; reject empty/multi-rune/ASCII word) in `internal/api/emoji_test.go`
- [X] T012 Implement `NormalizeProjectEmoji` in `internal/api/emoji.go`; make T011 green (stdlib only; per `research.md` §4)
- [X] T013 Define `Project`, `ProjectListQuery`, `ProjectWriteInput` types (per data-model) in `internal/api/project.go` (types only / stubs OK); update package doc in `internal/api/doc.go` to mention ProjectController facade
- [X] T014 Run `make test` after foundation (Makefile)

**Checkpoint**: Foundation ready — US5 (facade) can start.

---

## Phase 3: User Story 5 — Adapter coverage with unit tests (Priority: P1)

**Goal**: Фасад `ListProjects` / `GetProject` / `CreateProject` / `UpdateProject` / `DeleteProject` (+ `ArchiveProject` / `TrashProject`) поверх codegen; unit-тесты httptest для всех ProjectController_*; 404 → KindNotFound; create unwraps `.Project` only.

**Independent Test**: `go test ./internal/api/ -count=1` — happy path на list/create/get/update/delete; get 404 → `KindNotFound`; create mock with taskGroup ignored; archive/trash PATCH dates; без cobra.

### Tests for User Story 5 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before facade implementation.

- [X] T020 [P] [US5] Add failing httptest tests: `ListProjects` maps filters → query + returns projects in `internal/api/project_test.go`
- [X] T021 [P] [US5] Add failing httptest tests: `GetProject` / `CreateProject` / `UpdateProject` / `DeleteProject` happy paths in `internal/api/project_test.go`
- [X] T022 [P] [US5] Add failing test: get (or update/delete) HTTP 404 → `Classify` KindNotFound with `WithEntityID` in `internal/api/project_test.go`
- [X] T023 [US5] Add failing test: `CreateProject` unwraps `project` from `ProjectCreateResponseDto` and ignores `taskGroup` in `internal/api/project_test.go`
- [X] T024 [P] [US5] Add failing tests: `ArchiveProject` / `TrashProject` PATCH only `journalDate` / `deleteDate` in `internal/api/project_test.go`

### Implementation for User Story 5

- [X] T030 [US5] Implement `ListProjects` / `GetProject` in `internal/api/project.go` via `ProjectController*WithResponse` + `EnsureSuccess` + `Classify`
- [X] T031 [US5] Implement `CreateProject` (map `.JSON200.Project` only) and `UpdateProject` (partial DTO) in `internal/api/project.go`
- [X] T032 [US5] Implement `DeleteProject`, `ArchiveProject`, `TrashProject` in `internal/api/project.go`
- [X] T033 [US5] Make US5 tests green in `internal/api/project_test.go` (Authorization `Bearer test-token-…`)

### Coverage gate for US5

- [X] T034 Run `make test` after US5 (Makefile)

**Checkpoint**: All ProjectController_* reachable via facade + mocked unit tests (SC-002/003).

---

## Phase 4: User Story 1 — List and inspect projects (Priority: P1) 🎯 MVP

**Goal**: `singctl project list` (фильтры `--archived`/`--removed`/`--limit`/`--offset` + валидация) и `singctl project get <ID>`; рендер F06; json list=массив, get=объект; exit/streams F07.

**Independent Test**: httptest + `executeForTest`: list filters → query; list `-o json` → array; get `-o json` → object; 404 → ExitCode 3, empty stdout; bad `--limit` → ExitCode 1 without network.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US1] Add failing tests: `project list` happy path + empty list json `[]` + filter flags mapping in `internal/cli/project_list_test.go` (httptest + temp config like `task_list_test.go`)
- [X] T041 [P] [US1] Add failing tests: `--limit` out of 1…1000 / negative `--offset` → ExitCode 1, no HTTP hit in `internal/cli/project_list_test.go`
- [X] T042 [P] [US1] Add failing tests: `project get` json single object; 404 → ExitCode 3, empty stdout in `internal/cli/project_get_test.go`
- [X] T043 [US1] Add failing tests: no token → ExitCode 2 for `project list`/`get` in `internal/cli/project_list_test.go` or shared `internal/cli/project_auth_test.go`

### Implementation for User Story 1

- [X] T050 [US1] Add `newProjectCmd()` group and register on root in `internal/cli/project_cmd.go` + `internal/cli/root.go`
- [X] T051 [US1] Implement Project→`output.RecordSet` mapper (stable columns per `contracts/project-output.md`) in `internal/cli/project_render.go`
- [X] T052 [US1] Implement `project list` in `internal/cli/project_list.go` (flags, validation, session, `ListProjects`, `Render` SingleObject=false)
- [X] T053 [US1] Implement `project get` in `internal/cli/project_get.go` (`Render` SingleObject=true)
- [X] T054 [US1] Make US1 tests green; assert success stderr empty / error stdout empty where applicable

### Coverage gate for US1

- [X] T055 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP CLI — list/get usable for pipe scripts (SC-006).

---

## Phase 5: User Story 2 — Create and update projects (Priority: P1)

**Goal**: `project create` / `project update` с флагами ТЗ §6.2 + `--parent`; `--note` as-is; `--emoji` через `NormalizeProjectEmoji`; update без флагов → exit 1; stdout = полный проект.

**Independent Test**: create `--title` → json object; create without title → ExitCode 1 no network; create `--emoji 💞` → body emoji `1f49e`; update one flag → PATCH partial; update no flags → ExitCode 1; bad emoji → ExitCode 1.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US2] Add failing tests: `project create --title` happy path + optional flags (`--note`, `--notebook`, `--color`, `--parent`) in `internal/cli/project_create_test.go`
- [X] T061 [P] [US2] Add failing tests: create without `--title`; `--emoji` unicode→hex in request body; invalid emoji → ExitCode 1 no network in `internal/cli/project_create_test.go`
- [X] T062 [P] [US2] Add failing tests: `project update` partial flags (incl. `--parent`); update with no write flags → ExitCode 1; 404 → ExitCode 3 in `internal/cli/project_update_test.go`

### Implementation for User Story 2

- [X] T070 [US2] Implement `project create` in `internal/cli/project_create.go` (required title, write flags, `NormalizeProjectEmoji`, help note on delta for `--note`)
- [X] T071 [US2] Implement `project update` in `internal/cli/project_update.go` (require ≥1 write flag; partial `ProjectWriteInput`; emoji normalize when set)
- [X] T072 [US2] Make US2 tests green; stdout SingleObject project; streams/exit per F07

### Coverage gate for US2

- [X] T073 Run `make test` after US2 (Makefile)

**Checkpoint**: create/update closed (FR-004/005/004b).

---

## Phase 6: User Story 3 — Archive, trash, and permanent delete (Priority: P1)

**Goal**: `archive` / `trash` (`--date` или TodayCalendarDate) → полный проект в stdout; `delete` → пустой stdout; invalid date → exit 1. Без `--archive-date`/`--delete-date` на create/update.

**Independent Test**: archive/trash mock PATCH dates; default date = TodayCalendarDate; delete 204 + empty stdout; bad `--date` → ExitCode 1.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US3] Add failing tests: `project archive` / `project trash` with `--date` and without (TodayCalendarDate) in `internal/cli/project_archive_test.go` / `internal/cli/project_trash_test.go`
- [X] T081 [P] [US3] Add failing tests: invalid `--date` → ExitCode 1 no network in archive/trash tests
- [X] T082 [US3] Add failing tests: `project delete` success empty stdout; 404 → ExitCode 3 in `internal/cli/project_delete_test.go`

### Implementation for User Story 3

- [X] T090 [US3] Implement `project archive` in `internal/cli/project_archive.go` and `project trash` in `internal/cli/project_trash.go`
- [X] T091 [US3] Implement `project delete` in `internal/cli/project_delete.go` (no render on success)
- [X] T092 [US3] Make US3 tests green

### Coverage gate for US3

- [X] T093 Run `make test` after US3 (Makefile)

**Checkpoint**: archive/trash/delete closed (FR-006/006a/006b).

---

## Phase 7: User Story 4 — Discoverable CLI help (Priority: P2)

**Goal**: `project --help` и help каждой из семи подкоманд документируют scope F11; нет `section`/`column`; `--note` упоминает delta; `--emoji` — unicode+hex примеры; shared projects не обещать.

**Independent Test**: `executeForTest` / `--help` substring tests; unknown subcommand → ExitCode 1.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T100 [P] [US4] Add failing test: `project --help` lists `list|get|create|update|delete|archive|trash` in `internal/cli/project_help_test.go`
- [X] T101 [P] [US4] Add failing tests: each subcommand `--help` mentions key flags; create/update help mentions delta for `--note` and emoji examples; help MUST NOT claim section/column as available; MUST NOT promise shared projects in `internal/cli/project_help_test.go`
- [X] T102 [US4] Add failing test: unknown `project` subcommand → ExitCode 1, empty stdout in `internal/cli/project_help_test.go`

### Implementation for User Story 4

- [X] T110 [US4] Flesh out Short/Long/Example and flag usage strings across `internal/cli/project_cmd.go`, `project_list.go`, `project_get.go`, `project_create.go`, `project_update.go`, `project_delete.go`, `project_archive.go`, `project_trash.go`
- [X] T111 [US4] Make US4 help tests green

### Coverage gate for US4

- [X] T112 Run `make test` after US4 (Makefile)

**Checkpoint**: SC-004 help discoverability.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: сквозные проверки DoD, docs pointers, coverage.

- [X] T120 [P] Add/adjust cross-cutting exit+stream matrix sample for at least one mutate + list error path if gaps remain in `internal/cli/project_*_test.go`
- [X] T121 [P] Optional: one-line pointer from `docs/scriptability.md` that entity pipe examples also apply to `project list`/`create` (F11) — keep honest if only mock-proven
- [X] T122 Confirm `internal/api/doc.go` mentions project facade (if not fully done in T013)
- [X] T123 Run quickstart checks from `specs/011-project-crud/quickstart.md` (`make test` + optional `go build` help smoke)
- [X] T124 Coverage gate: final `make test` with no regression vs Phase 2 baseline (Makefile)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: depends on Setup — BLOCKS all user stories
- **US5 (Phase 3)**: depends on Foundational — BLOCKS CLI stories that need facade (US1–US3)
- **US1 (Phase 4)**: depends on US5 — MVP
- **US2 (Phase 5)**: depends on US5 + US1 group registration (`project_cmd` / render); can start after T050–T051
- **US3 (Phase 6)**: depends on US5 + US1 registration
- **US4 (Phase 7)**: depends on all seven subcommands existing (after US1–US3)
- **Polish (Phase 8)**: after desired stories complete

### User Story Dependencies

- **US5 (P1)**: after Foundational — no other story deps
- **US1 (P1)**: after US5 — MVP CLI
- **US2 (P1)**: after US5 (+ shared `project_cmd`/`project_render` from US1)
- **US3 (P1)**: after US5 (+ shared cmd/render from US1)
- **US4 (P2)**: after US1–US3 command files exist

### Within Each User Story

- Tests MUST fail before implementation (TDD)
- `make test` coverage MUST NOT drop at story checkpoints

### Parallel Opportunities

- T002–T004 skim in parallel
- T011 emoji tests parallel with later type stub prep after T010
- T020–T022, T024 facade tests in parallel (same file — serialize writes carefully; or one author)
- T040–T042 CLI list/get tests in parallel (different files)
- T060–T062 create/update tests in parallel (different files)
- T080–T081 archive/trash tests in parallel
- T100–T101 help tests in parallel
- After US5 + T050/T051: US2 and US3 can proceed in parallel by different authors

---

## Parallel Example: User Story 5

```bash
# Facade failing tests (coordinate if same project_test.go):
Task: "ListProjects httptest tests in internal/api/project_test.go"
Task: "Get/Create/Update/Delete happy paths in internal/api/project_test.go"
Task: "Archive/Trash PATCH tests in internal/api/project_test.go"
```

---

## Parallel Example: User Story 1

```bash
Task: "project list tests in internal/cli/project_list_test.go"
Task: "project get tests in internal/cli/project_get_test.go"
```

---

## Implementation Strategy

### MVP First (US5 + US1)

1. Phase 1 Setup → Phase 2 Foundation (emoji + types)
2. Phase 3 US5 facade
3. Phase 4 US1 list/get
4. **STOP and VALIDATE**: `project list` / `project get` on mocks

### Incremental Delivery

1. US5 → API coverage DoD
2. US1 → MVP pipe read path
3. US2 → write path + emoji
4. US3 → archive/trash/delete
5. US4 → help polish
6. Phase 8 → quickstart + final `make test`

### Suggested MVP scope

**US5 + US1** (facade + list/get) — минимальный полезный инкремент для скриптов.

---

## Notes

- [P] = different files / no incomplete deps
- Reuse `internal/output` SingleObject and `api.TodayCalendarDate` — do not reimplement
- No `project section` / `column` in F11
- No `--archive-date` / `--delete-date` on create/update
- Create response: project only (ignore taskGroup)
- Verify tests fail before implementing (TDD)
- Confirm `make test` coverage does not drop at checkpoints
- Commit after each task or logical group when asked
