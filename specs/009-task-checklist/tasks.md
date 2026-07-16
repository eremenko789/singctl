# Tasks: Task Checklist (F09)

**Input**: Design documents from `/specs/009-task-checklist/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Токены только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Labels: `[US5]` API checklist facade (все ChecklistItemController_*), `[US1]` list/get CLI (+ pre-check), `[US2]` add/update CLI, `[US3]` delete CLI, `[US4]` help discoverability. Фасад (US5) — до CLI-историй. Зависит от F08 (`GetTask`, output SingleObject, `task` group).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]`…`[US5]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: baseline и контракты F09 перед TDD.

- [X] T001 Confirm F08/module builds green before F09 edits: `go test ./internal/api/ ./internal/output/ ./internal/cli/ ./cmd/singctl/` (or `make test`)
- [X] T002 [P] Skim `specs/009-task-checklist/contracts/api-checklist-facade.md` and `specs/009-task-checklist/data-model.md`
- [X] T003 [P] Skim `specs/009-task-checklist/contracts/cli-checklist.md` and `specs/009-task-checklist/contracts/checklist-output.md`
- [X] T004 [P] Skim `specs/009-task-checklist/research.md` (CLI pre-check via GetTask; list parent-only; no `--order`; done/undone)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: типы ChecklistItem/query/write и подтверждение F08-зависимостей — блокируют фасад и CLI. SingleObject уже есть из F08 (не дублировать).

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F09 changes
- [X] T011 Confirm `Session.GetTask` exists in `internal/api/task.go` and `RenderOptions.SingleObject` works in `internal/output/` (smoke via existing F08 tests; no new output APIs)
- [X] T012 Define `ChecklistItem`, `ChecklistListQuery`, `ChecklistWriteInput` types (per `specs/009-task-checklist/data-model.md`) in `internal/api/checklist.go` (types only / stubs OK); update package doc in `internal/api/doc.go`
- [X] T013 Run `make test` after foundation (Makefile)

**Checkpoint**: Foundation ready — US5 (facade) can start.

---

## Phase 3: User Story 5 — Adapter coverage with unit tests (Priority: P1)

**Goal**: Фасад `ListChecklistItems` / `GetChecklistItem` / `CreateChecklistItem` / `UpdateChecklistItem` / `DeleteChecklistItem` поверх codegen; unit-тесты httptest для всех ChecklistItemController_*; 404 → KindNotFound. Фасад **не** вызывает GetTask.

**Independent Test**: `go test ./internal/api/ -count=1` — happy path на list/create/get/update/delete; get 404 → `KindNotFound`; list query содержит только `parent`; без cobra.

### Tests for User Story 5 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before facade implementation.

- [X] T020 [P] [US5] Add failing httptest test: `ListChecklistItems` maps `Parent` only → query + returns items from `checklistItems` in `internal/api/checklist_test.go`
- [X] T021 [P] [US5] Add failing httptest tests: `GetChecklistItem` / `CreateChecklistItem` / `UpdateChecklistItem` / `DeleteChecklistItem` happy paths in `internal/api/checklist_test.go`
- [X] T022 [P] [US5] Add failing test: get (or update/delete) HTTP 404 → `Classify` KindNotFound with `WithEntityID` in `internal/api/checklist_test.go`
- [X] T023 [US5] Add failing test: create body has parent+title (optional done); update sends only set fields (title and/or done); no parentOrder/crypted in `internal/api/checklist_test.go`

### Implementation for User Story 5

- [X] T030 [US5] Implement `ListChecklistItems` / `GetChecklistItem` in `internal/api/checklist.go` via `ChecklistItemController*WithResponse` + `EnsureSuccess` + `Classify`
- [X] T031 [US5] Implement `CreateChecklistItem` and `UpdateChecklistItem` (partial DTO; no parentOrder/parent change) in `internal/api/checklist.go`
- [X] T032 [US5] Implement `DeleteChecklistItem` in `internal/api/checklist.go` (204 success)
- [X] T033 [US5] Make US5 tests green in `internal/api/checklist_test.go` (Authorization `Bearer test-token-…`)

### Coverage gate for US5

- [X] T034 Run `make test` after US5 (Makefile)

**Checkpoint**: All ChecklistItemController_* reachable via facade + mocked unit tests (SC-002/003).

---

## Phase 4: User Story 1 — List and inspect checklist items (Priority: P1) 🎯 MVP

**Goal**: `singctl task checklist list <TASK_ID>` (pre-check GetTask → list parent-only) и `task checklist get <ID>`; рендер F06; json list=массив, get=объект; exit/streams F07; ответ pre-check не в stdout.

**Independent Test**: httptest + `executeForTest`: list after task 200 → checklist list; list json array; get json object; task 404 → ExitCode 3, no checklist HTTP; item 404 → ExitCode 3; empty checklist → `[]`.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US1] Add failing tests: `task checklist list` happy path + empty list json `[]` + assert GetTask then checklist list with parent in `internal/cli/task_checklist_list_test.go` (httptest + temp config like task tests)
- [X] T041 [P] [US1] Add failing tests: unknown TASK_ID (task 404) → ExitCode 3, empty stdout, checklist endpoint not hit in `internal/cli/task_checklist_list_test.go`
- [X] T042 [P] [US1] Add failing tests: `task checklist get` json single object; item 404 → ExitCode 3, empty stdout in `internal/cli/task_checklist_get_test.go`
- [X] T043 [US1] Add failing tests: no token → ExitCode 2; empty/whitespace TASK_ID/item id → ExitCode 1 in list/get tests or `internal/cli/task_checklist_auth_test.go`

### Implementation for User Story 1

- [X] T050 [US1] Add `newTaskChecklistCmd()` group and register under `newTaskCmd()` in `internal/cli/task_checklist_cmd.go` + update `internal/cli/task_cmd.go` (remove «checklist недоступны» from Long if still present)
- [X] T051 [US1] Implement ChecklistItem→`output.RecordSet` mapper (columns per `contracts/checklist-output.md`) in `internal/cli/task_checklist_render.go`
- [X] T052 [US1] Implement `task checklist list` in `internal/cli/task_checklist_list.go` (validate TASK_ID → GetTask → `ListChecklistItems` → `Render` SingleObject=false; discard task body)
- [X] T053 [US1] Implement `task checklist get` in `internal/cli/task_checklist_get.go` (`Render` SingleObject=true; no task pre-check)
- [X] T054 [US1] Make US1 tests green; assert success stderr empty / error stdout empty where applicable

### Coverage gate for US1

- [X] T055 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP CLI — list/get usable for pipe scripts (SC-006).

---

## Phase 5: User Story 2 — Add and update checklist items (Priority: P1)

**Goal**: `task checklist add` / `update` с `--title` / `--done` / `--undone`; pre-check на add; empty title → exit 1; update без флагов → exit 1; stdout = полный пункт.

**Independent Test**: add `--title` → json object after GetTask; add without title / whitespace title → ExitCode 1 no network; task 404 on add → ExitCode 3 no create; update `--done`; update no flags → ExitCode 1; done+undone → ExitCode 1.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US2] Add failing tests: `task checklist add --title` happy path + optional `--done`; assert GetTask then POST create in `internal/cli/task_checklist_add_test.go`
- [X] T061 [P] [US2] Add failing tests: add without `--title`; whitespace `--title`; task 404 → ExitCode 3 no create in `internal/cli/task_checklist_add_test.go`
- [X] T062 [P] [US2] Add failing tests: `task checklist update` with `--title` / `--done` / `--undone`; no write flags → ExitCode 1; done+undone → ExitCode 1; empty `--title` → ExitCode 1; item 404 → ExitCode 3 in `internal/cli/task_checklist_update_test.go`

### Implementation for User Story 2

- [X] T070 [US2] Implement `task checklist add` in `internal/cli/task_checklist_add.go` (title validation before GetTask; GetTask; CreateChecklistItem; SingleObject stdout)
- [X] T071 [US2] Implement `task checklist update` in `internal/cli/task_checklist_update.go` (≥1 write flag; mutually exclusive done/undone; partial write; no parent/order flags)
- [X] T072 [US2] Make US2 tests green; streams/exit per F07

### Coverage gate for US2

- [X] T073 Run `make test` after US2 (Makefile)

**Checkpoint**: add/update closed.

---

## Phase 6: User Story 3 — Delete checklist items (Priority: P1)

**Goal**: `task checklist delete <ID>` → пустой stdout; 404 → exit 3; empty id → exit 1; без confirm.

**Independent Test**: delete 204 + empty stdout; 404 → ExitCode 3; whitespace id → ExitCode 1 no network.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US3] Add failing tests: `task checklist delete` success empty stdout; 404 → ExitCode 3 in `internal/cli/task_checklist_delete_test.go`
- [X] T081 [US3] Add failing tests: empty/whitespace id → ExitCode 1 no HTTP in `internal/cli/task_checklist_delete_test.go`

### Implementation for User Story 3

- [X] T090 [US3] Implement `task checklist delete` in `internal/cli/task_checklist_delete.go` (no render on success)
- [X] T091 [US3] Make US3 tests green

### Coverage gate for US3

- [X] T092 Run `make test` after US3 (Makefile)

**Checkpoint**: delete closed (FR-006).

---

## Phase 7: User Story 4 — Discoverable CLI help (Priority: P2)

**Goal**: `task checklist --help` и help пяти подкоманд документируют scope F09; `task --help` показывает `checklist`; нет TUI/kanban/`move`/`--order`/pagination promises.

**Independent Test**: `executeForTest` / `--help` substring tests; unknown checklist subcommand → ExitCode 1.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T100 [P] [US4] Add failing test: `task --help` lists `checklist`; `task checklist --help` lists `list|get|add|update|delete` in `internal/cli/task_checklist_help_test.go`
- [X] T101 [P] [US4] Add failing tests: each checklist subcommand `--help` mentions key args/flags; MUST NOT claim `--order`, pagination, TUI, kanban/move in `internal/cli/task_checklist_help_test.go`
- [X] T102 [US4] Add failing test: unknown `task checklist` subcommand → ExitCode 1, empty stdout in `internal/cli/task_checklist_help_test.go`

### Implementation for User Story 4

- [X] T110 [US4] Flesh out Short/Long/Example and flag usage across `internal/cli/task_checklist_cmd.go`, `task_checklist_list.go`, `task_checklist_get.go`, `task_checklist_add.go`, `task_checklist_update.go`, `task_checklist_delete.go`; sync `task_cmd.go` help text
- [X] T111 [US4] Make US4 help tests green

### Coverage gate for US4

- [X] T112 Run `make test` after US4 (Makefile)

**Checkpoint**: SC-004 help discoverability.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: сквозные проверки DoD, docs pointers, coverage.

- [X] T120 [P] Add/adjust cross-cutting exit+stream matrix sample for list pre-check failure + one mutate path if gaps remain in `internal/cli/task_checklist_*_test.go`
- [X] T121 [P] Optional: one-line pointer in `docs/scriptability.md` that pipe examples apply to `task checklist list`/`add` (F09) — keep honest if only mock-proven
- [X] T122 Confirm `internal/api/doc.go` mentions checklist facade
- [X] T123 Run quickstart checks from `specs/009-task-checklist/quickstart.md` (`make test` + optional `go build` help smoke)
- [X] T124 Coverage gate: final `make test` with no regression vs Phase 2 baseline (Makefile)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: depends on Setup — BLOCKS all user stories
- **US5 (Phase 3)**: depends on Foundation — BLOCKS CLI stories that call facade
- **US1 (Phase 4)**: depends on US5 — MVP CLI; registers `task checklist` group
- **US2 (Phase 5)**: depends on US5 + preferably US1 (shared render/group)
- **US3 (Phase 6)**: depends on US5 + preferably US1 (shared group)
- **US4 (Phase 7)**: depends on commands existing (after US1–US3 ideally)
- **Polish (Phase 8)**: after desired stories complete

### User Story Dependencies

- **US5 (P1)**: after Foundation — no other story deps
- **US1 (P1)**: after US5 — MVP
- **US2 (P1)**: after US5; shares CLI group/render with US1
- **US3 (P1)**: after US5; shares CLI group with US1
- **US4 (P2)**: after CLI commands exist (US1–US3)

### Within Each User Story

- Tests MUST fail before implementation (TDD)
- Facade methods before CLI that call them
- `make test` coverage MUST NOT drop at story checkpoints

### Parallel Opportunities

- T002–T004 (skim) in parallel
- T020–T022 (US5 tests) in parallel before T030
- T040–T042 (US1 tests) in parallel
- T060–T062 (US2 tests) in parallel
- T080–T081 (US3 tests) in parallel
- T100–T101 (US4 tests) in parallel
- After US5: US2/US3 can proceed in parallel if US1 group/render stubs exist

---

## Parallel Example: User Story 5

```bash
# Launch US5 failing tests together (before implementation):
Task: "ListChecklistItems httptest in internal/api/checklist_test.go"
Task: "Get/Create/Update/Delete happy paths in internal/api/checklist_test.go"
Task: "404 KindNotFound in internal/api/checklist_test.go"
```

---

## Parallel Example: User Story 1

```bash
# Launch US1 failing tests together:
Task: "list happy path + empty [] in internal/cli/task_checklist_list_test.go"
Task: "task 404 pre-check in internal/cli/task_checklist_list_test.go"
Task: "get single object + 404 in internal/cli/task_checklist_get_test.go"
```

---

## Implementation Strategy

### MVP First (US5 + US1)

1. Phase 1 Setup + Phase 2 Foundation
2. Phase 3 US5 facade
3. Phase 4 US1 list/get (+ checklist group)
4. **STOP and VALIDATE**: list/get independently (SC-006)
5. Continue US2 → US3 → US4 → Polish

### Incremental Delivery

1. Setup + Foundation → ready
2. US5 → API DoD for ChecklistItemController_*
3. US1 → MVP CLI
4. US2 → write path
5. US3 → delete
6. US4 → help polish
7. Polish → quickstart + final `make test`

### Suggested MVP scope

**US5 + US1** (facade + list/get with pre-check). Enough to prove F09 read path and ChecklistItemController list/getById.

---

## Notes

- [P] = different files, no incomplete deps
- Pre-check GetTask lives in **CLI**, not checklist facade (research §2)
- No `--limit`/`--offset`/`--removed`/`--order`; no TUI
- Verify tests fail before implementing (TDD)
- Commit after each task or logical group (when asked)
- Avoid: skipping US5 before CLI, leaking task get body to stdout, hand-written HTTP
