# Tasks: Task CRUD (F08)

**Input**: Design documents from `/specs/008-task-crud/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Токены только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Labels: `[US5]` API task facade (все TaskController_*), `[US1]` list/get CLI, `[US2]` create/update CLI, `[US3]` archive/trash/delete CLI, `[US4]` help discoverability. Фасад (US5) — до CLI-историй: без него list/get не закрывают DoD.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]`…`[US5]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: baseline и контракты F08 перед TDD.

- [X] T001 Confirm module builds / existing tests green: `go test ./internal/api/ ./internal/output/ ./internal/cli/ ./cmd/singctl/` (or `make test`) before F08 edits
- [X] T002 [P] Skim `specs/008-task-crud/contracts/api-task-facade.md` and `specs/008-task-crud/data-model.md`
- [X] T003 [P] Skim `specs/008-task-crud/contracts/cli-task.md` and `specs/008-task-crud/contracts/task-output.md`
- [X] T004 [P] Skim `specs/008-task-crud/research.md` (deleteDate create=POST+PATCH, SingleObject, TodayLocal, columns)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: SingleObject в `internal/output`, типы Task/query/write, хелпер «сегодня» — блокируют фасад и CLI.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F08 changes
- [X] T011 [P] Add failing tests for `RenderOptions.SingleObject` (json/yaml one object; list still array; empty list `[]`; SingleObject with 0/>1 rows errors) in `internal/output/render_test.go` (and/or `render_json_test.go`)
- [X] T012 Implement `SingleObject` (or equivalent `RenderOne`) in `internal/output/model.go`, `internal/output/render.go`, `internal/output/render_json.go`, `internal/output/render_yaml.go`; make T011 green; keep table/csv one-row behavior
- [X] T013 [P] Add `TodayCalendarDate` (or `FormatCalendarDate`) helper next to `ParseDate` in `internal/api/date.go` + tests in `internal/api/date_test.go` (local `YYYY-MM-DD`)
- [X] T014 Define `Task`, `TaskListQuery`, `TaskWriteInput` types (per data-model) in `internal/api/task.go` (types only / stubs OK); update package doc in `internal/api/doc.go`
- [X] T015 Run `make test` after foundation (Makefile)

**Checkpoint**: Foundation ready — US5 (facade) can start.

---

## Phase 3: User Story 5 — Adapter coverage with unit tests (Priority: P1)

**Goal**: Фасад `ListTasks` / `GetTask` / `CreateTask` / `UpdateTask` / `DeleteTask` (+ `ArchiveTask` / `TrashTask`) поверх codegen; unit-тесты httptest для всех TaskController_*; 404 → KindNotFound.

**Independent Test**: `go test ./internal/api/ -count=1` — happy path на list/create/get/update/delete; get 404 → `KindNotFound`; create+deleteDate → POST затем PATCH; без cobra.

### Tests for User Story 5 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before facade implementation.

- [X] T020 [P] [US5] Add failing httptest tests: `ListTasks` maps filters → query + returns tasks in `internal/api/task_test.go`
- [X] T021 [P] [US5] Add failing httptest tests: `GetTask` / `CreateTask` / `UpdateTask` / `DeleteTask` happy paths in `internal/api/task_test.go`
- [X] T022 [P] [US5] Add failing test: get (or update/delete) HTTP 404 → `Classify` KindNotFound with `WithEntityID` in `internal/api/task_test.go`
- [X] T023 [US5] Add failing test: `CreateTask` with `DeleteDate` set issues POST `/v2/task` then PATCH `/v2/task/{id}` in `internal/api/task_test.go`
- [X] T024 [P] [US5] Add failing tests: `ArchiveTask` / `TrashTask` PATCH only `journalDate` / `deleteDate` in `internal/api/task_test.go`

### Implementation for User Story 5

- [X] T030 [US5] Implement `ListTasks` / `GetTask` in `internal/api/task.go` via `TaskController*WithResponse` + `EnsureSuccess` + `Classify`
- [X] T031 [US5] Implement `CreateTask` (incl. optional follow-up update for `deleteDate`) and `UpdateTask` (partial DTO) in `internal/api/task.go`
- [X] T032 [US5] Implement `DeleteTask`, `ArchiveTask`, `TrashTask` in `internal/api/task.go`
- [X] T033 [US5] Make US5 tests green in `internal/api/task_test.go` (Authorization `Bearer test-token-…`)

### Coverage gate for US5

- [X] T034 Run `make test` after US5 (Makefile)

**Checkpoint**: All TaskController_* reachable via facade + mocked unit tests (SC-002/003).

---

## Phase 4: User Story 1 — List and inspect tasks (Priority: P1) 🎯 MVP

**Goal**: `singctl task list` (фильтры + валидация limit/offset) и `singctl task get <ID>`; рендер F06; json list=массив, get=объект; exit/streams F07.

**Independent Test**: httptest + `executeForTest`: list filters → query; list `-o json` → array; get `-o json` → object; 404 → ExitCode 3, empty stdout; bad `--limit` → ExitCode 1 without network.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US1] Add failing tests: `task list` happy path + empty list json `[]` + filter flags mapping in `internal/cli/task_list_test.go` (httptest + temp config like `config_validate_test.go`)
- [X] T041 [P] [US1] Add failing tests: `--limit` out of 1…1000 / negative `--offset` → ExitCode 1, no HTTP hit in `internal/cli/task_list_test.go`
- [X] T042 [P] [US1] Add failing tests: `task get` json single object; 404 → ExitCode 3, empty stdout in `internal/cli/task_get_test.go`
- [X] T043 [US1] Add failing tests: no token → ExitCode 2 for `task list`/`get` in `internal/cli/task_list_test.go` or shared `internal/cli/task_auth_test.go`

### Implementation for User Story 1

- [X] T050 [US1] Add `newTaskCmd()` group and register on root in `internal/cli/task_cmd.go` + `internal/cli/root.go`
- [X] T051 [US1] Implement Task→`output.RecordSet` mapper (stable columns per `contracts/task-output.md`) in `internal/cli/task_render.go`
- [X] T052 [US1] Implement `task list` in `internal/cli/task_list.go` (flags, validation, session, `ListTasks`, `Render` SingleObject=false)
- [X] T053 [US1] Implement `task get` in `internal/cli/task_get.go` (`Render` SingleObject=true)
- [X] T054 [US1] Make US1 tests green; assert success stderr empty / error stdout empty where applicable

### Coverage gate for US1

- [X] T055 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP CLI — list/get usable for pipe scripts (SC-006).

---

## Phase 5: User Story 2 — Create and update tasks (Priority: P1)

**Goal**: `task create` / `task update` с флагами ТЗ + `--project`/`--parent`; `--note` as-is; update без флагов → exit 1; stdout = полная задача.

**Independent Test**: create `--title` → json object; create without title → ExitCode 1 no network; update one flag → PATCH partial; update no flags → ExitCode 1; bad priority → ExitCode 1.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US2] Add failing tests: `task create --title` happy path + optional flags in `internal/cli/task_create_test.go`
- [X] T061 [P] [US2] Add failing tests: create without `--title`; invalid `--priority`; create `--delete-date` triggers POST+PATCH (assert via mock) in `internal/cli/task_create_test.go`
- [X] T062 [P] [US2] Add failing tests: `task update` partial flags; update with no write flags → ExitCode 1; 404 → ExitCode 3 in `internal/cli/task_update_test.go`

### Implementation for User Story 2

- [X] T070 [US2] Implement `task create` in `internal/cli/task_create.go` (required title, write flags, help note on delta for `--note`)
- [X] T071 [US2] Implement `task update` in `internal/cli/task_update.go` (require ≥1 write flag; partial `TaskWriteInput`)
- [X] T072 [US2] Make US2 tests green; stdout SingleObject task; streams/exit per F07

### Coverage gate for US2

- [X] T073 Run `make test` after US2 (Makefile)

**Checkpoint**: create/update closed.

---

## Phase 6: User Story 3 — Archive, trash, and permanent delete (Priority: P1)

**Goal**: `archive` / `trash` (`--date` или TodayLocal) → полная задача в stdout; `delete` → пустой stdout; invalid date → exit 1.

**Independent Test**: archive/trash mock PATCH dates; default date = TodayLocal; delete 204 + empty stdout; bad `--date` → ExitCode 1.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US3] Add failing tests: `task archive` / `task trash` with `--date` and without (TodayLocal) in `internal/cli/task_archive_test.go` / `internal/cli/task_trash_test.go`
- [X] T081 [P] [US3] Add failing tests: invalid `--date` → ExitCode 1 no network in archive/trash tests
- [X] T082 [US3] Add failing tests: `task delete` success empty stdout; 404 → ExitCode 3 in `internal/cli/task_delete_test.go`

### Implementation for User Story 3

- [X] T090 [US3] Implement `task archive` in `internal/cli/task_archive.go` and `task trash` in `internal/cli/task_trash.go`
- [X] T091 [US3] Implement `task delete` in `internal/cli/task_delete.go` (no render on success)
- [X] T092 [US3] Make US3 tests green

### Coverage gate for US3

- [X] T093 Run `make test` after US3 (Makefile)

**Checkpoint**: archive/trash/delete closed (FR-006/007/008).

---

## Phase 7: User Story 4 — Discoverable CLI help (Priority: P2)

**Goal**: `task --help` и help каждой из семи подкоманд документируют scope F08; нет checklist/kanban/`move`; `--note` упоминает delta.

**Independent Test**: `executeForTest` / `--help` substring tests; unknown subcommand → ExitCode 1.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T100 [P] [US4] Add failing test: `task --help` lists `list|get|create|update|delete|archive|trash` in `internal/cli/task_help_test.go`
- [X] T101 [P] [US4] Add failing tests: each subcommand `--help` mentions key flags; create/update help mentions delta for `--note`; help MUST NOT claim checklist/kanban/move as available in `internal/cli/task_help_test.go`
- [X] T102 [US4] Add failing test: unknown `task` subcommand → ExitCode 1, empty stdout in `internal/cli/task_help_test.go`

### Implementation for User Story 4

- [X] T110 [US4] Flesh out Short/Long/Example and flag usage strings across `internal/cli/task_cmd.go`, `task_list.go`, `task_get.go`, `task_create.go`, `task_update.go`, `task_delete.go`, `task_archive.go`, `task_trash.go`
- [X] T111 [US4] Make US4 help tests green

### Coverage gate for US4

- [X] T112 Run `make test` after US4 (Makefile)

**Checkpoint**: SC-004 help discoverability.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: сквозные проверки DoD, docs pointers, coverage.

- [X] T120 [P] Add/adjust cross-cutting exit+stream matrix sample for at least one mutate + list error path if gaps remain in `internal/cli/task_*_test.go`
- [X] T121 [P] Optional: one-line pointer from `docs/scriptability.md` that entity pipe examples now apply to `task list`/`create` (F08) — keep honest if only mock-proven
- [X] T122 Update `internal/api/doc.go` to mention task facade if not done in T014
- [X] T123 Run quickstart checks from `specs/008-task-crud/quickstart.md` (`make test` + optional `go build` help smoke)
- [X] T124 Coverage gate: final `make test` with no regression vs Phase 2 baseline (Makefile)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: after Setup — BLOCKS all stories
- **US5 (Phase 3)**: after Foundational — BLOCKS US1–US4 CLI work that needs facade
- **US1 (Phase 4)**: after US5 (needs List/Get) — product MVP
- **US2 (Phase 5)**: after US5 (Create/Update); ideally after US1 for shared render/cmd group
- **US3 (Phase 6)**: after US5 (Archive/Trash/Delete); shares CLI patterns with US1/US2
- **US4 (Phase 7)**: after commands exist (after US1 minimum; best after US2–US3 so all seven present)
- **Polish (Phase 8)**: after desired stories complete

### User Story Dependencies

- **US5**: after Foundational only
- **US1**: after US5 (List/Get)
- **US2**: after US5; shares `task_cmd`/`task_render` with US1
- **US3**: after US5; shares CLI harness with US1/US2
- **US4**: after US1–US3 command files exist (can draft help earlier, assert completeness last)

### Within Each User Story

- Tests MUST fail before implementation (TDD)
- `make test` coverage MUST NOT drop at story checkpoints
- Story complete before next priority when sequential

### Parallel Opportunities

- T002–T004 skim in parallel
- T011 vs T013 in parallel (different packages) before T012/T014 sequencing as needed
- T020–T024 US5 tests in parallel after types exist
- T040–T043 US1 tests in parallel
- T060–T062 US2 tests in parallel
- T080–T081 US3 tests in parallel
- T100–T101 US4 tests in parallel
- After US5: US1 then US2/US3 can be sequential; two devs could split US2 vs US3 after US1 registers `task` group

---

## Parallel Example: User Story 5

```bash
# Launch facade tests together (REQUIRED — before implementation):
Task: "ListTasks httptest tests in internal/api/task_test.go"
Task: "Get/Create/Update/Delete happy paths in internal/api/task_test.go"
Task: "404 KindNotFound in internal/api/task_test.go"
Task: "Archive/Trash PATCH tests in internal/api/task_test.go"
```

## Parallel Example: User Story 1

```bash
# Launch CLI tests together:
Task: "task list filters/empty json in internal/cli/task_list_test.go"
Task: "limit/offset validation in internal/cli/task_list_test.go"
Task: "task get json object + 404 in internal/cli/task_get_test.go"
```

---

## Implementation Strategy

### MVP First (US5 + US1)

1. Phase 1 Setup + Phase 2 Foundational (SingleObject + types)
2. Phase 3 US5 facade (all operations mocked)
3. Phase 4 US1 list/get CLI
4. **STOP and VALIDATE**: list/get Independent Test + `make test`
5. Demo: `task list -o json` / `task get` against mock or live later (F33)

### Incremental Delivery

1. Foundation → US5 → API layer done
2. US1 → read CLI MVP
3. US2 → write path
4. US3 → archive/trash/delete
5. US4 → help polish
6. Phase 8 → quickstart + final coverage

### Parallel Team Strategy

1. Together: Setup + Foundational + US5
2. Then: Dev A US1 → US4 help; Dev B US2; Dev C US3 (after `task_cmd`/`task_render` from US1)

---

## Notes

- [P] = different files, no incomplete dependencies
- Do not hand-edit `internal/apiclient/`; use codegen methods only
- No checklist/kanban/`move` / alias `t` in F08
- Prefer `executeForTest` + httptest patterns from `internal/cli/config_validate_test.go`
- Commit after each task or logical group
- Suggested next command: `/speckit-implement`
