# REST API (чернетка контракту)

JSON, UTF-8. Помилки: `{ "code": "...", "message": "..." }`.
Автентифікація: `Authorization: Bearer <token>` (Mobile) або cookie `flashy_session` (Web).

| Метод | Шлях | Auth | Опис | Успіх |
|---|---|---|---|---|
| GET | `/health` | — | Перевірка стану | 200 |
| POST | `/auth/register` | — | `{name, surname, email, password}` | 201 |
| POST | `/auth/login` | — | `{email, password}` → `{token, user}` + cookie | 200 |
| POST | `/auth/logout` | ✓ | Завершення сесії | 204 |
| GET | `/users/me` | ✓ | Поточний користувач | 200 |
| POST | `/decks` | ✓ | `{title, description}` | 201 |
| GET | `/decks` | ✓ | Мої колоди | 200 |
| GET | `/decks/{id}` | ✓ | `{deck, cards}` | 200 |
| POST | `/decks/{id}/cards` | ✓ | `{term, translation, example}` (власник) | 201 |
| POST | `/decks/{id}/publish` | ✓ | Опублікувати в каталозі (власник, ≥1 картка) | 200 |
| GET | `/catalog/decks` | ✓ | Публічні колоди | 200 |
| GET | `/decks/{id}/study/queue?limit=` | ✓ | Черга карток | 200 |
| POST | `/study/sessions` | ✓ | `{deckId, mode}`, mode: FLASHCARD/TEST/WRITING | 201 |
| GET | `/study/sessions?status=active` | ✓ | Незавершені сесії | 200 |
| POST | `/study/sessions/{id}/reviews` | ✓ | `{cardId, rating, clientReviewId}` → `{newState, intervalSec, nextDueAt}` | 200 |
| POST | `/study/sessions/{id}/finish` | ✓ | Завершити сесію, оновити streak | 200 |
| GET | `/stats/me` | ✓ | `{currentStreak, longestStreak, lastStudyDate}` | 200 |
| POST | `/devices` | ✓ | `{fcmToken, platform}`, platform: IOS/ANDROID/WEB | 201 |

Коди помилок: `401` не автентифіковано, `403` недостатньо прав, `404` не знайдено,
`409` конфлікт (дублікат email, сесію завершено), `422` помилка валідації.
