# Implementation Plan: CLI Skeleton

**Branch**: `001-cli-skeleton` | **Date**: 2026-07-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-cli-skeleton/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

F01 создаёт минимальный исполняемый бинарник `singctl` на Go: корневая команда Cobra с глобальными флагами (`--config`, `--token`, `--output`/`-o`, `--no-color`, `--debug`), русскоязычные `--help` и ошибки парсера, команда `version` и флаг `--version` с одинаковым выводом (имя, версия, commit/date). Entity-команды, TUI и сеть вне скоупа; структура команд и Make-сборка (`./cmd/singctl`) закладывают расширение для F02+.

## Technical Context

**Language/Version**: Go (stable toolchain ≥ 1.22; модуль `github.com/eremenko789/singctl`)

**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/spf13/viper` (флаги + заготовка привязки; загрузка конфиг-файла — F02)

**Storage**: N/A (F01 только парсит `--config`/`--token`; чтение YAML — F02)

**Testing**: `go test ./...` (unit/contract-тесты CLI через `cobra` Execute + захват stdout/stderr); `make test` / `make build`

**Target Platform**: кроссплатформенный CLI (darwin/linux/windows); локальная валидация на macOS/Linux

**Project Type**: CLI (single binary)

**Performance Goals**: `singctl --help` и `version`/`--version` завершаются за < 2 с на чистой машине без сети (SC-001/SC-002)

**Constraints**: без сетевых вызовов и без entity/TUI-команд; русский UX для help/флагов/ошибок парсера; недопустимый `--output` блокирует help/version; ненулевой exit без аргументов

**Scale/Scope**: один бинарник; ~корень + `version`; 5 глобальных флагов; без API-слоя

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart через Spec Kit |
| G2 | Go Single Binary | PASS | один модуль, `make build` → `bin/singctl` |
| G3 | OpenAPI-Generated API Client | PASS (N/A) | HTTP/codegen вне F01; код клиента не пишется |
| G4 | Shared Client for CLI and TUI | PASS | CLI на Cobra; API/TUI не добавляются; точки расширения в `internal/` |
| G5 | Scriptability First | PASS | стабильные exit codes; контракт `--output`; stdout для version, stderr для ошибок |
| G6 | Honest API Boundaries | PASS | справка не обещает entity/TUI/API |
| G7 | Security of Credentials | PASS | `--token` только парсится; не пишется в файл/логи в F01 |
| G8 | Makefile + `.env` | PASS | сборка/тесты через существующие `make build` / `make test`; `.env` не требуется для F01 |

**Post-design re-check (Phase 1)**: все гейты PASS; violations нет → Complexity Tracking пуст.

## Project Structure

### Documentation (this feature)

```text
specs/001-cli-skeleton/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/singctl/
└── main.go                 # package main → internal/cli.Execute

internal/
├── buildinfo/
│   └── buildinfo.go        # Version, Commit, Date (ldflags / placeholders)
└── cli/
    ├── root.go             # root cobra.Command, глобальные флаги, RunE без args
    ├── version.go          # subcommand version
    └── root_test.go        # help, version, flags, invalid output, bare invoke

go.mod
go.sum
Makefile                    # уже есть: build → ./cmd/singctl, test → go test ./...
.bin/                       # артефакт сборки (gitignore), не исходники
```

**Structure Decision**: Entrypoint `cmd/singctl` (совпадает с текущим `Makefile`). Дерево Cobra и глобальные опции — в `internal/cli` (расширяемо без ломки `main`). Метаданные сборки — `internal/buildinfo`. Каталоги `internal/api`, `internal/apiclient`, `internal/config`, `internal/tui` и `cmd/...` entity-команд **не создаются** в F01 (появляются в зависимых фичах). Пакетный layout с `src/` не используется.

## Complexity Tracking

> Нет нарушений Constitution Check — таблица не заполняется.
