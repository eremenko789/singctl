# Tasks: Output Rendering (F06)

**Input**: Design documents from `/specs/006-output-rendering/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story с production-кодом: сначала падающие тесты, затем минимальная реализация. Coverage MUST NOT регрессировать (`make test`). Fixture без секретов (constitution VII).

**Organization**: Labels: `[US1]` cross-format render, `[US2]` color / no-ANSI, `[US3]` ResolveFormat, `[US4]` date_format / null dates. DoD = unit/harness only — **без** новой CLI-команды и без миграции `config show`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete work)
- **[Story]**: `[US1]` / `[US2]` / `[US3]` / `[US4]` for story phases; setup/foundation/polish — без story label
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: подтвердить baseline и контракты F06 перед TDD.

- [X] T001 Confirm module builds: `go test ./...` (or at least existing packages) before creating `internal/output`
- [X] T002 [P] Skim `specs/006-output-rendering/contracts/output-render.md` for ResolveFormat / ColorEnabled / Render invariants to mirror in tests
- [X] T003 [P] Skim `specs/006-output-rendering/data-model.md` and `specs/006-output-rendering/research.md` (RecordSet, Format, date/color decisions)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: пакет-скелет + типы модели; зелёный coverage baseline до фичи.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

- [X] T010 Run coverage-gated baseline via `make test` (Makefile) and record green baseline before F06 changes
- [X] T011 Create package overview in `internal/output/doc.go`
- [X] T012 [P] Add `Format` constants (`table|json|yaml|csv`) in `internal/output/format.go` (ResolveFormat later in US3)
- [X] T013 Add `Column`, `RecordSet`, `RenderOptions` types in `internal/output/model.go` per `data-model.md`

**Checkpoint**: Foundation ready — US1 can start.

---

## Phase 3: User Story 1 — Same data in all output formats (Priority: P1) 🎯 MVP

**Goal**: `Render` пишет один и тот же Logical Record Set в `table|json|yaml|csv`; json/yaml — корневой массив; даты — строки по default layout; empty set валиден.

**Independent Test**: Fixture ≥3 rows с ≥1 date field → все 4 формата согласованы по числу записей и полям; empty → `[]` / header-only csv; csv escaping.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

> Write FIRST; ensure FAIL before Render / FormatDate implementation.

- [X] T020 [P] [US1] Add failing harness: fixture ≥3 rows / ≥1 date → all four formats agree on field keys and formatted values in `internal/output/render_test.go`
- [X] T021 [P] [US1] Add failing tests: empty RecordSet → json `[]`, yaml empty sequence, csv headers only, table without data rows in `internal/output/render_test.go`
- [X] T022 [P] [US1] Add failing tests: json/yaml root is array of objects (not wrapped); N elements == len(Rows) in `internal/output/render_test.go`
- [X] T023 [US1] Add failing tests: CSV header uses Column.Key; fields with commas/quotes/newlines escape correctly in `internal/output/render_test.go`

### Implementation for User Story 1

- [X] T030 [US1] Implement `NormalizeLayout` / `FormatDate` (default `2006-01-02`) in `internal/output/date.go` for use by renderers
- [X] T031 [US1] Implement shared cell normalization helpers (date → string, stringify) in `internal/output/model.go` (or `internal/output/cells.go` if split)
- [X] T032 [US1] Implement `Render` dispatcher in `internal/output/render.go`
- [X] T033 [P] [US1] Implement JSON renderer (root array; dates as formatted strings) in `internal/output/render_json.go`
- [X] T034 [P] [US1] Implement YAML renderer (root sequence; same logical values) in `internal/output/render_yaml.go`
- [X] T035 [P] [US1] Implement CSV renderer (`encoding/csv`) in `internal/output/render_csv.go`
- [X] T036 [US1] Implement plain table renderer (`tablewriter`, no color yet) in `internal/output/render_table.go`
- [X] T037 [US1] Make US1 tests green in `internal/output/render_test.go` (and date helpers as needed)

### Coverage gate for US1

- [X] T038 Run `make test` after US1 (Makefile)

**Checkpoint**: MVP — cross-format RecordSet render proven offline.

---

## Phase 4: User Story 2 — Pipe-safe output without ANSI (Priority: P1)

**Goal**: `ColorEnabled` + политика ANSI: machine formats never colored; table ANSI only when color on; non-TTY / `--no-color` / `NO_COLOR` → off.

**Independent Test**: ColorEnabled matrix; render json/yaml/csv never contain `\x1b[`; table with Color=false has no ANSI; Color=true MAY have ANSI only in table.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T040 [P] [US2] Add failing matrix tests for `ColorEnabled(isTTY, noColorFlag, noColorEnv, configColor)` in `internal/output/color_test.go`
- [X] T041 [P] [US2] Add failing tests: json/yaml/csv output never matches ANSI CSI (`\x1b[`) even if Color=true in `internal/output/render_test.go`
- [X] T042 [US2] Add failing tests: table with `opts.Color=false` has no ANSI; with `opts.Color=true` MAY contain ANSI only in table bytes in `internal/output/render_test.go`

### Implementation for User Story 2

- [X] T050 [US2] Implement `ColorEnabled` in `internal/output/color.go` per `contracts/output-render.md`
- [X] T051 [US2] Wire `opts.Color` into table renderer only in `internal/output/render_table.go` (machine formats ignore Color)
- [X] T052 [US2] Make US2 tests green in `internal/output/color_test.go` and `internal/output/render_test.go`

### Coverage gate for US2

- [X] T053 Run `make test` after US2 (Makefile)

**Checkpoint**: Pipe-safe / no-ANSI policy proven.

---

## Phase 5: User Story 3 — Choose format via flag or config (Priority: P2)

**Goal**: `ResolveFormat(flagSet, flagValue, configFormat)` — флаг > конфиг > `table`; invalid/empty config → `table`.

**Independent Test**: Три unit-сценария SC-005 без CLI.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T060 [P] [US3] Add failing tests: flag set wins over config in `internal/output/format_test.go`
- [X] T061 [P] [US3] Add failing tests: flag unset + valid config → config format in `internal/output/format_test.go`
- [X] T062 [US3] Add failing tests: flag unset + empty/invalid config → `table` in `internal/output/format_test.go`

### Implementation for User Story 3

- [X] T070 [US3] Implement `ResolveFormat` in `internal/output/format.go`
- [X] T071 [US3] Make US3 tests green in `internal/output/format_test.go`

### Coverage gate for US3

- [X] T072 Run `make test` after US3 (Makefile)

**Checkpoint**: Format precedence closed (library-only).

---

## Phase 6: User Story 4 — Consistent date display via date_format (Priority: P2)

**Goal**: `date_format` применяется во всех форматах; invalid layout → default; nil date → json/yaml `null`, table/csv empty cell.

**Independent Test**: Два layout → два литерала во всех форматах; nil date serialization; invalid layout fallback.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T080 [P] [US4] Add failing tests: empty/invalid layout → `NormalizeLayout` returns `2006-01-02` in `internal/output/date_test.go`
- [X] T081 [P] [US4] Add failing tests: two valid layouts produce two different literals via `FormatDate` in `internal/output/date_test.go`
- [X] T082 [US4] Add failing tests: nil date field → `null` in json/yaml and empty cell in table/csv across Render in `internal/output/render_test.go`
- [X] T083 [US4] Add failing tests: changing only `opts.DateLayout` changes date strings in all four formats without changing other fields in `internal/output/render_test.go`

### Implementation for User Story 4

- [X] T090 [US4] Harden `NormalizeLayout` / `FormatDate` invalid-layout fallback in `internal/output/date.go`
- [X] T091 [US4] Ensure Render null-date path (json/yaml `null`, table/csv `""`) in `internal/output/render_json.go`, `internal/output/render_yaml.go`, `internal/output/render_csv.go`, `internal/output/render_table.go` (and shared helpers)
- [X] T092 [US4] Make US4 tests green in `internal/output/date_test.go` and `internal/output/render_test.go`

### Coverage gate for US4

- [X] T093 Run `make test` after US4 (Makefile)

**Checkpoint**: Date policy + null handling closed for DoD.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: godoc, deps, quickstart DoD, final coverage.

- [X] T100 [P] Add/verify exported godoc (no package-stutter) on Format, ResolveFormat, ColorEnabled, FormatDate, Render, RecordSet in `internal/output/*.go`
- [X] T101 [P] If `go-isatty` is imported directly, promote `github.com/mattn/go-isatty` to direct in `go.mod` / `go.sum` (only if needed; ColorEnabled tests may stay injectable without isatty)
- [X] T102 Validate `specs/006-output-rendering/quickstart.md` DoD checklist against `go test ./internal/output/...` and `make test`
- [X] T103 Confirm no new CLI command and no required `config show` migration (FR-009) — leave `internal/cli/config_show_render.go` unchanged unless a tiny non-DoD follow-up is explicitly chosen
- [X] T104 Final coverage gate: `make test` (Makefile) — no regression vs baseline

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 (Phase 3)**: After Foundational — MVP
- **US2 (Phase 4)**: After US1 (needs table/json renderers for ANSI asserts); ColorEnabled unit tests could start earlier but wiring needs table
- **US3 (Phase 5)**: After Foundational — **independent of US1/US2** (can parallel with US1 if staffed)
- **US4 (Phase 6)**: After US1 (needs Render path); builds on date helpers from US1
- **Polish (Phase 7)**: After desired stories complete

### User Story Dependencies

- **US1 (P1)**: After Foundational — no story deps — **MVP**
- **US2 (P1)**: After US1 renderers exist (for ANSI on table/machine formats); ColorEnabled itself is pure
- **US3 (P2)**: After Foundational only — parallelizable with US1
- **US4 (P2)**: After US1 (Render + FormatDate baseline)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Implementation → make story tests green → `make test` coverage gate
- No demo CLI / no entity columns

### Parallel Opportunities

- T002/T003; T012 vs T013 (after T011)
- US1: T020–T022 parallel; T033–T035 parallel after T032
- US2: T040/T041 parallel
- US3: T060/T061 parallel; entire US3 parallel with US1 after Phase 2
- US4: T080/T081 parallel
- Polish: T100/T101 parallel

---

## Parallel Example: User Story 1

```bash
# Tests first (fail):
Task: "T020 harness all formats in internal/output/render_test.go"
Task: "T021 empty RecordSet in internal/output/render_test.go"
Task: "T022 json/yaml root array in internal/output/render_test.go"

# After dispatcher exists, parallel renderers:
Task: "T033 render_json.go"
Task: "T034 render_yaml.go"
Task: "T035 render_csv.go"
```

---

## Parallel Example: User Story 3 (alongside US1)

```bash
# After Phase 2, independent of Render:
Task: "T060–T062 ResolveFormat failing tests in internal/output/format_test.go"
Task: "T070 Implement ResolveFormat in internal/output/format.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 Setup + Phase 2 Foundational
2. Phase 3 US1 (Render all formats)
3. **STOP and VALIDATE**: `go test ./internal/output/...` + `make test`
4. Then US2 → US3 → US4 → Polish

### Incremental Delivery

1. Setup + Foundational → package skeleton
2. US1 → cross-format MVP
3. US2 → pipe-safe color policy
4. US3 → format precedence (or parallel after foundation)
5. US4 → date_format / null dates
6. Polish → quickstart DoD

### Parallel Team Strategy

1. Team finishes Setup + Foundational
2. Dev A: US1 → US2 → US4
3. Dev B: US3 (ResolveFormat) in parallel with US1
4. Merge + Polish

---

## Notes

- [P] = different files, no incomplete-task deps
- DoD = `internal/output` unit/harness only (FR-009 / clarify Q5)
- CSV headers use **Column.Key** (contract); table headers use Title else Key
- `api.ParseDate` (F05) is input parsing — do not conflate with `output.FormatDate`
- Avoid migrating `config show` in this feature
- Commit after each task or logical group
- Stop at checkpoints to validate independently
