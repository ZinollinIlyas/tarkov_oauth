# tarkov_oauth — сервис авторизации

Go + Gin + GORM (PostgreSQL) + bcrypt + JWT. Гексагональная структура:

```
cmd/main.go                      — точка входа
internal/config                  — конфиг из config.json (viper)
internal/server                  — инициализация DB, DI, роутер
internal/service                 — бизнес-логика (bcrypt, JWT)
internal/core/model              — модели (User -> auth.users)
internal/migrations              — SQL-миграции (embed), применяются при старте
internal/core/ports              — интерфейсы (UserRepository, AuthService)
internal/adapters/handler        — Gin-хендлеры
internal/adapters/repository     — GORM-репозиторий
```

## Запуск

Настройки лежат в `config.json` в корне (порт, DSN Postgres, секрет и TTL
JWT). Загружаются через viper; любое значение можно перекрыть env-переменной:
`http.port` → `HTTP_PORT`, `jwt.secret` → `JWT_SECRET`, `postgres.dsn` →
`POSTGRES_DSN`. Путь к файлу можно поменять через `CONFIG_PATH`.

```bash
go mod tidy
# отредактировать config.json (jwt.secret, postgres.dsn)
go run ./cmd
```

## Миграции

При старте сервис сам применяет миграции из `internal/migrations/sql/*.sql`
(по порядку имён). Применённые версии хранятся в `public.schema_migrations`,
каждая миграция выполняется один раз в транзакции. Новая миграция — новый файл
со следующим номером, например `0002_add_something.sql`. Уже применённые файлы
не редактировать.

## API

### POST /api/v1/auth/sign-up

```json
{ "nickname": "rat", "display_name": "Тарковская Крыса", "password": "secret123" }
```

→ `201 Created` с созданным пользователем. `409` если nickname занят.

### POST /api/v1/auth/sign-in

```json
{ "nickname": "rat", "password": "secret123" }
```

→ `200 OK`:

```json
{ "access_token": "eyJ...", "token_type": "Bearer", "expires_at": 1770000000 }
```

`401` при неверных данных, `403` если пользователь заблокирован.

### GET /api/v1/users/me

Заголовок `Authorization: Bearer <access_token>` → `200 OK`:

```json
{ "id": "…uuid…", "nickname": "rat", "display_name": "Тарковская Крыса" }
```

`401` — токен невалиден/просрочен, `403` — пользователь заблокирован.
Этим эндпоинтом игровой бэкенд (tarkov_back) проверяет токен клиента.

### POST /api/v1/users/lookup

С тем же заголовком. Тело `{ "ids": ["uuid", …] }` (до 200 id) → массив
объектов как у `/users/me`. Неизвестные id в ответе отсутствуют.

### GET /health

Проверка живости.
