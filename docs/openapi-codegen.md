# OpenAPI: обновление снимка и генерация Go-клиента

Повторяемые шаги запускаются через **Makefile** (constitution §VIII). Параметры — из `.env` (см. `.env.example`).

## Источники

| Что | URL |
|---|---|
| Wiki | https://singularity-app.ru/wiki/api/ |
| Swagger UI | https://api.singularity-app.com/v2/api |
| OpenAPI JSON | https://api.singularity-app.com/v2/api-json |
| OpenAPI YAML | https://api.singularity-app.com/v2/api-yaml |

Локально: `docs/api/openapi.yaml`, `docs/api/openapi.json`.
Матрица покрытия CLI: `docs/api/coverage.md`.

Base URL: `https://api.singularity-app.com`
Auth: `Authorization: Bearer <token>` (OpenAPI scheme `rest-token`).

## Команды

Таргеты **независимы** (нет Make-prerequisites друг на друга). Рекомендуемый порядок:

```text
make openapi-fetch → make api-coverage-check → make generate
```

```bash
cp .env.example .env          # при необходимости задать API_BASE_URL / SINGCTL_TOKEN
make openapi-fetch            # атомарно обновить пару JSON+YAML
make api-coverage-check       # ops == EXPECTED_API_OPS (51) + наличие coverage.md
make generate                 # oapi-codegen → internal/apiclient/client.gen.go
make smoke                    # нужен SINGCTL_TOKEN
```

`openapi-fetch` скачивает оба файла во временные пути и только затем заменяет целевые; при сбое второго скачивания прежняя пара снимка не рассинхронизируется.

`api-coverage-check` сверяет **только** число HTTP operations в JSON с `EXPECTED_API_OPS` и наличие файла `docs/api/coverage.md` (без парсинга строк матрицы / operationId).

Эквивалент вручную (не предпочтительно; без атомарности):

```bash
curl -fsSL "https://api.singularity-app.com/v2/api-json" -o docs/api/openapi.json
curl -fsSL "https://api.singularity-app.com/v2/api-yaml" -o docs/api/openapi.yaml
```

Предпочтительный helper: `scripts/openapi_fetch.sh`.

## Генерация Go-клиента

Генератор: [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen).

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

Конфиг `api/oapi-codegen.yaml`:

```yaml
package: apiclient
generate:
  models: true
  client: true
output: internal/apiclient/client.gen.go
```

Затем `make generate` (офлайн, по закоммиченному YAML-снимку).

Ручные DTO/HTTP CRUD запрещены (constitution). Допускаются только адаптеры над codegen.

## Что коммитить

| Файл | Коммитить? |
|---|---|
| `docs/api/openapi.yaml` / `.json` | Да |
| `docs/api/coverage.md` | Да (обновлять вместе со снимком) |
| `internal/apiclient/*.gen.go` | Да |
| `api/oapi-codegen.yaml` | Да |
| `.env` / токены | Нет |

## История снимка

Первый снимок: **2026-07-15**, OpenAPI `3.0.0`, API version `2.0`, **51** HTTP operation.
