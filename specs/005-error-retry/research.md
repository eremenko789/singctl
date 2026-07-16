# Research: Error Handling & Retry (F05)

**Feature**: `005-error-retry` | **Date**: 2026-07-16

Все пункты Technical Context и отложенные из clarify константы закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где выполнять retry

**Decision**: Custom `http.RoundTripper`, оборачивающий `http.DefaultTransport` (или transport существующего client), устанавливается в `http.Client` сеанса при фабрике F04. Codegen-вызовы получают 429-retry прозрачно.

**Rationale**: Constitution III/IV — один слой для всех операций без обёртки каждой codegen-метода; F04 уже передаёт `WithHTTPClient`.

**Alternatives considered**:
- Retry только вокруг `ValidateConnectivity` — не масштабируется на будущие команды.
- Внешняя библиотека (`hashicorp/go-retryablehttp` и т.п.) — лишняя зависимость при узкой политике «только 429, 3 попытки».
- Retry в CLI — нарушает G4.

---

## 2. Константы backoff и потолка

**Decision**:

| Constant | Value | Role |
|----------|-------|------|
| Max HTTP attempts | `3` | первая + до двух повторов при 429 |
| Exponential delays (no/`Retry-After`) | before 2nd: `1s`; before 3rd: `2s` | база 1s, удвоение |
| `Retry-After` cap | `30s` | одна пауза не дольше 30s |
| Test sleeper | injectable `func(context.Context, time.Duration) error` | нулевые/мгновенные паузы в unit-тестах |

**Rationale**: Согласовано со spec assumptions (1s/2s) и clarify B (`Retry-After` + потолок). 30s достаточно уважать сервер, но не «замораживать» CLI на минуты при ошибочном заголовке.

**Alternatives considered**:
- Cap 60s/5m — слишком долго для интерактивного CLI.
- Jitter — полезен при thrashing; не требуется ТЗ; отложить.
- Игнорировать `Retry-After` — отклонено clarify.

---

## 3. Парсинг `Retry-After`

**Decision**: Поддержать RFC 7231: integer delay-seconds **или** HTTP-date. Невалидное/отрицательное/нулевое в смысле «уже прошло» для date → fallback на exponential для этой паузы. Итог `min(parsed, 30s)`; если parsed > cap — ждать cap (не отбрасывать заголовок полностью).

**Rationale**: Clarify B; предсказуемые тесты с `Retry-After: 1` и с завышенным значением.

**Alternatives considered**:
- Только seconds — проще, но неполно относительно HTTP.
- Strict follow without cap — отклонено clarify.

---

## 4. Taxonomy / Classify

**Decision**: Функция `Classify(err error, opts ...)` (или `ClassifyHTTP` + обёртка):

1. Если `errors.As` → `*HTTPError` — map status → `ClassifiedError{Kind, Message, Cause}`.
2. Иначе config/factory / typed missing-token → Kind Config.
3. Иначе DateError → Kind Validation (client).
4. Иначе transport/context/other → Kind Transport (или generic API failure) с сообщением без секретов.

Таблица сообщений (стабильные строки):

| Status | Kind | Message |
|--------|------|---------|
| 401 | Unauthorized | `Error: invalid token. Run 'singctl config set-token'` |
| 403 | Forbidden | `Error: insufficient token permissions` |
| 404 | NotFound | `Error: entity not found: <ID>` или `Error: entity not found` |
| 422 | Validation | body message if extractable; else `Error: validation failed` |
| 429 (after exhaust) | RateLimited | `Error: rate limited. Retry later` |
| 5xx | Server | `Error: server error, retry later` |
| other 4xx | Other | safe fallback `Error: request failed (HTTP NNN)` without body dump of secrets |

`Error()` у `ClassifiedError` возвращает Message (для stderr). `HTTPError.Error()` может остаться низкоуровневым; CLI печатает Classified.

**Rationale**: Spec US1 + clarify Q2; `errors.As` контракт F04.

**Alternatives considered**:
- Менять только `HTTPError.Error()` без нового типа — слабее для Kind/exit mapping.
- Парсить только строки — запрещено F04 clarify.

---

## 5. Извлечение тела 422

**Decision**: Если body UTF-8 text: попробовать JSON с полями `message` / `error` / `detail` (string); иначе trim plain text если printable и длина разумна (например ≤ 2KiB); иначе fallback `Error: validation failed`. Не печатать сырой бинарь.

**Rationale**: FR-003; устойчивость к разным формам API.

---

## 6. Entity ID для 404

**Decision**: Опция `ClassifyOption` / параметр `EntityID string`. Если Kind NotFound и ID non-empty → суффикс `: <ID>`. Probe `ValidateConnectivity` (list) обычно без ID → сообщение без суффикса.

**Rationale**: FR-004; list validate не обязан знать ID.

---

## 7. Client date validation

**Decision**: `ParseDate(s string) (time.Time, error)` в `internal/api` (или соседний файл пакета): strict `YYYY-MM-DD` (`time.Parse("2006-01-02", …)`), календарная валидность. Ошибка — типизированная `DateError` с текстом, содержащим `Expected: YYYY-MM-DD`. Только unit-тесты в DoD (clarify A).

**Rationale**: FR-008a; без CLI-флага.

**Alternatives considered**:
- Отдельный пакет `internal/dates` — лишний слой при одном хелпере.
- Отложить до F07/commands — отклонено backlog F05.

---

## 8. Missing token unification

**Decision**: Не ломать русские сообщения F02/F04 фабрики; `Classify` / exit helper распознают отсутствие токена (sentinel / `errors.Is` / известный текст factory) как Kind Config → exit 2. HTTP 401 остаётся EN-сообщением ТЗ.

**Rationale**: Scriptability по кодам важнее унификации языка; UX set-token уже есть.

---

## 9. Exit codes в process boundary

**Decision**: `cli.ExitCode(err error) int` по Kind:

| Kind | Code |
|------|------|
| nil | 0 |
| NotFound | 3 |
| Config (missing token, bad settings) | 2 |
| Unauthorized, Forbidden, Validation, RateLimited, Server, Other, Transport | 1 |

`cmd/singctl/main.go`: `os.Exit(cli.ExitCode(err))` вместо всегда `1`. Cobra `Execute` по-прежнему возвращает error; печать stderr — существующий путь `cli.Execute`.

**Rationale**: Clarify Q3/Q5; F07 может расширить pipe, но числовые коды для taxonomy нужны в F05 DoD.

**Alternatives considered**:
- Только library без main change — exit всегда 1, ломает SC/US4.
- Код 4 для transport — отклонено clarify.

---

## 10. Поведение RoundTripper при 429

**Decision**:
- На 429: прочитать/закрыть body текущей попытки (для заголовков достаточно; body финального ответа сохранить для Classify), sleep, повторить тот же `Request` (Clone с GetBody если нужно).
- После 3-й 429: вернуть этот `*http.Response` вызывающему (codegen → EnsureSuccess → HTTPError 429 → Classify RateLimited).
- На non-429: вернуть сразу (1 attempt).
- На transport error: вернуть сразу, без retry.
- Уважать `ctx.Done()` во время sleep → вернуть `ctx.Err()`.

**Rationale**: FR-005/006; прозрачность для `ClientWithResponses`.

---

## 11. Интеграция `config validate`

**Decision**: После probe error — `Classify(err)` (без EntityID); вернуть ClassifiedError (или fmt с Message). Не оборачивать так, чтобы потерять `errors.As` Kind. Тесты: мок 401/404/429×3/5xx → ожидаемые подстроки + ExitCode через executeForTest/`main` helper.

**Rationale**: FR-010a, SC-008.

---

## 12. Итог

Открытых NEEDS CLARIFICATION нет. Готово к Phase 1 design artifacts и `/speckit-tasks`.
