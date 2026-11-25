# Архітектура backend

Відповідає документу «Архітектура програмного забезпечення» (проєкт Flashy).

## Відповідність діаграмі класів

| Елемент діаграми | Реалізація |
|---|---|
| `AuthMiddleware` | `internal/handler/middleware.go` |
| `AuthHandler`, `DeckHandler`, `StudyHandler`, `StatsHandler`, `CatalogHandler`, `NotificationHandler` | `internal/handler/*.go` (`Catalog` — метод `DeckHandler`) |
| `AuthService`, `DeckService`, `StudyService`, `StatsService`, `NotificationService` | `internal/service/*.go` |
| `SRSScheduler` | `internal/service/srs.go` |
| `Scheduler` (фонова горутина) | `internal/scheduler/scheduler.go` |
| `User`, `Session`, `PushDevice` | `internal/domain/user.go` |
| `Deck`, `Card` | `internal/domain/deck.go` |
| `StudySession`, `CardProgress`, `ReviewLog`, `UserStats` | `internal/domain/study.go` |
| `*Repository` | інтерфейси — `internal/repository`, реалізація — `repository/memory` |
| `FCMPushSender` | інтерфейс `service.PushSender`, заглушка — `adapter/push` |

Відхилення від діаграми (свідомі):

- `Administrator` реалізовано як **роль** `User.Role` (у Go немає наслідування; використовується композиція).
- Репозиторії in-memory; PostgreSQL та Redis підключаються заміною реалізацій у `internal/app`.
- Ще не реалізовано: `ImportExportService`, `ModerationService`, `SupportService`,
  `UserService`, `GoogleOAuthProvider`, `EmailSender`, QR-вхід (заплановано на наступні етапи).

## Залежності між шарами

```
cmd/server → app → handler → (інтерфейси use case) ← service → repository (інтерфейси) ← memory
                                                     service → domain
```

Правило: залежності спрямовані **всередину** — `domain` не знає про жоден інший шар,
`service` не імпортує `handler`, `handler` не імпортує `service`
(залежить лише від власних інтерфейсів у `usecases.go`).

## Ключові рішення

- **Ідемпотентність оцінки картки.** `clientReviewId` + `ReviewLog.FindByClientID`: повторний запит
  після обриву зв'язку повертає попередній результат, не змінюючи прогрес.
- **Сесії.** У сховищі лежить лише SHA-256 хеш токена; клієнту видається відкритий токен
  (Bearer для Mobile, cookie `flashy_session` для Web).
- **Приховування чужих ресурсів.** Чужа непублічна колода чи сесія повертає `404`, а не `403`.
- **Час** абстраговано інтерфейсом `Clock`, тому логіку streak та SRS тестують детерміновано.
