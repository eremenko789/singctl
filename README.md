# sa-cli (`singctl`)

CLI и TUI-клиент для [SingularityApp REST API](https://singularity-app.ru/wiki/api/).

> Статус: спецификация и план (GitHub Spec Kit). Реализация кода — по `specs/001-singctl-cli-tui/tasks.md`.

## Документация

| Путь | Содержание |
|---|---|
| [`docs/`](./docs/README.md) | Индекс docs, OpenAPI, codegen |
| [`docs/tz/singularityapp-cli-tui-tz.md`](./docs/tz/singularityapp-cli-tui-tz.md) | Исходное ТЗ |
| [`specs/001-singctl-cli-tui/`](./specs/001-singctl-cli-tui/) | Spec Kit: spec → plan → tasks |
| [`.specify/memory/constitution.md`](./.specify/memory/constitution.md) | Принципы проекта |

## Стек (зафиксировано)

- **Go** — один бинарник `singctl`
- **Cobra / Viper** — CLI и конфиг
- **Bubble Tea** — TUI
- **OpenAPI codegen** — клиент API (`docs/openapi-codegen.md`)

## API

- Swagger UI: https://api.singularity-app.com/v2/api
- OpenAPI: https://api.singularity-app.com/v2/api-json
- Локальный снимок: [`docs/api/openapi.yaml`](./docs/api/openapi.yaml)

Токен создаётся в [личном кабинете](https://me.singularity-app.com).

## Spec-Driven Development

Репозиторий следует [GitHub Spec Kit](https://github.com/github/spec-kit):

1. Constitution → 2. Specify → 3. Plan → 4. Tasks → 5. Implement

См. [`docs/spec-kit.md`](./docs/spec-kit.md).

## Лицензия

См. [LICENSE](./LICENSE).
