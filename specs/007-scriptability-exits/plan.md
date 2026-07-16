# Implementation Plan: Scriptability & Exit Codes (F07)

**Branch**: `007-scriptability-exits` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/007-scriptability-exits/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F07. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F07 закрепляет **Scriptability First** (constitution V / ТЗ §10) как продуктовый контракт поверх уже сделанных F05 (`ExitCode` 0/1/2/3) и F06 (форматы, pipe без ANSI):

- публичная таблица exit codes в **`docs/scriptability.md`** + краткое упоминание в корневом `singctl --help`;
- жёсткое разделение потоков на DoD-поверхности: успех → data в stdout, stderr пуст; ошибка → stderr + пустой stdout + код по таблице;
- misuse CLI (неизвестный флаг/команда, неверный `--output`) → exit **`1`** (не `2`);
- контракт четырёх pipe-сценариев ТЗ §10 (проверяемо сейчас через F06 fixture / существующие команды, entity E2E — F08+);
- non-interactive: нет блокирующих prompt’ов; stdin pipe не подвешивает текущие команды.

DoD без новых пользовательских команд: `config show`, `config validate`, misuse-флаг, docs+`--help`, reuse F06 ANSI/fixture.

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: существующие `internal/cli` (cobra), `cli.ExitCode` (F05), `internal/output` (F06), `cmd/singctl/main.go`; без новых внешних библиотек.

**Storage**: N/A (документация + поведение process boundary).

**Testing**: TDD — сначала failing тесты streams/misuse/help/docs-presence; затем правки CLI/docs. `executeForTest` для stdout/stderr/err; `ExitCode(err)` для числовых кодов; reuse `internal/output` pipe/ANSI tests. `make test` + coverage; токены только `test-token-…`.

**Target Platform**: CLI `singctl` (macOS/Linux).

**Project Type**: documentation + thin CLI contract hardening внутри Go monorepo (без entity CRUD).

**Performance Goals**: unit/CLI harness быстрый; без live API smoke в DoD.

**Constraints**: constitution V/VII/IX; не менять числовую семантику exit F05; не вводить новых команд; не ослаблять F06 color/format; стриминг partial-output вне scope.

**Scale/Scope**: `docs/scriptability.md` + ссылка в `docs/README.md`; правка Long/`--help` root; усиление тестов `config show` / `config validate` / root misuse; контракты pipe/streams; минимальные фиксы печати, если тесты найдут утечки в stdout.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go + docs; один бинарь |
| G3 | OpenAPI-Generated API Client | PASS N/A | HTTP/codegen не затрагивается |
| G4 | Shared Client for CLI and TUI | PASS | taxonomy/exit остаются в api/cli; F07 не дублирует бизнес-логику в UI |
| G5 | Scriptability First | PASS | ядро фичи: exit docs, stdout/stderr, pipe contract |
| G6 | Honest API Boundaries | PASS | pipe-примеры с entity помечены F08+; без ложных обещаний live CRUD |
| G7 | Security of Credentials | PASS | тесты с `test-token-…`; docs без реальных секретов |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты до/вместе с правками CLI; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/007-scriptability-exits/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── exit-codes-public.md      # публичная таблица + misuse → 1; ссылка на F05 helper
│   ├── stream-separation.md      # stdout/stderr правила DoD
│   └── pipe-scenarios.md         # ТЗ §10 ×4 → свойства / статус F07 vs F08+
└── tasks.md                      # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
docs/
├── scriptability.md              # NEW: полная таблица exit + pipe/streams overview
└── README.md                     # UPDATE: строка в указателе → scriptability.md

internal/cli/
├── root.go                       # UPDATE: Long/`Example` или блок help с кодами 0/1/2/3
├── root_test.go                  # UPDATE: --help mentions exit codes; misuse → ExitCode 1 + empty stdout
├── exit.go                       # EXISTS (F05) — KEEP as SoT; MAY godoc cross-link docs
├── exit_test.go                  # EXISTS — MAY extend misuse classification note
├── config_show_test.go           # UPDATE: success → stderr empty; error → stdout empty
├── config_validate_test.go       # UPDATE: success stderr empty; errors stdout empty + ExitCode matrix
└── …                             # фикс печати только если тесты красные

internal/output/
└── *_test.go                     # REUSE: pipe/ANSI / json-csv fixture (SC-005); без обязательных правок API

cmd/singctl/
└── main.go                       # EXISTS: os.Exit(cli.ExitCode(err)) — KEEP
```

**Structure Decision**: Числовые коды остаются в `cli.ExitCode` (F05). F07 добавляет **user-facing docs + help**, ужесточает **stream assertions** на DoD-командах и фиксирует **pipe-сценарии ТЗ** контрактом. Entity-пайплайны — документальный контракт до F08+.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
- путь docs = `docs/scriptability.md` + индекс в `docs/README.md`;
- help: краткая таблица/строка в Long root (не дублировать весь docs);
- misuse → 1 через существующий `ExitCode` (не Kind Config);
- empty stdout on error / empty stderr on success — test-first на show/validate;
- pipe §10 — contract matrix, не live task CLI;
- stdin non-block — SetIn / closed stdin в harness.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — Exit Code Contract, Stream Policy, Pipe Scenario, Non-Interactive Context
- `contracts/exit-codes-public.md` — таблица 0/1/2/3 + misuse
- `contracts/stream-separation.md` — правила потоков DoD
- `contracts/pipe-scenarios.md` — четыре примера ТЗ §10
- `quickstart.md` — `make test` + ручная сверка docs/help

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. docs presence + `--help` substring tests → docs + root Long;
2. misuse ExitCode 1 + empty stdout tests → confirm/fix Execute path;
3. `config show` / `config validate` stream tests (empty stderr success / empty stdout error) → fixes;
4. optional stdin-pipe non-block test;
5. contract files already from plan; ensure tests map SC-*;
6. godoc/cross-links; `make test` coverage.
