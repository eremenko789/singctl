# Contract: API Session Adapter (F04)

**Feature**: `004-api-adapter-auth` | **Date**: 2026-07-16
**Package**: `internal/api` (поверх `internal/apiclient`)

Контракт для разработчиков CLI/TUI и автотестов. Имена символов — ориентир для implement; точные идентификаторы могут слегка отличаться при сохранении семантики.

---

## Factory

```text
NewFromSettings(settings) -> (Session, error)
NewSession(baseURL, token, timeout) -> (Session, error)
```

**Inputs**:
- `baseURL`: non-empty origin without `/v2` (e.g. `https://api.singularity-app.ru` or `http://127.0.0.1:PORT` in tests)
- `token`: bare API token (or already-`Bearer `-prefixed; must not double-prefix)
- `timeout`: Go duration string (`30s`, `1m`, …)

**Fail-fast errors** (Session not created, no network):
- empty/whitespace token
- empty/whitespace base URL
- invalid timeout

**On success**: Session embeds configured `*apiclient.ClientWithResponses` such that every request includes `Authorization` and respects timeout.

---

## Authorization header

| Token in settings | Header value |
|-------------------|--------------|
| `abc` | `Bearer abc` |
| `Bearer abc` | `Bearer abc` |
| `BearerBearer abc` (no space after first Bearer) | treat as bare → `Bearer BearerBearer abc` (edge; tests focus on normal bare + `Bearer ` prefix) |

Тесты MUST проверять happy path с `test-token-…` и заголовком `Authorization: Bearer test-token-…`.

---

## Response mapping

```text
EnsureSuccess(statusCode, body) -> error
```

| statusCode | Result |
|------------|--------|
| 200–299 | `nil` |
| other | `*HTTPError` with programmatic StatusCode; Body may be set |

Consumers of `*XXXResponse` from codegen SHOULD:
1. handle transport `error` from the call;
2. call EnsureSuccess / equivalent on `StatusCode()`;
3. then read `JSON200` (etc.) on success.

**Retry**: none (exactly one HTTP attempt per call at adapter layer).

---

## Probe / ValidateConnectivity

```text
Session.ValidateConnectivity(ctx) -> error
```

- Performs `ProjectControllerListWithResponse` (GET `/v2/project`) via session client
- Returns `nil` on 2xx after mapping
- Returns `*HTTPError` on non-2xx
- Returns transport/context error otherwise
- MUST NOT create/update/delete entities

---

## Testing contract

- Unit tests use `net/http/httptest`
- Pass mock server URL as `baseURL`
- Assert Bearer header on at least one successful request
- Assert non-2xx → `*HTTPError` and single request to mock
- No real SingularityApp required for DoD

---

## Non-goals

- Entity facade methods (`ListTasks`, …)
- Retry / backoff
- Status taxonomy (401 vs 404 messages) — F05
- Debug HTTP logging / token redact — F29
