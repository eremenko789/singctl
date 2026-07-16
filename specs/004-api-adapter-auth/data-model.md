# Data Model: API Adapter & Auth (F04)

**Feature**: `004-api-adapter-auth` | **Date**: 2026-07-16

Логические сущности адаптера (не схема БД). Типы конфига — из F02 (`internal/config`).

---

## Session

Сконфигурированный сеанс доступа к API.

| Field | Type / source | Rules |
|-------|---------------|-------|
| BaseURL | string (`api.base_url`) | Non-empty after trim; origin without `/v2`; trailing `/` добавляет codegen `NewClient` |
| Token | string (effective token) | Non-empty after trim; stored bare; Authorization = `Bearer ` + token (no double prefix) |
| Timeout | duration (`api.timeout`) | Valid `time.ParseDuration`; applied as `http.Client.Timeout` |
| Client | `*apiclient.ClientWithResponses` | Created by factory; not constructed in CLI/TUI |

**Factory validation (fail-fast)**:
- empty/whitespace token → error, no Session
- empty/whitespace base URL → error, no Session
- invalid timeout string → error, no Session

**Relationships**: Session *uses* Effective API Settings (F02); Session *owns* configured codegen client.

---

## Effective API Settings (input)

Вход фабрики (из `config.EffectiveSettings` / `Document` + overrides):

| Field | Meaning |
|-------|---------|
| `API.BaseURL` | Server origin for codegen |
| `API.Token` | Bare credential (or `--token` override already applied) |
| `API.Timeout` | Duration string, e.g. `30s` |

---

## Bearer Credential

Производное значение заголовка, не хранится в конфиге.

| Rule | Detail |
|------|--------|
| Input bare | `Authorization: Bearer <token>` |
| Input already `Bearer …` | Use as-is (no second `Bearer`) |
| Never log | Full value must not appear in user-facing error strings |

---

## Mapped Success Result

Для probe/happy path: результат `ProjectControllerListWithResponse` при 2xx.

| Aspect | Rule |
|--------|------|
| Status | 2xx → success path |
| Body | Prefer codegen `JSON200` when status 200 and decode succeeded; decode failure on 2xx → mapping/decode error (not silent zero value) |

Полные entity models остаются в `apiclient` (не дублировать DTO в `internal/api`).

---

## HTTPError (HTTP Failure Signal)

Типизированная ошибка неуспешного HTTP-ответа (контракт для F05).

| Field | Required | Notes |
|-------|----------|-------|
| StatusCode | yes | HTTP status from response |
| Body | optional | Raw response body bytes (may be empty) |
| Message | optional | Short text for `Error()` without embedding secrets |

**Behavior**:
- Returned by mapping helpers when status is not 2xx
- Inspectable via `errors.As` / `StatusCode` accessor
- No retry; no per-code taxonomy in F04

---

## Transport / Config Errors

Отдельные от `HTTPError` (нет HTTP status):

| Kind | When |
|------|------|
| Config/factory error | Empty token/URL, bad timeout |
| Transport error | connection refused, timeout from `http.Client`, context cancel |

---

## Probe Operation

| Attribute | Value |
|-----------|-------|
| operationId | `ProjectController_list` |
| Method/Path | `GET /v2/project` |
| Codegen | `ProjectControllerListWithResponse` |
| Used by | Unit happy path; `Session.ValidateConnectivity`; `singctl config validate` |

---

## State Transitions (Session)

```text
[Settings] --NewFromSettings--> [Session] --ValidateConnectivity--> [OK | HTTPError | transport error]
                 | fail-fast
                 v
            [error, no Session]
```

Нет долгоживущего persisted state у адаптера.
