# Build and deployment

Target frontend: `https://uptaris.linux123123.com` on Cloudflare Pages. Target API: `https://uptaris-api.linux123123.com` on a SparkedHost VPS with nginx and PostgreSQL.

## Frontend: Cloudflare Pages

Connect the repository to Pages and configure:

| Setting | Value |
| --- | --- |
| Root directory | `frontend` |
| Build command | `bun install --frozen-lockfile && bun run check && bun run build` |
| Output directory | `dist` |
| `BUN_VERSION` | `1.3.11` |
| `VITE_API_URL` | `https://uptaris-api.linux123123.com/api/v1` |
| Custom domain | `uptaris.linux123123.com` |

Set the public API URL before building; Vite embeds it into the output. Keep database credentials and signing secrets out of frontend environment variables. Pages provides SPA routing when no top-level `404.html` exists, so direct navigation to `/app/servers` serves the application. See [Pages build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration/), [build environment](https://developers.cloudflare.com/pages/configuration/build-image/), and [React routing behavior](https://developers.cloudflare.com/pages/framework-guides/deploy-a-react-site/).

For a local production build, from the repository root:

```bash
cd frontend
bun install --frozen-lockfile
VITE_API_URL=https://uptaris-api.linux123123.com/api/v1 bun run build
```

## API: build artifacts

Use Go 1.26 or newer. Build on the target architecture, or set `GOOS` and `GOARCH` explicitly for the VPS architecture.

```bash
cd backend
mkdir -p ../bin
go build -o ../bin/uptaris-api ./cmd/api
go build -o ../bin/uptaris-migrate ./cmd/migrate
```

The deployment bundle contains `bin/uptaris-api`, `bin/uptaris-migrate`, and `backend/migrations`. The migration binary resolves `migrations` relative to its working directory.

## Automated API deployment

GitHub Actions deploys the API to production automatically on pushes to `main`. Deployment host, user, release path, and environment are set in [`deploy.php`](../deploy.php). Frontend builds and deployments stay on Cloudflare Pages.

Add repository secrets `DEPLOY_HOST` (VPS IP address) and `PROD_PRIVATE_KEY` (SSH private key for the `uptaris` VPS account). Authorize its public key on the host and allow that account to clone this repository over SSH. Install Go 1.26 or newer and configure user systemd service `uptaris-api.service`. Its working directory must be the deployed `current` release root and its command should run `bin/uptaris-api`; keep `.env` in Deployer's shared deployment directory. The account also needs permission to connect to PostgreSQL and apply migrations.

Each deployment builds API and migration binaries on the VPS, applies pending database migrations, publishes release, then restarts user service. Migrations run before release publication; back up production data before schema changes.

## VPS configuration

Provision a dedicated `uptaris` operating-system account, PostgreSQL role, and database. Install the binaries and migrations together; run the migration binary from the directory containing `migrations` before starting the API. Back up the database before upgrading.

Use [`.env.example`](../.env.example) as the configuration reference. Set `APP_ENV=production`, the production `DATABASE_URL`, and independent random signing/refresh secrets generated with `openssl rand -hex 32`. Keep secrets outside the release bundle and restrict access to the service account or service manager.

Set `API_HOST=127.0.0.1`, `API_PORT=8080`, `TRUSTED_PROXIES=127.0.0.1/32`, and `CORS_ORIGINS=https://uptaris.linux123123.com`. Use `COOKIE_SECURE=true` and `COOKIE_SAME_SITE=lax` for the two HTTPS subdomains. Keep PostgreSQL private; use authenticated TLS for remote database connections.

Configure a service manager to run the API under the dedicated account. Configure nginx to terminate HTTPS for `uptaris-api.linux123123.com`, redirect HTTP to HTTPS, and proxy requests to `127.0.0.1:8080` while preserving paths and forwarding client IP/protocol headers. Add per-client authentication rate limits. Service and proxy configuration are managed outside this repository.

Validate nginx configuration before reloading, then check both API addresses:

```bash
sudo nginx -t
curl --fail http://127.0.0.1:8080/healthz
curl --fail https://uptaris-api.linux123123.com/healthz
```

## Initial administrator

Register the intended administrator through the normal UI. Registration grants viewer access. From a trusted database administration session, promote that active account to `admin` and revoke its existing sessions in the same transaction. Sign in again, then use the Users page to grant other roles. Do not run demo seeding in production.

## Release checks and recovery

Check public status, registration/sign-in, refresh after reload, nested server/monitor/incident CRUD, and viewer/operator/admin access. Confirm direct navigation to frontend routes works. Record deployed URLs and screenshots in the wiki report after these checks.

Retain the previous binaries and frontend build. If a release fails, restore the previous artifacts and restart the service. If a migration changes the schema incompatibly, use the database backup and a reviewed recovery plan; there is no `down` SQL migration in this repository.

## Logs and error reporting

The API writes JSON logs to stdout; use your service manager to collect them. Set `SENTRY_DSN` to enable Sentry error reporting; leave it empty to use local logs only. Request bodies, cookies, headers, and query strings are removed from Sentry events. Configure retention and access on the hosting platform.

Application authentication limits apply per process. Configure per-client limits in your reverse proxy. Monitor targets are never executed; status and incident changes remain manual.
