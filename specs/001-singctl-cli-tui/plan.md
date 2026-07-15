# Implementation Plan: singctl CLI/TUI

**Branch**: `001-singctl-cli-tui` | **Date**: 2026-07-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-singctl-cli-tui/spec.md`

## Summary

Консольный клиент `singctl` (репозиторий `sa-cli`) для SingularityApp REST API v2: CLI на Cobra + TUI на Bubble Tea, общий слой API. HTTP-клиент и модели генерируются из OpenAPI; ручной код — адаптеры, команды, UI.

## Technical Context

**Language/Version**: Go (актуальный stable toolchain, module path репозитория)

**Primary Dependencies**: `cobra`, `viper`, `bubbletea`/`lipgloss`/`bubbles`, `oapi-codegen` (generated client), `tablewriter`

**Storage**: локальные YAML-конфиг и опциональный файловый кэш (`~/.cache/singctl/`); без БД

**Testing**: `go test`, httptest/мок для API-адаптеров, bubbletea model-тесты

**Target Platform**: Linux / macOS / Windows (CLI); TUI — терминалы с ANSI

**Project Type**: CLI + TUI single-binary application

**Performance Goals**: интерактивный отклик CLI < ощутимой задержки сети; TUI 60fps не требуется, UI не блокируется на HTTP (async commands)

**Constraints**: только возможности публичного API; нет realtime; токен только локально; codegen обязателен

**Scale/Scope**: один пользовательский аккаунт на процесс; сущности OpenAPI v2 (~10 ресурсов)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status |
|---|---|
| Go single binary | PASS |
| OpenAPI-generated client | PASS (зафиксировано в plan/research/docs) |
| Shared client CLI+TUI | PASS (архитектура) |
| Scriptability (formats, exit codes) | PASS |
| Honest API boundaries | PASS (out-of-scope в spec) |
| Spec-driven artifacts present | PASS |

## Project Structure

### Documentation (this feature)

```text
specs/001-singctl-cli-tui/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/openapi.yaml
├── checklists/requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/singctl/
  main.go
internal/
  apiclient/          # generated OpenAPI client (*.gen.go)
  api/                # wrappers: auth, retry, error mapping, facades
  config/             # viper load/save, paths, validate
  output/             # table/json/yaml/csv renderers
  cache/              # optional TTL cache for projects/tags
  quickadd/           # parser for quick-add syntax
  tui/                # bubbletea app + views
    app.go
    tasks.go
    projects.go
    habits.go
    tags.go
    time.go
cmd/
  # cobra command tree may live under internal/cmd or cmd/singctl/commands
api/
  oapi-codegen.yaml   # generator config
docs/
  api/openapi.yaml
  openapi-codegen.md
config.yaml.example
```

**Structure Decision**: один Go-модуль в корне; `cmd/singctl` — entrypoint; `internal/*` — вся логика; сгенерированный клиент изолирован в `internal/apiclient`.

## Complexity Tracking

Нет нарушений constitution, требующих justification.
