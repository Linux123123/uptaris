# Production build and deployment

Run commands from repository root.

## Configure production environment

Create `.env` with production database and public URLs:

```dotenv
APP_ENV=production
API_PORT=8080
DATABASE_URL=postgres://USER:PASSWORD@HOST:5432/uptaris?sslmode=require
CORS_ORIGINS=https://app.example.com
VITE_API_URL=https://api.example.com/api/v1
SENTRY_DSN=
```

Replace example values before deployment.

## Build

```bash
cd frontend
bun install --frozen-lockfile
bun run build

cd ../backend
mkdir -p ../bin
go build -o ../bin/uptaris-api ./cmd/api
```

## Deploy

1. Serve `frontend/dist` from static host or web server.
2. Start API from repository root so it reads root `.env`:

   ```bash
   ./bin/uptaris-api
   ```

3. Confirm API health:

   ```bash
   curl --fail http://localhost:8080/healthz
   ```

Place TLS and reverse proxy in front of static frontend and API.
