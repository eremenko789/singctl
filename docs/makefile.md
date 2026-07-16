# Makefile — регулярные задачи разработки

Параметры окружения читаются из `.env` (см. `.env.example`). Файл `.env` не коммитится.

```bash
cp .env.example .env   # один раз; заполните секреты локально
make help
```

## Таргеты

| Target | Назначение |
|---|---|
| `help` | Список таргетов; упоминает рекомендуемый порядок OpenAPI |
| `openapi-fetch` | Атомарно скачать OpenAPI JSON/YAML в `docs/api/` (`scripts/openapi_fetch.sh`) |
| `api-coverage-check` | `scripts/count_openapi_ops.py`: ops == `EXPECTED_API_OPS` (51) + файл `docs/api/coverage.md` |
| `generate` | Генерация Go-клиента из `docs/api/openapi.yaml` + `api/oapi-codegen.yaml` |
| `build` | Сборка `singctl` |
| `test` | `go test` с coverage (без `*.gen.go`) + python/shell контракты OpenAPI |
| `pre-commit` | Все hooks из `.pre-commit-config.yaml` по всем файлам (`pre-commit run --all-files`) |
| `smoke` | Лёгкий smoke против API (нужен `SINGCTL_TOKEN` в `.env`) |

OpenAPI-таргеты **независимы**. Рекомендуемый порядок: `openapi-fetch` → `api-coverage-check` → `generate`.

Подробности codegen: [openapi-codegen.md](./openapi-codegen.md). Принцип: constitution §VIII.
