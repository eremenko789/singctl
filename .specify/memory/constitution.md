# sa-cli (singctl) Constitution

## Core Principles

### I. Spec-Driven Development

Спецификации в `specs/` — источник истины для реализации. Код следует спецификации, а не наоборот. При изменении требований сначала обновляются `spec.md` / `plan.md` / `tasks.md`, затем код.

Исходное продуктовое ТЗ хранится в `docs/tz/` и раскладывается в артефакты Spec Kit; расхождения между ТЗ и актуальным API фиксируются в `research.md`.

### II. Go Single Binary

Язык реализации — Go. Цель поставки — один статически линкуемый CLI/TUI-бинарник (`singctl`) без обязательного рантайма. Зависимости минимальны и обоснованы.

### III. OpenAPI-Generated API Client (NON-NEGOTIABLE)

HTTP-обёртки вокруг SingularityApp REST API **не пишутся вручную**. Клиент и модели генерируются из OpenAPI (`docs/api/openapi.yaml` / upstream `https://api.singularity-app.com/v2/api-json`).

Ручной код допускается только как тонкий адаптер поверх сгенерированного клиента (auth, retry, mapping ошибок, CLI/TUI-фасады). Повторная генерация документируется в `docs/openapi-codegen.md`.

### IV. Shared Client for CLI and TUI

CLI (`cobra`) и TUI (`bubbletea`) используют один и тот же API-слой. Бизнес-логика не дублируется в UI.

### V. Scriptability First

CLI должен быть удобен для pipe/скриптов: стабильные exit codes, форматы `table|json|yaml|csv`, авто-отключение цвета в non-TTY, предсказуемый stdout/stderr.

### VI. Honest API Boundaries

Нельзя обещать функции, которых нет в API (webhooks/SSE, создание повторяющихся задач, совместные проекты, офлайн-синхронизация). Ограничения явно отражаются в UX и документации.

### VII. Security of Credentials

Токен API хранится только в локальном конфиге пользователя, маскируется в `config show` и логах. Токен не коммитится в репозиторий. В debug-логах секреты редактируются.

## Technology Constraints

| Область | Выбор |
|---|---|
| Язык | Go (актуальный stable) |
| CLI | `spf13/cobra` + `spf13/viper` |
| TUI | `charmbracelet/bubbletea`, `lipgloss`, `bubbles` |
| API client | генерация из OpenAPI (`oapi-codegen` или эквивалент) |
| HTTP | stdlib `net/http` / транспорт сгенерированного клиента |
| Вывод таблиц | `olekukonko/tablewriter` (или совместимый) |
| Релизы | `goreleaser` (Phase 3) |

Альтернативы (Python и т.п.) из исходного ТЗ **не используются** в этом репозитории.

## Quality Gates

- Unit-тесты для адаптеров API (мок HTTP) и парсеров CLI.
- TUI: model-тесты на `tea.Msg` там, где это практично.
- Integration-тесты с реальным API — только при наличии тестового токена и явном включении в CI.
- Каждая команда имеет `--help`.
- Перед реализацией фичи: актуальные `spec.md` → `plan.md` → `tasks.md`.

## Documentation Expectations

Повторяемые операции (обновление OpenAPI-снимка, codegen, сборка, релизы) описываются в `docs/`. README ссылается на Spec Kit и docs, а не дублирует длинные процедуры.

## Governance

Constitution имеет приоритет над ad-hoc решениями в коде и чатах. Изменения принципов фиксируются в этом файле с обновлением версии и даты.

**Version**: 1.0.0 | **Ratified**: 2026-07-15 | **Last Amended**: 2026-07-15
