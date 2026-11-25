# Flashy — Backend

Серверна частина сервісу **Flashy** для вивчення карток за методом інтервальних повторень (SRS).
Це **лабораторна робота №4** з дисципліни «Архітектура програмного забезпечення»:
базова структура коду, що відповідає архітектурі, спроєктованій у попередній роботі.

## Команда

| Учасник | Група | Роль | Зона відповідальності |
| Буханцов Ілля | ІМ-42 | Back-End | Back-End |


## Архітектура

Монолітний Go-сервер, що працює в одному процесі. Web (React JS) і Mobile (React Native)
взаємодіють з ним через єдине REST API по HTTPS (JSON). Код розбито на три шари
відповідно до діаграми класів (Рис. 3.1):

```
HTTP ──► handler (API Handler) ──► service (бізнес-логіка) ──► repository (доступ до даних)
          AuthMiddleware             StudyService, DeckService,     інтерфейси + in-memory
          *Handler                   AuthService, StatsService,     реалізація (PostgreSQL/
                                     NotificationService,           Redis — наступний етап)
                                     SRSScheduler                   adapter: push, security
```

Докладніше — у [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), опис ендпоінтів — у [docs/API.md](docs/API.md).

## Структура проєкту

```
cmd/server/            точка входу (main.go), graceful shutdown
internal/
  domain/              сутності та правила предметної області (User, Deck, Card,
                       CardProgress, StudySession, UserStats, помилки)
  repository/          інтерфейси репозиторіїв (DIP)
    memory/            in-memory реалізація для розробки та тестів
  service/             бізнес-логіка: Auth, Deck, Study, Stats, Notification, SRSScheduler
  handler/             HTTP-хендлери, AuthMiddleware, роутер, формат помилок
  adapter/
    push/              адаптер push-сповіщень (заглушка FCM)
    security/          хешування паролів
  scheduler/           фоновий планувальник (горутина в тому ж процесі)
  config/              конфігурація зі змінних середовища
  app/                 composition root — складання залежностей + e2e-тест
  pkg/idgen/           генерація ID та токенів
docs/                  документація
```

## Реалізовано на цьому етапі

- Реєстрація, вхід, сесії (Bearer-токен або cookie `flashy_session`), ролі Guest/User/Administrator.
- Колоди й картки, публікація в каталозі, перегляд каталогу.
- SRS-тренування: черга карток, сесія, оцінка (Again/Hard/Good/Easy → 5 хв / 15 хв / 1 год / 4 дні),
  **ідемпотентність** за `clientReviewId`, відновлення списку активних сесій.
- Статистика й серія занять (streak).
- Реєстрація пристроїв і розсилка push-нагадувань (з видаленням недійсних токенів).
- Єдиний формат помилок `{ "code", "message" }` та стандартні HTTP-коди (401/403/404/409/422).

Заплановано далі: PostgreSQL та Redis, вхід через Google OAuth 2.0 і QR-код,
імпорт/експорт колод, модерація, звернення до підтримки, email-нагадування.

## Принципи SOLID у коді

| Принцип | Де застосовано |
|---|---|
| **S** | Кожен сервіс відповідає за одну сутність: `AuthService` — сесії, `DeckService` — колоди, `StatsService` — streak, `SRSScheduler` — лише інтервали. Хендлери лише валідують і делегують. |
| **O** | Нові режими тренування, канали сповіщень чи алгоритми SRS додаються новими реалізаціями інтерфейсів без зміни наявного коду. |
| **L** | Будь-яка реалізація `repository.*`, `PushSender`, `PasswordHasher`, `IntervalCalculator` взаємозамінна (in-memory ↔ PostgreSQL, заглушка ↔ FCM). |
| **I** | Вузькі інтерфейси: `handler.StatsUseCase` має один метод, окремі `UserRepository`, `SessionRepository` тощо замість одного «God-репозиторію». |
| **D** | Сервіси залежать від інтерфейсів репозиторіїв та `domain.IntervalCalculator`; конкретні типи підставляються лише в `internal/app`. |

## Запуск

Потрібен Go 1.22+. Зовнішніх залежностей немає.

```bash
git clone https://github.com/<org>/flashy-backend.git
cd flashy-backend

make test     # тести з покриттям
make run      # запуск на :8080
# або
go run ./cmd/server
```

Docker:

```bash
docker build -t flashy-backend .
docker run -p 8080:8080 flashy-backend
```

Змінні середовища (див. `.env.example`): `ADDR` (`:8080`), `SESSION_TTL_HOURS` (`720`),
`REMINDER_INTERVAL_MINUTES` (`60`).

### Швидка перевірка

```bash
curl -X POST localhost:8080/auth/register -d '{"name":"Ілля","email":"a@b.com","password":"secret123"}'
TOKEN=$(curl -s -X POST localhost:8080/auth/login -d '{"email":"a@b.com","password":"secret123"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')
curl -X POST localhost:8080/decks -H "Authorization: Bearer $TOKEN" -d '{"title":"English"}'
```

## Робота з Git

- Гілка `main` — стабільна; прямі коміти в неї заборонені.
- Кожне завдання — окрема гілка: `feature/auth-module`, `feature/deck-module`,
  `feature/study-module`, `feature/notifications`.
- Зміни потрапляють у `main` лише через Pull Request після код-рев'ю щонайменше одним учасником.
- Формат комітів: `feat: ...`, `fix: ...`, `test: ...`, `docs: ...`.
