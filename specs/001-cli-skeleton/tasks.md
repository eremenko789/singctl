# Tasks: CLI Skeleton

**Input**: Design documents from `/specs/001-cli-skeleton/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/, quickstart.md

**Tests**: Включены — constitution Quality Gates требуют unit-тесты парсеров CLI; plan/research фиксируют `internal/cli/root_test.go` и сценарии контракта.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- CLI entrypoint: `cmd/singctl/`
- Command tree: `internal/cli/`
- Build metadata: `internal/buildinfo/`
- Paths follow plan.md (no `src/` layout)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Инициализация Go-модуля и каркаса каталогов под F01

- [X] T001 Create directories `cmd/singctl/`, `internal/cli/`, `internal/buildinfo/` per plan.md
- [X] T002 Initialize Go module `github.com/eremenko789/singctl` in `go.mod` (toolchain ≥ 1.22)
- [X] T003 Add dependencies `github.com/spf13/cobra` and `github.com/spf13/viper` via `go get` and refresh `go.sum`
- [X] T004 Create entrypoint stub that calls `cli.Execute()` in `cmd/singctl/main.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Общая инфраструктура CLI, без которой нельзя начать user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Implement `VersionIdentity` placeholders (`Name=singctl`, `Version=dev`, `Commit=unknown`, `Date=unknown`) with ldflags-ready vars in `internal/buildinfo/buildinfo.go`
- [X] T006 Implement `Execute()` and minimal root `cobra.Command` (Use=`singctl`, SilenceErrors/SilenceUsage as needed) in `internal/cli/root.go`
- [X] T007 Wire root Russian `SetFlagErrorFunc` / error printing to stderr for unknown flags in `internal/cli/root.go`
- [X] T008 Define `GlobalOptions` struct (ConfigPath, Token, Output, NoColor, Debug) for session flags in `internal/cli/options.go`

**Checkpoint**: Foundation ready — `make build` can compile a stub binary; user stories may start

---

## Phase 3: User Story 1 — Discover the tool via help (Priority: P1) 🎯 MVP

**Goal**: Русскоязычная корневая справка, ошибка при вызове без аргументов, отсутствие entity/TUI-команд

**Independent Test**: `make build && ./bin/singctl --help` — RU help с глобальными флагами (когда появятся) / списком команд без `task|project|habit|tag|time|tui`; `./bin/singctl` → exit ≠ 0 и RU ошибка; сеть не нужна

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T009 [P] [US1] Add failing tests for `--help` (RU text, no entity/TUI commands) and bare invoke (nonzero exit, RU error on stderr) in `internal/cli/root_test.go`

### Implementation for User Story 1

- [X] T010 [US1] Set root `Short`/`Long` and help template texts in Russian in `internal/cli/root.go`
- [X] T011 [US1] Implement root `RunE` for bare invoke (RU: нет команды / TUI не реализован, exit ≠ 0) in `internal/cli/root.go`
- [X] T012 [US1] Ensure unknown subcommand path yields nonzero exit and RU/понятное сообщение with help hint in `internal/cli/root.go`
- [X] T013 [US1] Verify no registration of `task`, `project`, `habit`, `tag`, `time`, `tui` subcommands in `internal/cli/root.go`
- [X] T014 [US1] Make tests in `internal/cli/root_test.go` pass for help and bare/unknown-command scenarios

**Checkpoint**: User Story 1 independently testable via `make test` / `./bin/singctl --help` and bare invoke

---

## Phase 4: User Story 2 — Check installed version (Priority: P1)

**Goal**: `singctl version` и `singctl --version` печатают одинаковый stdout (имя, версия, commit/date)

**Independent Test**: `./bin/singctl version` и `./bin/singctl --version` — одинаковый непустой stdout с `singctl`, version, commit/date; exit 0; без сети

### Tests for User Story 2

- [X] T015 [P] [US2] Add failing tests for version parity (`version` vs `--version`) and required fields in `internal/cli/root_test.go`

### Implementation for User Story 2

- [X] T016 [P] [US2] Add `Format()` / display helper for VersionIdentity stdout payload in `internal/buildinfo/buildinfo.go`
- [X] T017 [US2] Implement `version` subcommand printing buildinfo to stdout in `internal/cli/version.go`
- [X] T018 [US2] Register `version` on root and configure Cobra `--version` with same template/payload as subcommand in `internal/cli/root.go`
- [X] T019 [US2] Make version parity tests pass in `internal/cli/root_test.go`

**Checkpoint**: User Stories 1 and 2 both work independently

---

## Phase 5: User Story 3 — Accept global flags for later use (Priority: P2)

**Goal**: Persistent global flags accepted; `--output`/`-o` validated as `table|json|yaml|csv` before help/version; Viper BindPFlag without ReadInConfig

**Independent Test**: Valid flags with `version`/`--help` — no unknown-flag error; invalid `--output` (e.g. `xml`) with `--help`/`version`/`--version` → exit ≠ 0, RU validation error, no help/version body

### Tests for User Story 3

- [X] T020 [P] [US3] Add failing tests for all valid global flags accepted and invalid `--output` blocking help/version in `internal/cli/root_test.go`

### Implementation for User Story 3

- [X] T021 [P] [US3] Implement `OutputFormat` as `pflag.Value` with enum `table|json|yaml|csv` and RU `Set` errors in `internal/cli/output.go`
- [X] T022 [US3] Register persistent flags `--config`, `--token`, `--output`/`-o`, `--no-color`, `--debug` on root and bind to `GlobalOptions` in `internal/cli/root.go`
- [X] T023 [US3] Bind persistent flags to Viper keys (`config`, `token`, `output`, `no-color`, `debug`) without `ReadInConfig` in `internal/cli/root.go`
- [X] T024 [US3] Ensure RU Usage strings for global flags appear in `--help` in `internal/cli/root.go`
- [X] T025 [US3] Make global-flag and invalid-output contract tests pass in `internal/cli/root_test.go`

**Checkpoint**: All three user stories independently functional; contracts/cli.md covered by tests

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Сборка, сверка с quickstart, мелкие UX/качество

- [X] T026 Confirm `make build` produces executable `bin/singctl` from `./cmd/singctl` via existing `Makefile`
- [X] T027 [P] Run and fix `make test` (`go test ./...`) until all CLI contract cases are green
- [X] T028 [P] Manually validate quickstart.md scenarios 1–6 against `./bin/singctl` (help, version parity, flags, invalid output, bare invoke, unknown command)
- [X] T029 Review stderr vs stdout separation (errors → stderr, version/help → stdout) in `internal/cli/root.go` and `internal/cli/version.go`
- [X] T030 Ensure `--token` is never written to files or debug logs in F01 code paths under `internal/cli/`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational — MVP
- **User Story 2 (Phase 4)**: Depends on Foundational; may follow or parallel US1 after T008 (shared `root.go` / `root_test.go` — coordinate edits)
- **User Story 3 (Phase 5)**: Depends on Foundational; ideally after US1 so help already shows RU texts for flags
- **Polish (Phase 6)**: Depends on US1–US3 completion

### User Story Dependencies

- **User Story 1 (P1)**: After Phase 2 — no dependency on US2/US3 for core help/bare invoke
- **User Story 2 (P1)**: After Phase 2 — independent of US3; uses `buildinfo` from T005/T016
- **User Story 3 (P2)**: After Phase 2 — extends root flags; invalid `--output` vs help integrates with US1 behavior

### Within Each User Story

- Tests FIRST and FAIL before implementation
- Story implementation before marking tests green
- Story complete before moving to next priority when editing the same files sequentially

### Parallel Opportunities

- T009 / T015 / T020 can be authored in parallel only if split carefully (prefer sequential appends to `root_test.go` when one developer)
- T016 (buildinfo Format) can run in parallel with US1 implementation if US2 not sharing `root.go` yet
- T021 (`output.go`) can start in parallel with US2 once Phase 2 is done
- T027 and T028 are parallelizable after implementation

---

## Parallel Example: User Story 1

```bash
# After Phase 2, single developer sequential on shared files is safer:
Task: "T009 Add failing help/bare-invoke tests in internal/cli/root_test.go"
Task: "T010–T013 Implement RU help, RunE, unknown command, no entity cmds in internal/cli/root.go"
Task: "T014 Make US1 tests pass in internal/cli/root_test.go"
```

## Parallel Example: After Foundation (two developers)

```bash
# Developer A — US1 on internal/cli/root.go + root_test.go
# Developer B — T016 Format() in internal/buildinfo/buildinfo.go (then wait to merge version cmd)
Task: "T016 Add Format() in internal/buildinfo/buildinfo.go"
Task: "T021 Implement OutputFormat in internal/cli/output.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: `./bin/singctl --help`, bare `./bin/singctl`, `make test` for US1 cases
5. Demo/onboard on help-only skeleton

### Incremental Delivery

1. Setup + Foundational → compilable stub
2. US1 → usable RU help / bare-error MVP
3. US2 → version / `--version` for CI diagnostics
4. US3 → global flag contract for F02+
5. Polish → quickstart + Makefile verification

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. After Foundational:
   - Prefer one owner for `internal/cli/root.go` / `root_test.go` sequencing US1 → US2 → US3
   - Parallelize only distinct files (`buildinfo.go`, `output.go`, `version.go`) with merge discipline

---

## Notes

- [P] = different files, no incomplete-task dependency
- Do not create `internal/api`, `internal/apiclient`, `internal/config`, `internal/tui`, or entity cmds in F01
- Viper: BindPFlag only — no `ReadInConfig` (F02)
- Exact numeric exit codes deferred to F07; F01 needs nonzero on errors
- Commit after each task or logical group when implementing
- Stop at checkpoints to validate each story independently
`)
