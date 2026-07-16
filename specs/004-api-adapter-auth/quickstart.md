# Quickstart: API Adapter & Auth (F04)

**Feature**: `004-api-adapter-auth` | **Date**: 2026-07-16

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [api-session.md](./contracts/api-session.md), [cli-config-validate.md](./contracts/cli-config-validate.md).

---

## Prerequisites

- Репозиторий с F02 (config) и F03 (`internal/apiclient/client.gen.go`)
- Go toolchain; зависимости модуля уже в `go.mod`
- Канон тестов: `make test`

---

## 1. Unit tests (adapter + validate wiring)

```bash
make test
```

**Expected**:
- exit 0
- зелёные тесты `internal/api` (happy path Bearer, fail-fast empty token, non-2xx → `HTTPError`, timeout/connectivity as covered)
- обновлённые тесты `config validate` без требования текста «заглушка» при успешном токене+моке

---

## 2. Manual remote validate (optional)

Если есть реальный токен в локальном конфиге / `.env` (не коммитить):

```bash
singctl config set-token '<your-token>'   # уже может быть сделано
singctl config show                       # токен замаскирован
singctl config validate
```

**Expected**:
- exit 0 и сообщение об успешной **удалённой** проверке при валидном токене и доступном API
- при неверном токене — ненулевой exit, без «удалённо OK»

Для изолированной ручной проверки без prod API поднимите любой mock, отдающий `200` на `GET /v2/project`, пропишите его URL в `api.base_url`, затем `config validate`.

---

## 3. Negative checks (quick)

```bash
# без токена (чистый XDG)
env XDG_CONFIG_HOME=/tmp/singctl-empty-$$ singctl config validate
# expect: nonzero + set-token hint
```

---

## Definition of done (checklist)

- [x] `internal/api` существует: factory + Bearer + timeout + `HTTPError` + probe
- [x] Unit-тесты с `httptest` (минимум 1 happy path + 1 non-2xx)
- [x] `singctl config validate` использует адаптер, не stub
- [x] Нет CLI entity CRUD и нет retry-логики
- [x] `make test` зелёный; фикстуры токенов gitleaks-safe
