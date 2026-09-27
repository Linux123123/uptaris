# Development

Run commands from repository root.

## Requirements

- Bun 1.3.11
- Go 1.26+
- Running PostgreSQL or CockroachDB database

## Start application

1. Create local configuration:

   ```bash
   cp .env.example .env
   ```

2. Create database `uptaris`, or update `DATABASE_URL` in `.env` with your database connection. Use `cockroachdb://` for CockroachDB so migrations select its driver; the API accepts the same URL and retries CockroachDB transactions. Keep the cluster's TLS parameters, such as `sslmode=verify-full` and `sslrootcert`, in the URL.

3. Start API in first terminal:

   ```bash
   cd backend
   go run ./cmd/migrate
   go run ./cmd/seed
   go run ./cmd/api
   ```

4. Start frontend in second terminal:

   ```bash
   cd frontend
   bun install
   bun run dev
   ```

Open `http://localhost:5173`.

## Local URLs

- Frontend: `http://localhost:5173`
- API: `http://localhost:8080`
- Health check: `http://localhost:8080/healthz`
- Swagger UI: `http://localhost:8080/swagger/index.html`

## Check changes

```bash
cd frontend
bun run build
bun run check
bun run lint
bun run format:check

cd ../backend
go vet ./...
go build ./...
```

## API documentation

Swagger UI runs at `http://localhost:8080/swagger/index.html`.
Regenerate the checked-in OpenAPI output after handler annotation changes:

```bash
cd backend
go generate ./cmd/api
```

## Full verification

See [testing guide](testing.md) for the Swagger API walkthrough and dependency checks. See the wiki for [API endpoints](https://github.com/Linux123123/uptaris/wiki/API-Reference) and the [requirements review](https://github.com/Linux123123/uptaris/wiki/Project-Report).

## Demo accounts

The seed command creates at least five servers, five monitors, and five incidents in an empty database. It refuses production mode and skips a database that already contains users.

| Email | Role | Access |
| --- | --- | --- |
| `viewer@uptaris.local` | Viewer | Read owned resources; starts without inventory |
| `operator@uptaris.local` | Operator | Manage seeded resources |
| `admin@uptaris.local` | Administrator | Manage all resources and user roles |

Demo password: `uptaris-demo-2026`. Public registration creates viewers. Use the administrator UI to grant operator access.

## Code organization

Custom React components each live in their own file. Small render callbacks for tables and forms stay with their callers. Generated shadcn primitives in `src/components/ui` retain their original structure. Shared field validation, options, and cache updates live in `src/lib`; API routes live in `backend/routers`, HTTP handling in `backend/internal/handlers`, and database operations in `backend/internal/inventory`, `auth`, and `users`.

Backend dependencies are constructed in `backend/cmd/api/main.go`. Route files register paths and middleware only. Request decoding and response formatting live in `backend/internal/request` and `response`; business operations accept `context.Context` and return errors. Keep Gin and HTTP responses out of the business packages. PATCH operations reload and lock the current row before applying changes.

Format frontend code with `bun run format` from `frontend`, and Go code with `gofmt` before running checks. Update both wiki languages when changing behavior. Monitoring stays manual; monitor intervals and expected health describe configuration, not scheduled background work.
