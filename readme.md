# go-bid

A real-time auction backend written in Go. Users sign up, create product auctions, and connect
over WebSocket to place bids that are broadcast live to everyone watching the same auction.

## Features

- Session-cookie authentication backed by PostgreSQL (`scs` + `pgxstore`)
- Create product auctions with a base price and an auction end time
- Real-time bidding over WebSocket, with live broadcast of new bids to all subscribers
- Per-auction rooms that automatically close and notify clients when the auction ends
- Request validation with structured field-level error responses

## Tech stack

- **Go 1.26**
- **chi** — HTTP routing and middleware
- **pgx** — PostgreSQL driver and connection pool
- **sqlc** — type-safe Go generated from SQL
- **tern** — database migrations
- **scs** — session management (Postgres-backed)
- **gorilla/websocket** — real-time connections
- **gorilla/csrf** — CSRF protection (wired up, disabled in development)
- **air** — live reload during development

## Prerequisites

- Go 1.26+
- Docker (for PostgreSQL), or a local PostgreSQL 17 instance
- These tools on your `PATH`:
  - [`tern`](https://github.com/jackc/tern) — migrations
  - [`sqlc`](https://sqlc.dev/) — code generation
  - [`air`](https://github.com/air-verse/air) — live reload (optional)

```sh
go install github.com/jackc/tern/v2@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/air-verse/air@latest
```

## Getting started

1. **Configure the environment.** Create a `.env` file in the project root:

   ```sh
   GOBID_DATABASE_PORT=5432
   GOBID_DATABASE_USER=postgres
   GOBID_DATABASE_PASSWORD=postgres
   GOBID_DATABASE_NAME=go_bid
   GOBID_DATABASE_HOST=localhost
   GOBID_CSRF_KEY=a-32-byte-long-secret-key-here!!
   ```

   All commands read these values from `.env`.

2. **Start PostgreSQL:**

   ```sh
   docker compose up -d
   ```

3. **Run the migrations:**

   ```sh
   go run ./cmd/terndotenv
   ```

4. **Run the API** (with live reload):

   ```sh
   make run-api
   ```

   Or without reload:

   ```sh
   go run ./cmd/api
   ```

   The server listens on `http://localhost:3080`.

## API

All routes are prefixed with `/api/v1`.

| Method | Path                                  | Auth | Description                          |
| ------ | ------------------------------------- | ---- | ------------------------------------ |
| POST   | `/users/signup`                       | No   | Register a new user                  |
| POST   | `/users/login`                        | No   | Log in and start a session           |
| POST   | `/users/`                             | Yes  | Log out the current user             |
| POST   | `/products/`                          | Yes  | Create a product auction             |
| GET    | `/products/ws/subscribe/{product_id}` | Yes  | Join an auction's WebSocket and bid  |

### WebSocket messages

Clients exchange JSON messages with a `kind` field. Incoming: `PlaceBid`. Outgoing:
`SuccessfullyPlacedBid`, `FailedToPlaceBid`, `NewBidPlaced`, `InvalidJSON`, `AuctionFinished`.

## Project layout

```
cmd/
  api/          Application entry point (composition root)
  terndotenv/   Loads .env, then runs tern migrations
internal/
  api/          HTTP handlers, routing, middleware
  services/     Business logic (users, products, bids, auction rooms)
  usecase/      Request DTOs and validation rules
  store/pgstore Generated DB layer (sqlc) — queries, migrations, models
  validator/    Validation helpers
  jsonutils/    JSON encode/decode + validation helpers
```

## Development

```sh
# Apply migrations
go run ./cmd/terndotenv

# Regenerate the DB layer after editing queries/ or migrations/
sqlc generate -f ./internal/store/pgstore/sqlc.yml

# Run tests
go test ./...
```

> **Note:** Files under `internal/store/pgstore` (`*.sql.go`, `models.go`, `db.go`) are generated
> by sqlc. Edit the SQL in `queries/` and the schema in `migrations/`, then rerun `sqlc generate`.
