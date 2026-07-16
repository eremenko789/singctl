# Quickstart: OpenAPI Codegen Pipeline (F03)

**Feature**: `003-openapi-codegen`

Проверка пайплайна разработчика end-to-end. CLI CRUD и auth-адаптер не нужны. Контракт таргетов: [contracts/make-openapi.md](./contracts/make-openapi.md).

---
## Preconditions

1. Репозиторий на ветке фичи; есть `docs/api/openapi.yaml` и `docs/api/openapi.json` (или будет `make openapi-fetch`).
2. Установлен генератор:

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
command -v oapi-codegen
```

3. (Опционально) `cp .env.example .env` — для F03 токен не обязателен.

---
## Step 1: Coverage-check на каноническом снимке

```bash
make api-coverage-check
```

Ожидаемо:
- exit 0;
- в выводе `operations=51 expected=51` (или актуальное `EXPECTED_API_OPS`);
- сообщение о наличии `docs/api/coverage.md`.

Негатив (опционально):

```bash
EXPECTED_API_OPS=0 make api-coverage-check; echo exit:$?
```

Ожидаемо: exit ≠ 0.

---
## Step 2: Конфиг генерации и generate (офлайн)

Убедиться, что существует `api/oapi-codegen.yaml` (package `apiclient`, output `internal/apiclient/client.gen.go`).

```bash
make generate
```

Ожидаемо:
- exit 0;
- файл `internal/apiclient/client.gen.go` существует;
- пакет компилируется:

```bash
go build ./internal/apiclient/...
```

Негатив: временно переименовать конфиг и убедиться, что `make generate` ≠ 0 с понятной ошибкой.

---
## Step 3 (опционально, нужна сеть): Refresh snapshot

```bash
make openapi-fetch
make api-coverage-check
make generate
```

Ожидаемо: все три с exit 0 при неизменном числе ops; при изменении upstream ops — coverage-check падает, пока не обновлены `EXPECTED_API_OPS` и матрица осознанно.

---
## Step 4: DoD checklist (ревью)

В git (staged/committed) должны быть:
- [ ] `api/oapi-codegen.yaml`
- [ ] `docs/api/openapi.yaml` / `openapi.json`
- [ ] `docs/api/coverage.md`
- [ ] `internal/apiclient/*.gen.go`

Не должны попасть: `.env`, реальные токены.

Docs синхронизированы: `docs/openapi-codegen.md`, `docs/makefile.md` (рекомендуемый порядок; независимость таргетов; строгость check).

---
## Notes

- `make generate` **не** требует зелёного `api-coverage-check`.
- Автоматический drift/no-diff gate после generate в F03 не проверяется.
- Smoke API (`make smoke`) — вне acceptance F03.
