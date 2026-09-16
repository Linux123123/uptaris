# Development

Run commands from repository root.

## Requirements

- Bun 1.3+
- Go 1.25+
- Running PostgreSQL database

## Start application

1. Create local configuration:

   ```bash
   cp .env.example .env
   ```

2. Create PostgreSQL database `uptaris`, or update `DATABASE_URL` in `.env` with your database connection.

3. Start API in first terminal:

   ```bash
   cd backend
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
cd frontend && bun run check && bun run lint
cd backend && go test ./...
```
