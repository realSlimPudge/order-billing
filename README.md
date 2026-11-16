# order-billing

Учебный проект: **двухсервисная система на Go**, построенная для изучения **gRPC**, **Protocol Buffers (proto)** и базовых распределённых паттернов. Не продакшен, а тренировочный проект.

## Зачем этот проект

Изучить на практике:

- как описывать API на **proto3** и генерировать Go-код;
- как поднять **gRPC-сервер** и подключиться к нему **gRPC-клиентом**;
- как спрятать сгенерированный gRPC-код за доменными интерфейсами с обеих сторон;
- как переводить доменные ошибки в **gRPC status codes** и обратно;
- как строить **идемпотентные** операции в распределённой системе.

## Архитектура

Два независимых сервиса, у каждого **своя база данных** (своя схема миграций), общаются только по gRPC.

```
        HTTP  :8080            gRPC  :50051
Client ──────► [order]  ─────────────► [billing]  ──► Postgres
                │  orders DB              │  accounts + charges DB
```

| Сервис | Транспорт | Порт | База данных | Роль |
|---|---|---|---|---|
| **order** | HTTP | `:8080` | `orderbilling_orders` | принимает заказы, оркестрирует оплату |
| **billing** | gRPC | `:50051` | `orderbilling_billing` | владеет деньгами: балансы и журнал списаний |

Каждый сервис — слои `model → repository → service → handler/server` (+ адаптер к внешнему миру).
Order зависит от billing **только через интерфейс** (`BillingClient`), billing реализуютт gRPC-адаптер `server.Server` — сгенерированный код не протекает в домен.

## Ключевые учебные темы

### proto3 и gRPC
Контракт в `proto/billing/billing.proto`:

```proto
service BillingService {
    rpc Ping(PingRequest) returns (PingResponse);
    rpc Charge(ChargeRequest) returns (ChargeResponse);
}
```

Из него генерируется Go-код (`billing_grpc.pb.go`, `billing.pb.go`).

### Маппинг ошибок — зеркально
Один и тот же «язык» gRPC-кодов используется на обоих краях:

- **billing (сервер):** доменная ошибка → `status.Error(codes.X)`
  - `ErrUserNotFound` → `codes.NotFound`
  - `ErrInsufficientFunds` → `codes.FailedPrecondition`
  - `ErrInvalidAmount`/`ErrInvalidID` → `codes.InvalidArgument`
- **order (клиент):** `status.Code(err)` → доменная ошибка
  - `codes.NotFound` → `model.ErrUserNotFound`
  - `codes.FailedPrecondition` → `model.ErrDeclined`
  - всё остальное (timeout/Unavailable/Internal) → `model.ErrBillingUnavailable`

Адаптеры `billing/server` и `order/billingclient` — единственные места, где живёт знание о gRPC.

### Идемпотентность (самая ценная часть)
Двойная защита от повторного списания:

- **order:** `INSERT ... ON CONFLICT (id) DO NOTHING` — повторный `POST` с тем же `id` не создаёт второй заказ;
- **billing:** `charges.order_id` — первичный ключ + «реплей» из журнала. Повторный `Charge` с тем же `order_id` возвращает сохранённый результат вместо повторного списания.

Charge выполняется в **транзакции** с `SELECT ... FOR UPDATE` (нужна блокировка строки при конкурентных списаниях).

### Статусы заказа и «исход неизвестен»
```
pending = заказ создан, но исход оплаты неизвестен (неопределённый сбой)
paid    = оплата прошла
failed  = оплата отклонена по известной причине (нет средств/нет юзера)
```

- `failed` — известный отказ → причина пишется в `fail_reason`.
- `pending` — **неопределённый** сбой (billing упал/таймаут). Повторный `POST` с тем же `id` «доигрывает» заказ: снова зовёт `Charge` и обновляет статус (безопасно благодаря идемпотентности billing).

Ограничение выбранного подхода: заказ может навсегда остаться `pending`, если клиент не повторит запрос. В продакшене это закрывается фоновым reconcile-воркером — здесь намеренно опущено (повторение паттерна из `task-scheduler` не было целью).

## HTTP API (order)

```
POST /orders           — создать заказ (идемпотентно по id)
GET  /orders/{id}      — получить заказ
```

**Запрос:**
```json
{ "id": "ord-1", "user_id": "user1", "amount": 500 }
```
`id` — клиентский ключ идемпотентности.

**Ответ** (заказ с финальным статусом):
```json
{ "id": "ord-1", "user_id": "user1", "amount": 500,
  "status": "paid", "created_at": "..." }
```
У `failed`-заказа добавляется `fail_reason`; у `paid`/`pending` его нет.

**Коды:** `201` создан, `200` повторно получен существующий, `400` невалидный запрос, `404` заказ не найден, `500` системная ошибка. Факт оплаты/отказа отражается в `status` тела, а не в HTTP-коде.

## Миграции

Нумеруются **в рамках каждой базы** (`schema_migrations` хранится в самой базе):

```
migrations/
  billing/  000001_billing   — accounts, charges
  orders/   000001_orders    — orders
            000002_orders    — + fail_reason
```

## Запуск

Требуется один Postgres (две базы: `orderbilling_orders` и `orderbilling_billing`) и `migrate` (golang-migrate).

```bash
# миграции для каждой базы
migrate -path migrations/billing -database "postgres://user:pass@localhost:5432/orderbilling_billing?sslmode=disable" up
migrate -path migrations/orders  -database "postgres://user:pass@localhost:5432/orderbilling_orders?sslmode=disable" up

# оплатить счёт для тестов (баланс user1)
psql -d orderbilling_billing -c "INSERT INTO accounts(user_id, balance) VALUES ('user1', 1000) ON CONFLICT (user_id) DO UPDATE SET balance = accounts.balance + EXCLUDED.balance;"

# сервисы
go run ./cmd/billing
go run ./cmd/order

# проверить
curl -X POST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"id":"ord-1","user_id":"user1","amount":500}'
```

Конфигурация — через `.env` (переменные: `BILLING_PORT`, `BILLING_HOST`, `ORDER_PORT`, `BILLING_DB_NAME`, `ORDER_DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_SSLMODE`).

## Стек

- Go, стандартный `net/http` + `pgx/v5`
- `google.golang.org/grpc`, `google.golang.org/protobuf`
- Postgres, golang-migrate
