# keeps-keeper

Минималистичные self-hosted общие заметки в Markdown с многопользовательским доступом.

## Стек

- `server/` — Go, gorilla/mux, pgx, JWT в httpOnly-cookie.
- `web/` — React + Vite + TypeScript, react-markdown.
- PostgreSQL 17. Схема применяется при старте сервера (`server/internal/postgres/schema.sql`).

## Запуск

```bash
cp .env.example .env   # заполнить секреты
docker compose up -d --build
docker compose exec app /keeper user add <login>
```

## Пользователи

```bash
docker compose exec app /keeper user add <login>     # создать
docker compose exec app /keeper user passwd <login>  # сменить пароль
docker compose exec app /keeper user del <login>     # удалить (сессии сразу недействительны)
```

Без TTY пароль читается из stdin: `echo pass | docker compose exec -T app /keeper user add bob`.

## Переменные

| Переменная | Описание | По умолчанию | Обязательная |
|---|---|---|---|
| `POSTGRES_DSN` | строка подключения к Postgres | — | да |
| `APP_JWT_SECRET` | секрет подписи токенов | — | да |
| `APP_JWT_TTL_HOURS` | время жизни сессии, ч | `720` | нет |
| `APP_COOKIE_SECURE` | флаг `Secure` у cookie | `true` | нет |
| `APP_SERVER_PORT` | порт HTTP-сервера | `80` | нет |
| `APP_SERVER_SHUTDOWN_TIMEOUT` | таймаут остановки, с | `10` | нет |
| `APP_STATIC_DIR` | каталог собранного клиента | `./static` | нет |

В `docker-compose.yml` используются `POSTGRES_PASSWORD`, `APP_JWT_SECRET`, `APP_PORT`, `APP_COOKIE_SECURE` из `.env`.

## API

Все `/api/*`, кроме входа и выхода, требуют cookie сессии. Ошибки: `{"error": "..."}`.

| Метод и путь | Тело | Ответ |
|---|---|---|
| `POST /api/login` | `{"login","password"}` | `{"login"}` + cookie |
| `POST /api/logout` | — | 204 |
| `GET /api/me` | — | `{"login"}` |
| `GET /api/folders` | — | `[{"id","name"}]` |
| `POST /api/folders`, `PUT /api/folders/{id}` | `{"name"}` | папка |
| `DELETE /api/folders/{id}` | — | 204, заметки остаются без папки |
| `GET /api/notes` | — | список без `body` |
| `GET /api/notes/{id}` | — | `{"id","title","body","folder_id","updated_by","updated_at"}` |
| `POST /api/notes`, `PUT /api/notes/{id}` | `{"title","body","folder_id"}` | заметка |
| `DELETE /api/notes/{id}` | — | 204 |

## Разработка

```bash
cd server && go install github.com/matryer/moq@v0.5.3
go generate ./... && go vet ./... && go test ./...

docker compose up -d db   # опубликуйте порт db или задайте свой POSTGRES_DSN
POSTGRES_DSN=... APP_JWT_SECRET=dev APP_COOKIE_SECURE=false APP_SERVER_PORT=8090 go run ./cmd/keeper

cd web && npm install && npm run dev   # /api проксируется на 127.0.0.1:8090
```

Моки (`*_moq_test.go`) не коммитятся, генерируются `go generate`.
