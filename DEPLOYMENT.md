# Timeful Deployment Guide

Production deployment using Docker Compose behind a Caddy reverse proxy.

## Prerequisites

- Docker and Docker Compose
- Caddy on the host (for reverse proxy + automatic HTTPS, although you can use any reverse proxy)
- Domain with DNS pointing to your server

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/schej-it/timeful.app
cd timeful.app

# 2. Create server environment file
cp server/.env.template server/.env
# Edit server/.env with your values (see Configuration below)

# 3. Build and start services
docker compose up -d --build

# 4. Configure Caddy
sudo cp Caddyfile.example /etc/caddy/Caddyfile
# Edit /etc/caddy/Caddyfile with your domain
sudo systemctl reload caddy
```

## Services

| Service    | Description                                  | Port           |
| ---------- | -------------------------------------------- | -------------- |
| `postgres` | PostgreSQL 16 database                       | Internal only  |
| `server`   | Go backend with embedded Vue frontend assets | 127.0.0.1:3002 |

The application container writes `config.js` when it starts. These environment
variables are read at container startup rather than embedded during image build:

| Variable                       | Browser configuration       |
| ------------------------------ | --------------------------- |
| `VUE_APP_POSTHOG_API_KEY`      | PostHog project API key     |
| `VUE_APP_GOOGLE_CLIENT_ID`     | Google OAuth client ID      |
| `VUE_APP_MICROSOFT_CLIENT_ID`  | Microsoft OAuth client ID   |
| `VUE_APP_PRIVACY_POLICY_URL`    | Optional privacy policy URL |

All four values are public because browsers can read them. Leave
`VUE_APP_PRIVACY_POLICY_URL` empty to use the bundled privacy policy page. The
custom URL must allow iframe embedding. The image also accepts `CLIENT_ID` and
`MICROSOFT_CLIENT_ID` as fallbacks for their frontend counterparts, so existing
values in `server/.env` work with Compose. Restart the server container after
changing them. The server marks `config.js` as non-cacheable, so browsers
receive the new values.

## Caddy

The example Caddyfile proxies all traffic to the Go backend on port 3002. Caddy handles:

- Automatic HTTPS certificates
- HTTP → HTTPS redirect
- www → non-www redirect
- Compression (gzip/zstd)
- Security headers

Edit `/etc/caddy/Caddyfile` with your domain before reloading.

## Commands

```bash
docker compose up -d              # Start services
docker compose logs -f            # View logs
docker compose logs -f server     # View specific service logs
docker compose up -d --build      # Rebuild after code changes
docker compose down               # Stop services
docker compose down -v            # Stop and remove volumes (deletes data!)
```

## Data & Backup

Data is persisted in Docker volumes: `postgres_data` and `server_logs`.

```bash
# Backup PostgreSQL
docker compose exec -T postgres pg_dump -U timeful -d timeful -Fc > timeful.dump

# Restore PostgreSQL
docker compose exec -T postgres pg_restore -U timeful -d timeful --clean --if-exists < timeful.dump
```

## Troubleshooting

```bash
# Container won't start
docker compose logs server
ls -la server/.env

# PostgreSQL connection issues
docker compose ps
docker compose exec postgres pg_isready -U timeful -d timeful

# Frontend not loading
docker compose exec server ls -la /app/frontend/dist
```

---

## Configuration

### Required Environment Variables

Create `server/.env` from the template (`server/.env.template`).

#### Required

| Variable         | Description                                                                 |
| ---------------- | --------------------------------------------------------------------------- |
| `CLIENT_ID`      | Google OAuth client ID                                                      |
| `CLIENT_SECRET`  | Google OAuth client secret                                                  |
| `ENCRYPTION_KEY` | Key for encrypting sensitive data (generate with `openssl rand -base64 32`) |
| `SESSION_SECRET` | Session cookie encryption key (generate with `openssl rand -base64 32`)     |
| `DATABASE_URL`   | PostgreSQL connection URL (provided by Compose in production)               |

#### Optional — Payments

| Variable                | Description                        |
| ----------------------- | ---------------------------------- |
| `STRIPE_API_KEY`        | Stripe API key                     |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret      |
| `STRIPE_*_PRICE_ID`     | Stripe price IDs for various plans |

#### Optional — Additional Calendars

| Variable                  | Description                             |
| ------------------------- | --------------------------------------- |
| `MICROSOFT_CLIENT_ID`     | Microsoft OAuth client ID (for Outlook) |
| `MICROSOFT_CLIENT_SECRET` | Microsoft OAuth client secret           |

#### Optional — CORS

| Variable       | Description                                                                                                          |
| -------------- | -------------------------------------------------------------------------------------------------------------------- |
| `CORS_ORIGINS` | Comma-separated allowed origins (default: production domains). For local development, set to `http://localhost:8080` |

#### Optional — Other Services

| Variable                                     | Description                                  |
| -------------------------------------------- | -------------------------------------------- |
| `ANALYTICS_USERNAME` / `ANALYTICS_PASSWORD`  | Basic auth for /api/analytics routes         |
| `SERVICE_ACCOUNT_KEY_PATH`                   | Google Cloud service account for Cloud Tasks |
| `SLACK_*_WEBHOOK_URL`                        | Slack webhooks for notifications             |
| `GMAIL_APP_PASSWORD` / `SCHEJ_EMAIL_ADDRESS` | Gmail SMTP for sending emails                |
| `LISTMONK_*`                                 | Listmonk email service configuration         |
| `DISCORD_BOT_TOKEN` / `GUILD_ID`             | Discord bot integration                      |

See `server/.env.template` for the complete list.

### PostgreSQL

Set a strong database password before starting Compose:

```bash
POSTGRES_PASSWORD='replace-with-a-long-random-secret' \
  docker compose up -d --build
```

Compose sets `DATABASE_URL` for server. GORM migrates relational tables for
users, events, responses, attendees, folders, folder-event links, daily logs,
friend requests, and OTP codes during startup.

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable the following APIs:
   - Google Calendar API
   - People API (Contacts)
   - Admin SDK API (Directory)
4. Create OAuth 2.0 credentials (Web application type)
5. Add authorized redirect URIs:
   - `https://yourdomain.com/api/auth/callback`
   - `http://localhost:3002/api/auth/callback` (for development)
6. Copy the Client ID and Client Secret to your `.env`
