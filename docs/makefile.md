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
| `test` | `go test` **с coverage**; MUST не допускать падения покрытия (constitution IX) |
| `pre-commit` | Все hooks из `.pre-commit-config.yaml` по всем файлам (`pre-commit run --all-files`) |
| `smoke` | Лёгкий smoke против API (нужен `SINGCTL_TOKEN` в `.env`) |

Подробности codegen: [openapi-codegen.md](./openapi-codegen.md). Принцип: constitution §VIII.
