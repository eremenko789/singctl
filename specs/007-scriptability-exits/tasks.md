# Tasks: Scriptability & Exit Codes (F07)

**Input**: Design documents from `/specs/007-scriptability-exits/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом или проверяемым артефактом DoD: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Токены только `test-token-…` / `fake-…` (constitution VII).

**Organization**: Labels: `[US1]` documented exit codes + help + misuse→1, `[US2]` stdout/stderr streams, `[US3]` pipe §10 contract, `[US4]` non-interactive stdin. DoD без новых пользовательских команд — `config show` / `config validate` / misuse / docs+`--help` / reuse F06.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]` / `[US2]` / `[US3]` / `[US4]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: подтвердить baseline и контракты F07 перед TDD.

- [X] T001 Confirm module builds / existing CLI tests green: `go test ./internal/cli/ ./internal/output/ ./cmd/singctl/` (or `make test`) before F07 edits
- [X] T002 [P] Skim `specs/007-scriptability-exits/contracts/exit-codes-public.md` and `specs/005-error-retry/contracts/cli-exit-codes.md` for ExitCode SoT + misuse→1
- [X] T003 [P] Skim `specs/007-scriptability-exits/contracts/stream-separation.md` and `specs/007-scriptability-exits/contracts/pipe-scenarios.md`
- [X] T004 [P] Skim `specs/007-scriptability-exits/research.md` and `specs/007-scriptability-exits/data-model.md` (docs path, help blurb, stream rules)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: coverage baseline; убедиться что `cli.ExitCode` + `main` wiring на месте (F05) — F07 их не заменяет.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F07 changes
- [X] T011 Confirm `cli.ExitCode` exists and `cmd/singctl/main.go` uses `os.Exit(cli.ExitCode(err))` (read-only check; no competing helper)
- [X] T012 [P] Confirm `executeForTest` in `internal/cli/root.go` captures stdout/stderr separately (harness for US1–US4)

**Checkpoint**: Foundation ready — US1 can start.

---

## Phase 3: User Story 1 — Stable documented exit codes (Priority: P1) 🎯 MVP

**Goal**: Публичная таблица `0/1/2/3` в `docs/scriptability.md` + краткое упоминание в корневом `--help`; misuse CLI → exit `1`; матрица `config validate` согласована с таблицей.

**Independent Test**: `--help` содержит коды 0–3; `docs/scriptability.md` содержит полную таблицу; `--unknown-flag` / invalid `--output` → `ExitCode==1`, stdout пуст; validate 0/1/2/3 как в F05.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before docs/help implementation where content is missing.

- [X] T020 [P] [US1] Add failing test: `singctl --help` stdout mentions exit codes `0`, `1`, `2`, `3` in `internal/cli/root_test.go`
- [X] T021 [P] [US1] Add failing test: `docs/scriptability.md` exists and contains table meanings for codes 0–3 (read via relative path from module root) in `internal/cli/scriptability_docs_test.go` (new) or `docs/scriptability_test.go` if preferred under `docs/` package — prefer `internal/cli/scriptability_docs_test.go` reading `../../docs/scriptability.md`
- [X] T022 [P] [US1] Add failing test: unknown flag → `ExitCode(err)==1`, stdout empty after trim, stderr non-empty in `internal/cli/root_test.go`
- [X] T023 [P] [US1] Add failing test: invalid `--output` value → `ExitCode(err)==1`, stdout empty in `internal/cli/root_test.go`
- [X] T024 [US1] Assert (or extend) `config validate` exit matrix still maps success→0, no-token→2, mock 404→3, mock 401→1 with `ExitCode` in `internal/cli/config_validate_test.go` (strengthen if gaps)

### Implementation for User Story 1

- [X] T030 [US1] Create `docs/scriptability.md` with full exit table (TZ §10 / `contracts/exit-codes-public.md`), brief stdout/stderr rules, pointer to pipe scenarios
- [X] T031 [P] [US1] Add link row to `docs/scriptability.md` in `docs/README.md` index table
- [X] T032 [US1] Extend root command `Long` (or `Example`) with brief exit-codes blurb + docs pointer in `internal/cli/root.go`
- [X] T033 [US1] Make US1 tests green (`internal/cli/root_test.go`, `internal/cli/scriptability_docs_test.go`, validate tests); fix only if misuse somehow maps to 2/3

### Coverage gate for US1

- [X] T034 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP — documented exits + help + misuse→1 proven.

---

## Phase 4: User Story 2 — Clean stdout vs stderr for pipes (Priority: P1)

**Goal**: На DoD-поверхности успех → data в stdout, stderr пуст; ошибка → сообщение в stderr, stdout пуст.

**Independent Test**: `config show` OK: stderr empty, stdout has data; show/validate/misuse errors: stdout empty, stderr has message.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US2] Add failing assertions: successful `config show` → `strings.TrimSpace(stderr)==""` and non-empty stdout in `internal/cli/config_show_test.go`
- [X] T041 [P] [US2] Add failing assertions: error paths of `config show` → `strings.TrimSpace(stdout)==""` and non-empty stderr in `internal/cli/config_show_test.go`
- [X] T042 [P] [US2] Add failing assertions: successful `config validate` → stderr empty in `internal/cli/config_validate_test.go`
- [X] T043 [US2] Add failing assertions: all error paths in `config_validate_test.go` (no token, 401, 404, 429, 5xx) → stdout empty after trim (not only `_` ignore)
- [X] T044 [US2] Confirm misuse tests from US1 also assert empty stdout (shared with stream contract) in `internal/cli/root_test.go`

### Implementation for User Story 2

- [X] T050 [US2] Fix any command paths that write errors or success diagnostics to the wrong stream in `internal/cli/config_show.go`, `internal/cli/config_validate.go`, and/or `internal/cli/root.go` (`Execute` / `executeForTest`) until US2 tests pass
- [X] T051 [US2] Make US2 tests green in `internal/cli/config_show_test.go`, `internal/cli/config_validate_test.go`, `internal/cli/root_test.go`

### Coverage gate for US2

- [X] T052 Run `make test` after US2 (Makefile)

**Checkpoint**: Stream separation proven on DoD commands.

---

## Phase 5: User Story 3 — Pipe-ready contract for TZ §10 (Priority: P1)

**Goal**: Четыре pipe-сценария ТЗ §10 зафиксированы в user docs + specs contract; проверяемые сейчас свойства закрыты F06 fixture + F07 streams/exits (entity E2E — F08+).

**Independent Test**: `docs/scriptability.md` (or linked section) mentions all four scenario IDs/statuses; `go test ./internal/output/` still green (JSON array / CSV / no ANSI); `contracts/pipe-scenarios.md` complete (already from plan — keep in sync with docs).

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US3] Add failing test: `docs/scriptability.md` references all four pipe scenario ids (`json-redirect`, `list-jq-xargs`, `csv-awk`, `xargs-create`) or equivalent TZ examples in `internal/cli/scriptability_docs_test.go`
- [X] T061 [P] [US3] Add failing test (or document-assert): `specs/007-scriptability-exits/contracts/pipe-scenarios.md` lists Status for each of four scenarios in `internal/cli/scriptability_docs_test.go` (read contract file) OR skip file read if prefer docs-only — prefer assert both docs + contract exist and mention four ids
- [X] T062 [US3] Run / assert existing F06 pipe-safe tests still pass: `go test ./internal/output/ -count=1` (no ANSI / json-csv) — add thin wrapper test in `internal/cli/scriptability_pipe_reuse_test.go` only if needed to bind SC-005 into F07 package; otherwise rely on `make test` including `./internal/output`

### Implementation for User Story 3

- [X] T070 [US3] Extend `docs/scriptability.md` with pipe §10 matrix (properties + verifiable_now vs F08+) aligned with `specs/007-scriptability-exits/contracts/pipe-scenarios.md`
- [X] T071 [US3] Sync wording in `specs/007-scriptability-exits/contracts/pipe-scenarios.md` if docs introduce clearer labels (keep SC-004 coverage)
- [X] T072 [US3] Make US3 doc/contract tests green in `internal/cli/scriptability_docs_test.go`

### Coverage gate for US3

- [X] T073 Run `make test` after US3 (Makefile)

**Checkpoint**: Pipe scenarios contract-level DoD closed.

---

## Phase 6: User Story 4 — Predictable behavior when stdin is a pipe (Priority: P2)

**Goal**: Piped/closed stdin не блокирует существующие команды; ошибки без interactive prompt.

**Independent Test**: `config show` or `version` with `SetIn` finite/closed reader completes; error path still stderr + non-zero ExitCode without hang.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US4] Add failing/timeout-guarded test: `executeForTest` (or variant with `cmd.SetIn`) with closed/empty stdin + `version` or `config show` completes successfully without blocking in `internal/cli/root_test.go` or `internal/cli/noninteractive_test.go`
- [X] T081 [US4] Add test: error path with piped stdin (e.g. unknown command) → non-nil err, ExitCode≠0, stderr non-empty, no hang in `internal/cli/noninteractive_test.go`

### Implementation for User Story 4

- [X] T090 [US4] Extend test harness if needed (`executeForTestWithIn` in `internal/cli/root.go`) to inject stdin for US4 tests
- [X] T091 [US4] Fix any accidental stdin reads/prompts on DoD paths if tests hang (unlikely — document in comment if none found)
- [X] T092 [US4] Make US4 tests green in `internal/cli/noninteractive_test.go` / `internal/cli/root_test.go`

### Coverage gate for US4

- [X] T093 Run `make test` after US4 (Makefile)

**Checkpoint**: Non-interactive stdin behavior proven.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: godoc, quickstart validation, final coverage.

- [X] T100 [P] Add/refresh godoc on `ExitCode` in `internal/cli/exit.go` with cross-link to `docs/scriptability.md` / F07 contract
- [X] T101 [P] Ensure `specs/007-scriptability-exits/quickstart.md` DoD checklist items are achievable; tweak docs if paths drifted
- [X] T102 Walk `specs/007-scriptability-exits/quickstart.md` manually or via notes: `--help`, misuse exit, `make test`
- [X] T103 [P] Final coverage gate: `make test` (Makefile) — no regression vs Phase 2 baseline
- [X] T104 Confirm no new user-facing CLI commands were added (diff `internal/cli/*_cmd*.go` / root AddCommand)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: нет зависимостей
- **Foundational (Phase 2)**: после Setup — BLOCKS all stories
- **US1 (Phase 3)**: после Foundation — MVP
- **US2 (Phase 4)**: после Foundation; ideally after US1 misuse tests exist (shares empty-stdout asserts) — can start after T022/T023 written
- **US3 (Phase 5)**: после US1 docs file exists (T030) — extends same `docs/scriptability.md`
- **US4 (Phase 6)**: после Foundation; independent of US3
- **Polish (Phase 7)**: после желаемых story checkpoints

### User Story Dependencies

- **US1 (P1)**: нет зависимостей от других stories
- **US2 (P1)**: логически усиливает те же CLI пути; независимо тестируемо
- **US3 (P1)**: зависит от появления `docs/scriptability.md` (US1)
- **US4 (P2)**: независимо после Foundation

### Within Each User Story

- Tests MUST fail before implementation (TDD)
- `make test` coverage MUST NOT drop at story checkpoint

### Parallel Opportunities

- T002–T004 skim in parallel
- T020–T023 US1 tests in parallel
- T040–T043 US2 tests in parallel (different assertion sites)
- T060–T061 US3 tests in parallel
- T080 || early US4 after Foundation
- T100–T101 polish in parallel

---

## Parallel Example: User Story 1

```bash
# Tests first (parallel):
Task: "T020 --help mentions exit codes in internal/cli/root_test.go"
Task: "T021 docs/scriptability.md presence test in internal/cli/scriptability_docs_test.go"
Task: "T022 unknown flag ExitCode 1 + empty stdout in internal/cli/root_test.go"
Task: "T023 invalid --output ExitCode 1 in internal/cli/root_test.go"

# Then implementation:
Task: "T030 Create docs/scriptability.md"
Task: "T031 Link in docs/README.md"
Task: "T032 Root Long exit blurb in internal/cli/root.go"
```

---

## Parallel Example: User Story 2

```bash
Task: "T040 config show success stderr empty in config_show_test.go"
Task: "T041 config show error stdout empty in config_show_test.go"
Task: "T042 config validate success stderr empty in config_validate_test.go"
# Then T043 error stdout empty on all validate failures
# Then T050 fixes if any
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1–2 Setup + Foundation
2. Phase 3 US1 (docs + help + misuse→1)
3. **STOP**: validate `--help`, docs table, misuse exit
4. Continue US2–US4 for full DoD

### Incremental Delivery

1. US1 → documented scriptability contract visible
2. US2 → pipe-safe streams on existing commands
3. US3 → TZ §10 matrix in docs/contracts
4. US4 → non-interactive stdin
5. Polish → `make test` + quickstart

### Parallel Team Strategy

1. Shared: Phase 1–2
2. Dev A: US1 → US3 (docs continuum)
3. Dev B: US2 streams (after or with US1 misuse tests)
4. Dev C: US4 noninteractive

---

## Notes

- [P] = different files, no incomplete-task dependency
- Не менять числовую семантику `ExitCode` (FR-010); не добавлять entity-команды
- Reuse F06 `internal/output` for ANSI/json/csv — не дублировать рендерер
- Commit after each task or logical group
- Skip sample template tasks — this file is the authoritative list
