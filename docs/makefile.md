# Makefile — регулярные задачи разработки

Параметры окружения читаются из `.env` (см. `.env.example`). Файл `.env` не коммитится.

```bash
cp .env.example .env   # один раз; заполните секреты локально
make help
```

## Таргеты

| Target | Назначение |
|---|---|
| `help` | Список таргетов |
| `openapi-fetch` | Скачать OpenAPI JSON/YAML в `docs/api/` |
| `api-coverage-check` | Проверить, что в снимке 51 operation (актуальный v2) |
| `generate` | Генерация Go-клиента из OpenAPI (после появления `api/oapi-codegen.yaml`) |
| `build` | Сборка `singctl` |
| `test` | `go test ./...` |
| `smoke` | Лёгкий smoke против API (нужен `SINGCTL_TOKEN` в `.env`) |

Подробности codegen: [openapi-codegen.md](./openapi-codegen.md). Принцип: constitution §VIII.
