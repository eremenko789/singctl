# Data Model: Error Handling & Retry (F05)

**Feature**: `005-error-retry` | **Date**: 2026-07-16

Логические сущности поверх F04 (`HTTPError`, `Session`). Не схема БД.

---

## ClassifiedError

Пользовательски значимая ошибка после taxonomy.

| Field | Type | Rules |
|-------|------|-------|
| Kind | enum | Unauthorized, Forbidden, NotFound, Validation, RateLimited, Server, Other, Config, Transport, Date |
| Message | string | Стабильный user-facing текст; без токена/Authorization |
| Cause | error | Опционально: `*HTTPError`, transport, factory, DateError |
| EntityID | string | Опционально; влияет на Message при NotFound |
| StatusCode | int | Опционально; из HTTPError если есть |

**Relationships**: ClassifiedError *wraps* HTTPError / client errors; CLI ExitCode *reads* Kind.

---

## HTTPError (F04, unchanged core)

| Field | Notes |
|-------|-------|
| StatusCode | Input to Classify |
| Body | Used for 422 extraction; not dumped raw on other codes |

EnsureSuccess по-прежнему возвращает `*HTTPError` на non-2xx; Classify вызывается на границе UX/CLI или в validate helper.

---

## RetryPolicy

| Field | Value |
|-------|-------|
| MaxAttempts | 3 |
| BackoffWithoutRetryAfter | [1s, 2s] before attempts 2 and 3 |
| RetryAfterCap | 30s |
| RetryStatuses | only 429 |
| Sleeper | injectable for tests |

**State** (one logical call):

```text
attempt=1 --429--> sleep --attempt=2 --429--> sleep --attempt=3 --429--> Classified RateLimited
                 \-2xx--> success
                 \-other status--> Classified (no more attempts)
                 \-transport--> Transport error (no retry)
```

---

## DateError / ParseDate

| Input | Result |
|-------|--------|
| `YYYY-MM-DD` valid calendar date | `time.Time` UTC/local per `time.Parse` (document: date-only, location Local or UTC — implement picks UTC midnight; tests compare year/month/day) |
| empty / wrong layout / impossible day | `*DateError` with Message containing `Expected: YYYY-MM-DD` |

Kind for ExitCode: treat as Validation → exit **1** (client input error, not config). Spec exit table: config=2, not found=3, else API=1; date is client validation → **1**.

---

## Config / missing token

| Signal | Kind | Exit |
|--------|------|------|
| Empty token / factory fail-fast / validate pre-check | Config | 2 |

Message MAY remain Russian F02/F04 hint with `config set-token`.

---

## Transport failure

| Signal | Kind | Exit | Message (suggested) |
|--------|------|------|---------------------|
| connection refused, timeout, DNS, etc. (no HTTP status) | Transport | 1 | `Error: could not reach API` (or RU equivalent without secrets) |

---

## Exit code mapping

| Kind | ExitCode |
|------|----------|
| (nil error) | 0 |
| NotFound | 3 |
| Config | 2 |
| Unauthorized, Forbidden, Validation, Date, RateLimited, Server, Other, Transport | 1 |

---

## Message catalog (HTTP)

Canonical strings for tests (exact or prefix-stable):

| Code | Message |
|------|---------|
| 401 | `Error: invalid token. Run 'singctl config set-token'` |
| 403 | `Error: insufficient token permissions` |
| 404 | `Error: entity not found` / `Error: entity not found: <ID>` |
| 422 | extracted / `Error: validation failed` |
| 429 exhausted | `Error: rate limited. Retry later` |
| 5xx | `Error: server error, retry later` |
