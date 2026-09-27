# Project instructions

## Code style

- Write simple, readable code. Use clear names and early returns; avoid unnecessary abstractions, duplication, and compatibility layers.
- Add blank lines between logical steps, guard clauses, functions, and final returns. Keep related declarations together.
- Expand long expressions, configuration objects, and query chains onto multiple lines.
- Comment intent, security decisions, and concurrency requirements. Do not restate obvious code.
- Prefer established libraries for OAuth, password validation, QR codes, and UI behavior.

## Frontend

- Use Bun, TypeScript, React, and existing TanStack patterns.
- Keep page implementations in `frontend/src/routes`, directly in their route files.
- Put reusable components in appropriate folders: `components/dialogs`, `forms`, `feedback`, and `layout`. Small shared widgets can stay in `components`.
- Keep each custom reusable component in its own file. Small render callbacks can stay with their caller.
- Use shadcn components where possible; primitives belong in `components/ui`.
- Run `bunx shadcn add <component>` from `frontend/`. Keep `components.json`, TypeScript aliases, and Vite aliases aligned with `src/`.
- Import React hooks and types by name. No React namespace imports or `React.xxx`.
- Derive navigation types from TanStack Router; do not hardcode allowed destination strings.
- Route files own validated search parameters. Forms receive values through props.

## Backend

- Keep one database model per file in `backend/internal/models`. DTOs, filters, and aggregate results belong outside that package.
- Route registration belongs in `backend/routers`; HTTP handling in `internal/handlers`; request/response types in their respective packages.
- Business operations accept `context.Context` and return errors. Keep Gin out of business packages.
- Keep OAuth generic, with provider adapters using `golang.org/x/oauth2`. Password management belongs to account settings.
- Never link accounts by email. Linking requires an authenticated flow; signup requires a verified provider email.
- Issue normal sessions only after all required factors pass. Preserve transaction locks, one-time state, and replay protection.
- Keep secrets out of logs and browser storage. Store provider tokens encrypted.

## Tools and checks

Use `rg` and `rg --files` for searches. Preserve existing user edits; do not stage or commit unless requested.

From `frontend/`:

```bash
bunx eslint . --fix
bun run format
bun run check
bun run lint
bun run format:check
bun run build
```

From `backend/`:

```bash
gofmt -w .
go vet ./...
go build ./...
```

After handler annotation changes, run `go generate ./cmd/api` from `backend/`. Do not hand-edit generated API docs or `frontend/src/routeTree.gen.ts`.

Do not add or run tests or E2E suites unless explicitly requested. Report completed checks and remaining issues briefly.

Keep `docs/development.md` limited to run/check commands, environment keys, and defaults. Do not add implementation explanations or key creation tutorials.
