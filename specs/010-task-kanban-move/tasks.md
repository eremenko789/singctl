# Tasks: Task Kanban Link & Move (F10)

**Input**: Design documents from `/specs/010-task-kanban-move/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Токены только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Labels: `[US6]` API kanban facade (все KanbanTaskStatusController_* + `MoveTaskToKanban`), `[US1]` list/get CLI, `[US2]` create/update CLI, `[US3]` move CLI, `[US4]` delete CLI, `[US5]` help discoverability. Фасад (US6) — до CLI-историй. Зависит от F08 (`GetTask`, output SingleObject, `task` group). F09 checklist соседний, не блокирует.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]`…`[US6]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: baseline и контракты F10 перед TDD.

- [X] T001 Confirm F08/module builds green before F10 edits: `go test ./internal/api/ ./internal/output/ ./internal/cli/ ./cmd/singctl/` (or `make test`)
- [X] T002 [P] Skim `specs/010-task-kanban-move/contracts/api-kanban-facade.md` and `specs/010-task-kanban-move/data-model.md`
- [X] T003 [P] Skim `specs/010-task-kanban-move/contracts/cli-kanban.md` and `specs/010-task-kanban-move/contracts/kanban-output.md`
- [X] T004 [P] Skim `specs/010-task-kanban-move/research.md` (Move in facade; list no pre-check; create no uniqueness; no `--order` on move; same-column still update)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: типы KanbanLink/query/write и подтверждение F08-зависимостей — блокируют фасад и CLI. SingleObject уже есть из F08 (не дублировать).

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F10 changes
- [X] T011 Confirm `Session.GetTask` exists in `internal/api/task.go` and `RenderOptions.SingleObject` works in `internal/output/` (smoke via existing F08 tests; no new output APIs)
- [X] T012 Define `KanbanLink`, `KanbanLinkListQuery`, `KanbanLinkWriteInput` types (per `specs/010-task-kanban-move/data-model.md`) in `internal/api/kanban_task_status.go` (types only / stubs OK); update package doc in `internal/api/doc.go`
- [X] T013 Run `make test` after foundation (Makefile)

**Checkpoint**: Foundation ready — US6 (facade) can start.

---

## Phase 3: User Story 6 — Adapter coverage with unit tests (Priority: P1)

**Goal**: Фасад `ListKanbanLinks` / `GetKanbanLink` / `CreateKanbanLink` / `UpdateKanbanLink` / `DeleteKanbanLink` + `MoveTaskToKanban` поверх codegen; unit-тесты httptest для всех KanbanTaskStatusController_*; 404 → KindNotFound; move 0/1/>1. Фасад **не** вызывает GetTask.

**Independent Test**: `go test ./internal/api/ -count=1` — happy path list/create/get/update/delete; get 404 → `KindNotFound`; list query только taskId/statusId когда заданы; Move 0→create, 1→update statusId only (в т.ч. same column), >1→KindValidation (или эквивалент); без cobra.

### Tests for User Story 6 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before facade implementation.

- [X] T020 [P] [US6] Add failing httptest test: `ListKanbanLinks` maps optional TaskId/StatusId only → query + returns links from `kanbanTaskStatuses` in `internal/api/kanban_task_status_test.go`
- [X] T021 [P] [US6] Add failing httptest tests: `GetKanbanLink` / `CreateKanbanLink` / `UpdateKanbanLink` / `DeleteKanbanLink` happy paths in `internal/api/kanban_task_status_test.go`
- [X] T022 [P] [US6] Add failing test: get (or update/delete) HTTP 404 → `Classify` KindNotFound with `WithEntityID` in `internal/api/kanban_task_status_test.go`
- [X] T023 [US6] Add failing tests: create body has taskId+statusId (optional kanbanOrder); update sends only set fields; no externalId; create does not list-first for uniqueness in `internal/api/kanban_task_status_test.go`
- [X] T024 [US6] Add failing tests: `MoveTaskToKanban` 0 links → POST create (no kanbanOrder); 1 link → PATCH statusId only (even if same column); >1 links → error, no write in `internal/api/kanban_task_status_test.go`

### Implementation for User Story 6

- [X] T030 [US6] Implement `ListKanbanLinks` / `GetKanbanLink` in `internal/api/kanban_task_status.go` via `KanbanTaskStatusController*WithResponse` + `EnsureSuccess` + `Classify`
- [X] T031 [US6] Implement `CreateKanbanLink` and `UpdateKanbanLink` (partial DTO; optional order; no client uniqueness check) in `internal/api/kanban_task_status.go`
- [X] T032 [US6] Implement `DeleteKanbanLink` in `internal/api/kanban_task_status.go` (204 success)
- [X] T033 [US6] Implement `MoveTaskToKanban` (list→create|update statusId|ambiguous KindValidation) in `internal/api/kanban_task_status.go`
- [X] T034 [US6] Make US6 tests green in `internal/api/kanban_task_status_test.go` (Authorization `Bearer test-token-…`)

### Coverage gate for US6

- [X] T035 Run `make test` after US6 (Makefile)

**Checkpoint**: All KanbanTaskStatusController_* + Move reachable via facade + mocked unit tests (SC-002/003).

---

## Phase 4: User Story 1 — List and inspect kanban links (Priority: P1) 🎯 MVP

**Goal**: `singctl task kanban list` (опциональные `--task`/`--status`, **без** pre-check) и `task kanban get <LINK_ID>`; рендер F06; json list=массив, get=объект; exit/streams F07.

**Independent Test**: httptest + `executeForTest`: list → json array; filters in query; get json object; link 404 → ExitCode 3; empty list → `[]`; with `--task` assert **no** `GET /v2/task/` call.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US1] Add failing tests: `task kanban list` happy path + empty list json `[]` + optional `--task`/`--status` query mapping; assert no GetTask in `internal/cli/task_kanban_list_test.go`
- [X] T041 [P] [US1] Add failing tests: `task kanban get` json single object; link 404 → ExitCode 3, empty stdout in `internal/cli/task_kanban_get_test.go`
- [X] T042 [US1] Add failing tests: no token → ExitCode 2; empty/whitespace link id → ExitCode 1 in list/get tests or `internal/cli/task_kanban_auth_test.go`

### Implementation for User Story 1

- [X] T050 [US1] Add `newTaskKanbanCmd()` group and register under `newTaskCmd()` in `internal/cli/task_kanban_cmd.go` + update `internal/cli/task_cmd.go` (remove «kanban и move недоступны» from Long; mention `kanban`)
- [X] T051 [US1] Implement KanbanLink→`output.RecordSet` mapper (columns per `contracts/kanban-output.md`) in `internal/cli/task_kanban_render.go`
- [X] T052 [US1] Implement `task kanban list` in `internal/cli/task_kanban_list.go` (optional filters → `ListKanbanLinks` → `Render` SingleObject=false; no GetTask)
- [X] T053 [US1] Implement `task kanban get` in `internal/cli/task_kanban_get.go` (`Render` SingleObject=true)
- [X] T054 [US1] Make US1 tests green; assert success stderr empty / error stdout empty where applicable

### Coverage gate for US1

- [X] T055 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP CLI — list/get usable for pipe scripts (SC-006 read path).

---

## Phase 5: User Story 2 — Create and update kanban links (Priority: P1)

**Goal**: `task kanban create` / `update` с `--task` / `--column` / `--order`; pre-check на create; без клиентской уникальности; update без флагов → exit 1; stdout = полная связь.

**Independent Test**: create → json object after GetTask; missing flags → ExitCode 1 no network; task 404 → ExitCode 3 no create; update `--column`; update no flags → ExitCode 1; create with existing link still POSTs (httptest).

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US2] Add failing tests: `task kanban create --task --column` happy path + optional `--order`; assert GetTask then POST create in `internal/cli/task_kanban_create_test.go`
- [X] T061 [P] [US2] Add failing tests: create missing flags; whitespace ids; task 404 → ExitCode 3 no create; negative `--order` → ExitCode 1 in `internal/cli/task_kanban_create_test.go`
- [X] T062 [P] [US2] Add failing tests: `task kanban update` with `--task`/`--column`/`--order`; no write flags → ExitCode 1; link 404 → ExitCode 3 in `internal/cli/task_kanban_update_test.go`

### Implementation for User Story 2

- [X] T070 [US2] Implement `task kanban create` in `internal/cli/task_kanban_create.go` (validate → GetTask → `CreateKanbanLink`; SingleObject stdout; no uniqueness list)
- [X] T071 [US2] Implement `task kanban update` in `internal/cli/task_kanban_update.go` (≥1 write flag; partial write; no task pre-check)
- [X] T072 [US2] Make US2 tests green; streams/exit per F07

### Coverage gate for US2

- [X] T073 Run `make test` after US2 (Makefile)

**Checkpoint**: create/update closed.

---

## Phase 6: User Story 3 — Move task between columns (Priority: P1)

**Goal**: `singctl task move <TASK_ID> --column <COLUMN_ID>`: GetTask → `MoveTaskToKanban`; без `--order`; stdout = полная связь; >1 связей → exit 1; intermediate list не в stdout.

**Independent Test**: 0 links → create path; 1 link → update path (incl. same column); >1 → ExitCode 1; task 404 → ExitCode 3; missing `--column` → ExitCode 1; no `--order` flag in help/parse.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US3] Add failing tests: `task move` 0 links → create after GetTask; 1 link → update; same-column still update; assert list not on stdout in `internal/cli/task_move_test.go`
- [X] T081 [P] [US3] Add failing tests: >1 links → ExitCode 1 no write; task 404 → ExitCode 3; missing `--column`/TASK_ID → ExitCode 1; reject `--order` in `internal/cli/task_move_test.go`

### Implementation for User Story 3

- [X] T090 [US3] Implement `task move` in `internal/cli/task_move.go` and register in `internal/cli/task_cmd.go` (validate → GetTask → `MoveTaskToKanban` → SingleObject render)
- [X] T091 [US3] Make US3 tests green; map ambiguous move error → ExitCode 1 + stderr hint

### Coverage gate for US3

- [X] T092 Run `make test` after US3 (Makefile)

**Checkpoint**: move UX closed (FR-007, SC-001 move path).

---

## Phase 7: User Story 4 — Delete kanban links (Priority: P1)

**Goal**: `task kanban delete <LINK_ID>` → пустой stdout; 404 → exit 3; empty id → exit 1; без confirm.

**Independent Test**: delete 204 + empty stdout; 404 → ExitCode 3; whitespace id → ExitCode 1 no network.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T100 [P] [US4] Add failing tests: `task kanban delete` success empty stdout; 404 → ExitCode 3 in `internal/cli/task_kanban_delete_test.go`
- [X] T101 [US4] Add failing tests: empty/whitespace id → ExitCode 1 no HTTP in `internal/cli/task_kanban_delete_test.go`

### Implementation for User Story 4

- [X] T110 [US4] Implement `task kanban delete` in `internal/cli/task_kanban_delete.go` (no render on success)
- [X] T111 [US4] Make US4 tests green

### Coverage gate for US4

- [X] T112 Run `make test` after US4 (Makefile)

**Checkpoint**: delete closed (FR-006).

---

## Phase 8: User Story 5 — Discoverable CLI help (Priority: P2)

**Goal**: `task kanban --help` и help пяти подкоманд + `task move --help` документируют scope F10; `task --help` показывает `kanban` и `move`; нет TUI / column CRUD / `--order` на move / pagination promises.

**Independent Test**: `executeForTest` / `--help` substring tests; unknown kanban subcommand → ExitCode 1.

### Tests for User Story 5 (REQUIRED — TDD) ⚠️

- [X] T120 [P] [US5] Add failing test: `task --help` lists `kanban` and `move`; `task kanban --help` lists `list|get|create|update|delete` in `internal/cli/task_kanban_help_test.go`
- [X] T121 [P] [US5] Add failing tests: each kanban subcommand + `task move --help` mention key flags; move MUST NOT document `--order`; MUST NOT claim TUI / `project column` / pagination in `internal/cli/task_kanban_help_test.go` (and/or `task_move_test.go` help cases)
- [X] T122 [US5] Add failing test: unknown `task kanban` subcommand → ExitCode 1, empty stdout in `internal/cli/task_kanban_help_test.go`

### Implementation for User Story 5

- [X] T130 [US5] Flesh out Short/Long/Example and flag usage across `internal/cli/task_kanban_cmd.go`, `task_kanban_list.go`, `task_kanban_get.go`, `task_kanban_create.go`, `task_kanban_update.go`, `task_kanban_delete.go`, `task_move.go`; sync `task_cmd.go` help text
- [X] T131 [US5] Make US5 help tests green

### Coverage gate for US5

- [X] T132 Run `make test` after US5 (Makefile)

**Checkpoint**: SC-004 help discoverability.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: сквозные проверки DoD, docs pointers, coverage.

- [X] T140 [P] Add/adjust cross-cutting exit+stream matrix sample for create/move pre-check failure + ambiguous move if gaps remain in `internal/cli/task_kanban_*_test.go` / `task_move_test.go`
- [X] T141 [P] Optional: one-line pointer in `docs/scriptability.md` that pipe examples apply to `task kanban list` / `task move` (F10) — keep honest if only mock-proven
- [X] T142 Confirm `internal/api/doc.go` mentions kanban-task-status facade + Move
- [X] T143 Run quickstart checks from `specs/010-task-kanban-move/quickstart.md` (`make test` + optional `go build` help smoke)
- [X] T144 Coverage gate: final `make test` with no regression vs Phase 2 baseline (Makefile)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: depends on Setup — BLOCKS all user stories
- **US6 (Phase 3)**: depends on Foundation — BLOCKS CLI stories that call facade
- **US1 (Phase 4)**: depends on US6 — MVP CLI; registers `task kanban` group
- **US2 (Phase 5)**: depends on US6 + preferably US1 (shared render/group)
- **US3 (Phase 6)**: depends on US6 (`MoveTaskToKanban`) + preferably US1 (shared render); registers `task move`
- **US4 (Phase 7)**: depends on US6 + preferably US1 (shared group)
- **US5 (Phase 8)**: depends on commands existing (after US1–US4 ideally)
- **Polish (Phase 9)**: after desired stories complete

### User Story Dependencies

- **US6 (P1)**: after Foundation — no other story deps
- **US1 (P1)**: after US6 — MVP
- **US2 (P1)**: after US6; shares CLI group/render with US1
- **US3 (P1)**: after US6; may share render with US1
- **US4 (P1)**: after US6; shares CLI group with US1
- **US5 (P2)**: after CLI commands exist (US1–US4)

### Within Each User Story

- Tests MUST fail before implementation (TDD)
- Facade methods before CLI that call them
- `make test` coverage MUST NOT drop at story checkpoints

### Parallel Opportunities

- T002–T004 (skim) in parallel
- T020–T022 (US6 tests) in parallel before T030
- T040–T041 (US1 tests) in parallel
- T060–T062 (US2 tests) in parallel
- T080–T081 (US3 tests) in parallel
- T100–T101 (US4 tests) in parallel
- T120–T121 (US5 tests) in parallel
- After US6 + US1 group/render: US2 / US3 / US4 can proceed in parallel if staffed

---

## Parallel Example: User Story 6

```bash
# Launch US6 failing tests together (before implementation):
Task: "ListKanbanLinks httptest in internal/api/kanban_task_status_test.go"
Task: "Get/Create/Update/Delete happy paths in internal/api/kanban_task_status_test.go"
Task: "404 KindNotFound in internal/api/kanban_task_status_test.go"
Task: "MoveTaskToKanban 0/1/>1 in internal/api/kanban_task_status_test.go"
```

---

## Parallel Example: User Story 1

```bash
# Launch US1 failing tests together:
Task: "list happy path + empty [] + no GetTask in internal/cli/task_kanban_list_test.go"
Task: "get single object + 404 in internal/cli/task_kanban_get_test.go"
```

---

## Implementation Strategy

### MVP First (US6 + US1)

1. Phase 1 Setup + Phase 2 Foundation
2. Phase 3 US6 facade (+ Move)
3. Phase 4 US1 list/get (+ kanban group)
4. **STOP and VALIDATE**: list/get independently (SC-006 read)
5. Continue US2 → US3 → US4 → US5 → Polish

### Incremental Delivery

1. Setup + Foundation → ready
2. US6 → API DoD for KanbanTaskStatusController_* + Move
3. US1 → MVP CLI read
4. US2 → explicit write path
5. US3 → move UX
6. US4 → delete
7. US5 → help polish
8. Polish → quickstart + final `make test`

### Suggested MVP scope

**US6 + US1** (facade + list/get). Enough to prove F10 read path and KanbanTaskStatusController list/getById. Full acceptance needs US2–US4 (+ US5 for help).

---

## Notes

- [P] = different files, no incomplete deps
- Pre-check GetTask lives in **CLI** (create/move only), not kanban facade (research §2)
- Move orchestration lives in **facade** (`MoveTaskToKanban`) for TUI reuse (research §1)
- No `--limit`/`--offset`/`--removed`; no `--order` on move; no column CRUD / TUI
- Verify tests fail before implementing (TDD)
- Commit after each task or logical group (when asked)
- Avoid: skipping US6 before CLI, leaking task get / list body to stdout on move, hand-written HTTP, client-side create uniqueness
