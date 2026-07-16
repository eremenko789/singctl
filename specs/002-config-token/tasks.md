# Tasks: Config & Token Storage (F02)

**Input**: Design documents from `/specs/002-config-token/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: MANDATORY (constitution IX — TDD). Для каждой user story, которая меняет production-code, сначала добавляются тесты (они должны падать), затем минимальная реализация делает тесты зелёными. Покрытие не должно регрессировать (`make test`).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: which user story phase this task belongs to (`[US1]..[US5]`); setup/foundation/polish — без story label

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: dependencies и каркас каталогов под F02 (поверх существующего F01 CLI).

- [X] T001 [P] Create folders `internal/config/` and `internal/cli/` helpers for F02 (`internal/config/`, `internal/cli/`)
- [X] T002 [P] Add direct dependency for table rendering `github.com/olekukonko/tablewriter` in `go.mod` / `go.sum` (Makefile/go.mod)
- [X] T003 [P] Ensure YAML support is usable via explicit import path (tidy modules) in `go.mod` / `go.sum` (`go.mod`, `go.sum`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: общий резолвинг конфигурации, нормализация токена, маскирование, и безопасные IO без CLI-слоя.

⚠️ CRITICAL: No user story work can begin until this phase is complete.

### Tests for Foundational Phase (REQUIRED — TDD) ⚠️

> **NOTE**: write these tests FIRST; ensure they fail before implementation. Coverage must not regress.

- [X] T010 [P] Add failing unit tests for config path resolution priority (`--config` vs `$XDG_CONFIG_HOME` vs `~/.config` vs `./.singctl.yaml`) in `internal/config/resolver_test.go`
- [X] T011 [P] Add failing unit tests for token normalization (reject `set-token` input starting with `Bearer` + whitespace) in `internal/config/masking_test.go`
- [X] T012 [P] Add failing unit tests for token masking output format (`first4 + **** + last4`, short tokens -> `****`) in `internal/config/masking_test.go`
- [X] T013 [P] Add failing unit tests for config load/save YAML round-trip for known fields in `internal/config/io_test.go`

### Implementation for Foundational Phase

- [X] T020 [P] Implement config path resolution logic in `internal/config/resolver.go`
- [X] T021 [P] Implement config document types + defaults in `internal/config/types.go`
- [X] T022 [P] Implement YAML load/save and file create-dir behavior in `internal/config/io.go`
- [X] T023 [P] Implement token normalization and masking helpers in `internal/config/masking.go`
- [X] T024 [P] Implement effective settings builder (effective token applies `--token` runtime override without changing file) in `internal/config/effective.go`
- [X] T025 [P] Add coverage-gated tests update until all foundational tests pass (`internal/config/*_test.go`)

### Coverage gate for Foundation

- [X] T026 Run coverage-gated tests via `make test` after Foundation (Makefile)

**Checkpoint**: Foundation ready — user stories can now be implemented and tested independently.

---

## Phase 3: User Story 1 — Save API token locally (Priority: P1) 🎯 MVP

**Goal**: Команда `singctl config set-token <TOKEN>` сохраняет токен в `api.token` эффективного файла конфигурации (с созданием каталогов/файла при необходимости) в виде "голого" токена без `Bearer`.

**Independent Test**: В изолированной директории выполнить `singctl config set-token "$TOKEN"` и затем `singctl config show` (пока может существовать минимальная реализация show) — убедиться, что полный токен не печатается.

### Tests for User Story 1 (REQUIRED — TDD) ⚠️

- [X] T030 [US1] Add failing command-level tests for `singctl config set-token` writes bare token to resolved config path in `internal/cli/config_cmd_test.go`
- [X] T031 [US1] Add failing command-level tests for `singctl config set-token` rejecting input that starts with `Bearer` (with whitespace after) and not modifying config file in `internal/cli/config_cmd_test.go`
- [X] T032 [US1] Add failing command-level tests for `--config` precedence: token must be written only to explicit `--config` path in `internal/cli/config_cmd_test.go`

### Implementation for User Story 1

- [X] T040 [US1] Implement `config set-token` subcommand wiring and argument validation in `internal/cli/config_set_token.go`
- [X] T041 [US1] Use `internal/config/resolver.go` + `internal/config/io.go` to create/update YAML file in `internal/config/` and call from `internal/cli/config_set_token.go`
- [X] T042 [US1] Enforce token input rule ("bare token only") via `internal/config/masking.go` from `internal/cli/config_set_token.go`
- [X] T043 [US1] Register `config` root command and `set-token` subcommand in `internal/cli/root.go`
- [X] T044 [US1] Make `singctl config --help` include RU descriptions for `set-token` in `internal/cli/config_cmd.go`

### Coverage gate for US1

- [X] T045 Run coverage-gated tests via `make test` after US1 (Makefile)

---

## Phase 4: User Story 2 — Inspect effective configuration (Priority: P1)

**Goal**: Команда `singctl config show` выводит effective конфигурацию после резолвинга источников и применения runtime-override `--token`, маскируя `api.token` по формату из spec.

### Tests for User Story 2 (REQUIRED — TDD) ⚠️

- [X] T050 [US2] Add failing command-level tests for `singctl config show` with no config: non-zero exit + safe RU message (no panic, no token leak) in `internal/cli/config_show_test.go`
- [X] T051 [US2] Add failing command-level tests for `singctl config show` masking rule and token selection with `--token` override precedence in `internal/cli/config_show_test.go`
- [X] T052 [US2] Add failing command-level tests for `singctl config show` output format defaults: without `-o` -> human YAML, with `-o json` -> JSON in `internal/cli/config_show_test.go`
- [X] T053 [US2] Add failing tests for `singctl config show -o csv` and `-o table` ensuring token remains masked in `internal/cli/config_show_test.go`

### Implementation for User Story 2

- [X] T060 [US2] Implement `config show` subcommand behavior in `internal/cli/config_show.go`
- [X] T061 [US2] Implement output rendering helpers for `--output/-o` formats (yaml/json/csv/table) in `internal/cli/config_show_render.go`
- [X] T062 [US2] Respect output rule: when `--output/-o` flag not explicitly set -> default to YAML (derive from Cobra flag `.Changed`) in `internal/cli/config_show.go`
- [X] T063 [US2] Ensure `api.token` is always masked in all output formats using `internal/config/masking.go` in `internal/cli/config_show_render.go`
- [X] T064 [US2] Register `config show` subcommand in `internal/cli/config_cmd.go`

### Coverage gate for US2

- [X] T065 Run coverage-gated tests via `make test` after US2 (Makefile)

---

## Phase 5: User Story 3 — Resolve configuration from multiple locations (Priority: P1)

**Goal**: Команды `config show` и `config set-token` (и `config set` далее) используют единое правило приоритета источников конфигурации, как в spec: `--config` → XDG → `~/.config` → `./.singctl.yaml`.

### Tests for User Story 3 (REQUIRED — TDD) ⚠️

- [X] T070 [US3] Add failing integration tests that create multiple config files and verify `config show` reads the highest-priority existing file in `internal/cli/config_resolve_integration_test.go`
- [X] T071 [US3] Add failing integration tests that verify `config set-token` updates only the effective file when no `--config` is provided in `internal/cli/config_resolve_integration_test.go`

### Implementation for User Story 3

- [X] T080 [US3] Verify and refactor command implementations to use one shared effective-config resolver entrypoint (`internal/config/effective.go`) in `internal/cli/config_show.go` and `internal/cli/config_set_token.go`
- [X] T081 [US3] Ensure write path creation for "no config exists" uses default canonical file (`XDG` or `~/.config`) and not `./.singctl.yaml` (spec FR-005) in `internal/config/resolver.go`
- [X] T082 [US3] Make US3 integration tests pass without changing public CLI contract in `internal/cli/config_*`

### Coverage gate for US3

- [X] T083 Run coverage-gated tests via `make test` after US3 (Makefile)

---

## Phase 6: User Story 4 — Set arbitrary config keys (Priority: P2)

**Goal**: Команда `singctl config set <key> <value>` обновляет допустимые поля конфигурации по dotted path, валидирует тип/enum, и сохраняет обратно YAML без повреждения существующей конфигурации при ошибках.

### Tests for User Story 4 (REQUIRED — TDD) ⚠️

- [X] T090 [US4] Add failing tests for `config set` accepting valid keys/values and persisting them in effective config file in `internal/cli/config_set_test.go`
- [X] T091 [US4] Add failing tests for `config set` rejecting unknown keys with non-zero exit and without modifying file in `internal/cli/config_set_test.go`
- [X] T092 [US4] Add failing tests for `config set` rejecting invalid values by type/enum (e.g., output.format) and preserving existing known fields in `internal/cli/config_set_test.go`
- [X] T093 [US4] Add failing tests for `--config` precedence: `config set` writes only to explicit `--config` path in `internal/cli/config_set_test.go`

### Implementation for User Story 4

- [X] T100 [US4] Implement key/value parsing + validation and dotted-path mapping for supported schema keys in `internal/config/setter.go`
- [X] T101 [US4] Implement `config set` command wiring in `internal/cli/config_set.go`
- [X] T102 [US4] Ensure `config set api.token <TOKEN>` applies same bare-token rule as `set-token` via `internal/config/masking.go` in `internal/config/setter.go`
- [X] T103 [US4] Register `config set` subcommand in `internal/cli/config_cmd.go`

### Coverage gate for US4

- [X] T104 Run coverage-gated tests via `make test` after US4 (Makefile)

---

## Phase 7: User Story 5 — Validate API connectivity (Priority: P2)

**Goal**: Команда `singctl config validate` проверяет наличие токена (файл или runtime-override) и возвращает честный результат в режиме заглушки: не выполняет CRUD и не обещает удалённый OK до появления HTTP-клиента.

### Tests for User Story 5 (REQUIRED — TDD) ⚠️

- [X] T110 [US5] Add failing tests for `config validate` when token missing: non-zero exit + RU hint to `config set-token` (no token output) in `internal/cli/config_validate_test.go`
- [X] T111 [US5] Add failing tests for `config validate` when token present: exit code 0 and message indicates stub/local check (no token leak) in `internal/cli/config_validate_test.go`
- [X] T112 [US5] Add failing tests for runtime override precedence: `config validate --token <TOKEN>` uses override token in `internal/cli/config_validate_test.go`

### Implementation for User Story 5

- [X] T120 [US5] Implement `config validate` subcommand stub behavior in `internal/cli/config_validate.go`
- [X] T121 [US5] Ensure validate reads effective token from `internal/config/effective.go` and never logs the full token in `internal/cli/config_validate.go`
- [X] T122 [US5] Register `config validate` subcommand in `internal/cli/config_cmd.go`

### Coverage gate for US5

- [X] T123 Run coverage-gated tests via `make test` after US5 (Makefile)

---

## Phase N: Polish & Cross-Cutting Concerns (Final)

**Purpose**: единый UX, помощь, безопасность форматов и quickstart consistency.

- [X] T130 Add/adjust Russian error/help texts for all `config` subcommands in `internal/cli/config_*.go` to match spec FR-014
- [X] T131 [P] Add edge-case tests from spec Edge Cases to `internal/config/*_test.go` and `internal/cli/*_test.go` (token empty, file YAML invalid, I/O errors) where applicable
- [X] T132 Ensure quickstart scenarios are consistent with actual behavior described in `specs/002-config-token/quickstart.md`
- [X] T133 [P] Final coverage gate for feature via `make test` (Makefile)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - can start immediately
- **Phase 2 (Foundational)**: Blocks all user stories
- **Phase 3+ (User Stories)**: All depend on Foundation completion
- **Final Phase**: Depends on all desired user stories being complete

### Within each User Story

- Tests must be written and failing before implementation (TDD — constitution IX)
- Models/helpers before CLI wiring
- Use shared effective-resolver to avoid divergence
