# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository layout

Monorepo for Timeful (formerly Schej.it), a group availability/scheduling app.

- `frontend/` — Vue 2 + Vuetify + Tailwind single-page app (Vue CLI). Built output lands in `frontend/dist`.
- `server/` — Go (Gin) HTTP API backed by PostgreSQL through GORM. Also serves the built frontend as static files at the root.
- `compose.yaml` — Docker Compose: `postgres` + `frontend` (build-only, writes dist to a shared volume) + `server` (binds `127.0.0.1:3002`, mounts the dist volume read-only). See `DEPLOYMENT.md`.
- `PLUGIN_API_README.md` — `window.postMessage` API used by browser plugins to read/write availability on the frontend.

Internal identifiers (Go module `schej.it/server`, prod email "Schej.it") still use the old name — leave them alone unless rebranding is the explicit task.

## Common commands

### Frontend (`cd frontend`)
- `npm run serve` — dev server with hot reload (port 8080).
- `npm run build` — production build into `frontend/dist`.
- `npm run test:unit` — Vitest (config in `vitest.config.mjs`, matches `src/**/*.test.js`, alias `@` → `src/`).
- `npm run test:unit:watch` — Vitest watch mode.
- Run a single test: `npx vitest run src/utils/date_utils.test.js` (or `-t "test name"`).

### Backend (`cd server`)
- `air` — live-reload dev (install: `go install github.com/cosmtrek/air@latest`). Runs `main.go`, listens on `:3002` (`:3003` if `NODE_ENV=staging`).
- `go run main.go` — run without live reload. Pass `-release` to force `GIN_MODE=release`.
- `go test ./...` — run all Go tests.
- `go test ./db -run TestName` — run a single test (e.g. `./services/microsoftgraph`, `./services/gcloud`, `./services/listmonk`).
- `swag init` (in `server/`) — regenerate Swagger docs in `server/docs/` after editing route comments. Swagger UI is served at `http://localhost:3002/swagger/index.html`.
- PostgreSQL backup/restore: `pg_dump "$DATABASE_URL" -Fc -f timeful.dump` / `pg_restore --clean --if-exists -d "$DATABASE_URL" timeful.dump`.

### Required env vars for local server boot
`SESSION_SECRET` (≥32 chars) is enforced at startup. `CLIENT_ID`/`CLIENT_SECRET` (Google OAuth) and `ENCRYPTION_KEY` are required for most flows. See `server/.env.template` and `DEPLOYMENT.md` for the full list (Stripe, Microsoft, Listmonk, Slack, Discord, Gmail, etc.).

For local frontend → local backend, set `CORS_ORIGINS=http://localhost:8080` in `server/.env`.

## Architecture

### Backend (Gin + PostgreSQL/GORM)
`server/main.go` wires everything: CORS, cookie sessions, GORM init (`db.Init`), Google Cloud Tasks init (`services/gcloud.InitTasks`), then mounts API groups under `/api` via `routes.Init*` and `slackbot.InitSlackbot`. After API routes, it walks `frontend/dist` and registers each file as a static route, loads `index.html` as a template, and falls back to a `NoRoute` handler that injects per-route OG meta tags (e.g. for `/e/:eventId` it looks up the event to set the title and OG image).

- `routes/` — HTTP handlers grouped by domain: `auth.go`, `user.go`, `users.go`, `events.go`, `folders.go`, `analytics.go`, `stripe.go`. Route comments use Swag annotations; `swag init` regenerates `docs/`.
- `models/` — API and GORM models (`Event`, `User`, `Response`, `Folder`, `Attendee`, `Calendar`, `Set`, `Otp`, `FriendRequest`, `Location`, `DailyUserLog`).
- `db/` — GORM accessors per model (`events.go`, `users.go`, `folders.go`, `analytics.go`, `utils.go`) plus `init.go`.
- `services/` — external integrations. Notable: `calendar/` (Google, Outlook/Graph, Apple CalDAV via `jonyTF/go-webdav`, generic ICS), `auth/`, `contacts/`, `gcloud/` (Cloud Tasks for scheduled jobs), `listmonk/`, `microsoftgraph/`.
- `middleware/auth.go` — session-based auth middleware applied selectively by `routes.Init*`.
- `slackbot/` and `discord_bot/` — bot integrations registered as additional handlers.
- `utils/` — generic helpers (`array_utils`, `db_utils`, `mail_utils`, `request_utils`, `response_utils`).
- `logger/` — wraps log file (`logs.log`) + stdout via `gin.DefaultWriter`.

### Frontend (Vue 2 SPA)
- `src/router/index.js` — routes (`Landing`, `Home`, `Event`, `Group`, `Friends`, `Settings`, `SignIn`/`SignUp`/`Auth`, `StripeRedirect`, etc. — see `src/views/`).
- `src/store/index.js` — single Vuex store (auth user, events, snackbar, dialogs).
- `src/components/` — organized by feature folder (`event/`, `groups/`, `home/`, `landing/`, `pricing/`, `settings/`, `schedule_overlap/`, `calendar_permission_dialogs/`, `sign_up_form/`, `general/`) plus top-level shared components.
- `src/utils/` — date math (`date_utils.js`, uses `dayjs`/`moment`/`spacetime`), `fetch_utils.js` (API client), `plugin_utils.js` (handles the postMessage plugin API — see `PLUGIN_API_README.md`), `sign_in_utils.js`, `location_utils.js`, `services/` (calendar-provider abstractions on the client side).
- Tailwind + Vuetify coexist; `tailwind.config.js` purges `src/**/*.{vue,js,...}`.
- Service worker is registered via `register-service-worker`; `kill-sw.js` at the repo root is a kill switch script if needed.

### Frontend ↔ backend contract
- Same-origin in production: Caddy → Go on `:3002`, Go serves `/api/*` and falls through to `index.html` for SPA routes.
- Local dev: Vue CLI serves `:8080`, frontend calls `http://localhost:3002/api/*` (must whitelist via `CORS_ORIGINS`). Session cookie is `session` (cookie store, signed with `SESSION_SECRET`).
- Event IDs may be either a 24-character long ID or a short ID; `db.GetEventByEitherId` handles both — prefer it when looking up events from route params.

### Plugin (browser extension) API
The frontend exposes `get-slots` / `set-slots` over `window.postMessage` with a `FILL_CALENDAR_EVENT` type and `requestId` for response matching. Implementation lives in `src/utils/plugin_utils.js`; spec in `PLUGIN_API_README.md`. Don't change message shapes without also updating that doc.

## Conventions worth knowing

- The Go module path is `schej.it/server`; imports use that prefix throughout. Don't rename.
- GORM migrations run during startup. Add indexes and constraints through model tags or explicit idempotent migrations.
- New API routes need Swag comments above the handler so `swag init` picks them up; otherwise they're invisible in `/swagger`.
- The server panics on startup if `SESSION_SECRET` is missing or shorter than 32 chars (`validateSessionSecret` in `main.go`).
- `frontend/dist` is consumed by the Go server at runtime — local server boot tries `./frontend/dist` then `../frontend/dist`, or honors `FRONTEND_DIST` env var.
