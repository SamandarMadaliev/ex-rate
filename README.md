# Ex-Rate

A small Go service that tracks currency exchange rates. A client requests a rate for a
currency pair; the API stores it as `pending` and returns immediately, while a background
worker fetches the price from an external rate provider and updates the row to
`completed` or `failed`.

## Stack

- Go, [chi](https://github.com/go-chi/chi) router
- PostgreSQL ([pgx](https://github.com/jackc/pgx) driver), SQL migrations via [golang-migrate](https://github.com/golang-migrate/migrate)
- An in-process worker pool for background price-fetch jobs
- Swagger/OpenAPI docs via [swaggo](https://github.com/swaggo/swag)

## Quick start

Prerequisites: Docker (with Compose).

```bash
git clone git@github.com:SamandarMadaliev/ex-rate.git
cd ex-rate
cp .env.example .env          # fill in EX_RATE_API_TOKEN / EX_RATE_API_URL
docker compose up -d --build  # starts Postgres, runs migrations, starts the API
curl http://localhost:9090/api/v1/health   # -> OK
```

To stop everything: `docker compose down` (add `-v` to also drop the Postgres volume).

### Running without Docker

```bash
make migrate-up                 # apply migrations against a Postgres you already have running
# edit .env: set DATABASE_HOST=localhost (and any other DB settings to match)
make run                        # go run ./cmd/main.go, reads .env via godotenv
```

## Endpoints

| Method | Path                                  | Description                                     |
| ------ | ------------------------------------- | ----------------------------------------------- |
| GET    | `/api/v1/health`                    | DB connectivity check                           |
| GET    | `/api/v1/currencies`                | List available currencies                       |
| POST   | `/api/v1/rates`                     | Create a rate for a currency pair (async fetch) |
| GET    | `/api/v1/rates/{id}`                | Get a rate by ID                                |
| GET    | `/api/v1/rates/latest?base=&quote=` | Latest rate for a currency pair                 |

Interactive docs: `http://localhost:9090/swagger/index.html`.

## Configuration

All config is read from environment variables (see `.env.example`) — server timeouts,
`DATABASE_*` connection settings, `WORKER_COUNT` / `WORKER_BUFFER_SIZE` / `WORKER_JOB_TIMEOUT`
for the background pool, and `EX_RATE_API_URL` / `EX_RATE_API_TOKEN` for the price provider.

## Development

```bash
make test      # go test ./...
make swagger   # regenerate docs/ after changing an endpoint or its schemas
```

### What I would add

- No retries or backoff on a failed price fetch. On failed status rates we can set the retry.
- No pagination on `GET /currencies` or a list endpoint for rates. But because of list of currencies are currently limited, it was not added for now.
- Also we could add auth. At least base auth. But it was not listed in the requirements.
- Putting validation on creating a rate. If the same request for the same pair of currency was made during the N period of time do not send a concurrent request for the external api but return the last result.


s
