# Development

Requirements: Go 1.26+, Bun 1.3.11, and a running PostgreSQL or CockroachDB database.
Run commands from repository root unless noted.

## Configuration

```bash
cp .env.example .env
```

Set `DATABASE_URL` for an existing database. Use `postgres://` for PostgreSQL or
`cockroachdb://` for CockroachDB. Other local settings can use `.env.example` defaults.

### Application and sessions

| Environment key        | Local value / purpose                                                                       |
| ---------------------- | ------------------------------------------------------------------------------------------- |
| `APP_ENV`              | `development`; also supports `production`                                                   |
| `API_HOST`             | `127.0.0.1`                                                                                 |
| `API_PORT`             | `8080`                                                                                      |
| `DATABASE_URL`         | `postgres://uptaris:uptaris@localhost:5432/uptaris?sslmode=disable`                         |
| `VITE_API_URL`         | `http://localhost:8080/api/v1`                                                              |
| `FRONTEND_URL`         | `http://localhost:5173`                                                                     |
| `WEBAUTHN_RP_ID`       | `localhost` locally; required in production and scoped to frontend domain                   |
| `CORS_ORIGINS`         | `http://localhost:5173`; comma-separated frontend origins                                   |
| `JWT_ACCESS_SECRET`    | `development-only-change-me` locally; independent secret of at least 32 bytes in production |
| `REFRESH_TOKEN_PEPPER` | `development-only-change-me` locally; independent secret of at least 32 bytes in production |
| `ACCESS_TOKEN_TTL`     | `15m`                                                                                       |
| `REFRESH_TOKEN_TTL`    | `720h`                                                                                      |
| `COOKIE_SECURE`        | `false` locally; `true` in production                                                       |
| `COOKIE_SAME_SITE`     | `lax`; `none` requires secure HTTPS cookies                                                 |
| `TRUSTED_PROXIES`      | Empty; optional comma-separated proxy IPs/CIDRs                                             |
| `SENTRY_DSN`           | Empty; optional error reporting                                                             |

Production frontend, API callback, and CORS URLs must use HTTPS.

### Optional OAuth and two-factor authentication

Leave `OAUTH_PROVIDERS` empty to disable OAuth. Set it to `github`, `google`, or
`github,google` and supply credentials for each enabled provider.

| Environment key              | Value / requirement                                                                                              |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `OAUTH_PROVIDERS`            | Empty by default; comma-separated provider IDs                                                                   |
| `OAUTH_GITHUB_CLIENT_ID`     | Required when GitHub is enabled                                                                                  |
| `OAUTH_GITHUB_CLIENT_SECRET` | Required when GitHub is enabled                                                                                  |
| `OAUTH_GITHUB_REDIRECT_URL`  | `http://localhost:8080/api/v1/auth/oauth/github/callback`                                                        |
| `OAUTH_GOOGLE_CLIENT_ID`     | Required when Google is enabled                                                                                  |
| `OAUTH_GOOGLE_CLIENT_SECRET` | Required when Google is enabled                                                                                  |
| `OAUTH_GOOGLE_REDIRECT_URL`  | `http://localhost:8080/api/v1/auth/oauth/google/callback`                                                        |
| `OAUTH_TOKEN_ENCRYPTION_KEY` | Required with OAuth; persistent 32-byte key encoded as 64 hex characters                                         |
| `TWO_FACTOR_ENCRYPTION_KEY`  | Required for authenticator enrollment/verification; separate persistent 32-byte key encoded as 64 hex characters |

Provider callback URLs must match their configured redirect URLs exactly. Keep encryption keys stable.

## Run

API, migrations, and optional demo data:

```bash
cd backend
go run ./cmd/migrate
go run ./cmd/seed
go run ./cmd/api
```

Frontend, in another terminal:

```bash
cd frontend
bun install
bun run dev
```

| Service    | URL                                        |
| ---------- | ------------------------------------------ |
| Frontend   | `http://localhost:5173`                    |
| API        | `http://localhost:8080/api/v1`             |
| Health     | `http://localhost:8080/healthz`            |
| Swagger UI | `http://localhost:8080/swagger/index.html` |

Demo accounts:

- `viewer@uptaris.local`
- `operator@uptaris.local`
- `admin@uptaris.local`

Password: `UptarisDemo!2026`.

Seeding skips databases that already contain users and refuses production mode.

## Build and check

From `frontend/`:

```bash
bun run format
bun run check
bun run lint
bun run build
```

From `backend/`:

```bash
gofmt -w .
go vet ./...
go build ./...
go generate ./cmd/api
```

The final command regenerates API documentation.
