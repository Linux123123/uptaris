# Validation and API demonstration

## Backend and security checks

```bash
cd backend
go vet ./...
go build ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
```

## Interactive API demonstration

Start the normal development app and open `/swagger/index.html`. Use `/auth/login` with a demo identity and enter only the raw `accessToken` in **Authorize**; Swagger UI adds the `Bearer ` prefix automatically. Browser stores the HttpOnly refresh cookie automatically. Every domain route has a JSON schema and success/error examples.

## Long-lived API testing token

From `backend`, issue a token for an existing user. The example lifetime is 30 days:

```bash
go run ./cmd/token -email admin@uptaris.local -ttl 720h
```

Both flags are required. Durations use Go syntax (`24h`, `720h`, `8760h`); days such as `30d` are not supported. The command loads the same environment configuration as the API, including `DATABASE_URL` and `JWT_ACCESS_SECRET`. Run migrations and create or seed the user first.

Only the JWT is written to stdout, so it can be captured directly:

```bash
TOKEN=$(go run ./cmd/token -email admin@uptaris.local -ttl 720h)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/auth/me
```

The command creates a database session with the requested lifetime and uses the user's current role. It does not change the normal login token lifetime or issue a refresh cookie. Treat the token as a credential: anyone holding it has that user's API access until expiry or revocation. Revoke it with `POST /api/v1/auth/logout` using the same bearer token. User deletion or a role change also invalidates it.

## Frontend checks

From `frontend`, run `bun run check`, `bun run lint`, `bun run format:check`, and `bun run build`. Browser checks are manual: review desktop/mobile navigation, form validation, create/edit/delete dialogs, filter changes, pagination, loading/errors, and session restoration. Check each role with the demo accounts.

## Manual API walkthrough

Use Swagger UI at `http://localhost:8080/swagger/index.html` with a seeded development database. Expand an operation, select **Try it out**, provide its input, then select **Execute**. Record created IDs for nested paths. The [API reference](https://github.com/Linux123123/uptaris/wiki/API-Reference) lists all 26 operations.

1. Read public `/status`. Register a temporary viewer with `/auth/register` (201).
2. Sign in as the demo administrator with `/auth/login` (200). Paste the raw access token into **Authorize**. Read `/auth/me` and `/dashboard` (200).
3. Call `/auth/refresh` (200); the browser sends its refresh cookie. Replace the token in **Authorize** with the returned access token.
4. List `/admin/users` (200), then PATCH the temporary user's role to `operator` (200).
5. Create a temporary server (201). List, read, and PATCH it (200).
6. Under that server, create a monitor (201). List, read, and PATCH it (200).
7. Under that monitor, create an incident (201). List, read, and PATCH it (200). Read the aggregate `/incidents` list (200).
8. DELETE the incident, monitor, server, and temporary user (204 each). Call `/auth/logout` (204).

Verify these error and access cases separately:

- Missing, expired, or revoked token: 401. Reusing a token after logout must fail.
- Viewer attempting writes or accessing administrator routes: 403.
- Another user's inventory, wrong parent-child IDs, or deleted resources: 404.
- Malformed JSON, unknown fields, invalid IDs, or invalid list filters: 400.
- Invalid field values or missing required fields: 422.
- Duplicate registration or changing/deleting your own administrator account: 409.
- Unsupported body content type: 415. Body exceeding 64 KiB: 413.
- Role changes and account deletion revoke existing sessions.
- Deleting a server or monitor makes its descendants inaccessible.
- Concurrent PATCH requests changing different fields preserve both changes.

Clean up temporary records after interrupted checks. Keep passwords, access tokens, and refresh cookies out of screenshots and committed documentation. No Postman collection, Go test files, or browser test suite is maintained in the repository.
