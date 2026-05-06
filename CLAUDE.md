# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Go service that polls the GroupAlarm API for open events/alarms and publishes new alarms to an MQTT broker. Runs as a Docker container in production and also as a Home Assistant add-on.

## Commands

```bash
# Run locally (requires config.ini in src/)
cd src && go run main.go

# Build binary
cd src && go build -o ../app ./...

# Build Docker image from repo root
docker build -f docker/Dockerfile -t groupalarm .

# Build multi-platform image (edit docker/docker-build.sh first)
./docker/docker-build.sh

# Run via Docker Compose
docker compose -f docker/docker-compose.yml up
```

No lint or test commands are configured.

## Repository Structure

```
src/          Go source code (main.go, go.mod, go.sum)
docker/       Dockerfile, docker-entrypoint.sh, docker-compose.yml, docker-build.sh
homeassistant/ Home Assistant add-on manifest (config.yaml, DOCS.md) — reuses the main image
.github/      GitHub Actions workflows
repository.yaml  HA add-on repository descriptor
```

## Architecture

Single-file application (`src/main.go`). The main loop runs every 5 seconds using a `time.Ticker`:

1. For each configured organisation ID, fetch open events from `https://app.groupalarm.com/api/v1/events/open`
2. For each event, fetch associated alarms from `https://app.groupalarm.com/api/v1/alarms`
3. For each alarm:
   - If not yet seen: publish the raw message and event title to the alarm's base topic (once)
   - On every tick: publish feedback counters (`positive`/`negative`/`unknown`) and a JSON meta payload as subtopics — these update live as responders react
4. Topic template placeholders: `{id}` → alarm ID (default), `{org}` → organisation ID (legacy)

Graceful shutdown is handled via `os.Signal` (SIGINT/SIGTERM).

## Configuration

Config is loaded from `config.ini` (INI format). In Docker, `docker/docker-entrypoint.sh` generates this file from environment variables at startup. In the HA add-on, `homeassistant/run.sh` reads from `/data/options.json` (HA supervisor format).

| Section | Key | Env var |
|---------|-----|---------|
| `groupalarm` | `api_key` | `GROUPALARM_APIKEY` |
| `groupalarm` | `organisations` | `GROUPALARM_ORGS` (comma-separated IDs) |
| `mqtt` | `host` | `MQTT_HOST` |
| `mqtt` | `port` | `MQTT_PORT` |
| `mqtt` | `user` | `MQTT_USER` |
| `mqtt` | `password` | `MQTT_PASSWORD` |
| `mqtt` | `topic` | `MQTT_TOPIC` (supports `{id}` = alarm ID, `{org}` = org ID) |
| `mqtt` | `client` | — |

## Docker Build Context

The `docker/Dockerfile` and `homeassistant/Dockerfile` both use the **repo root** as their Docker build context. Always build from the repo root:

```bash
docker build -f docker/Dockerfile .
docker build -f homeassistant/Dockerfile .
```

## CI

`.github/workflows/docker-build.yml` triggers on GitHub releases and pushes:
- `ghcr.io/{owner}/groupalarm` — single image used for both Docker and HA add-on

## Key Limitations

- **Memory growth**: Seen alarm IDs accumulate in RAM indefinitely; long-running instances will grow.
- **No persistence**: Restarting re-publishes alarms from currently open events.
- **No tests**: There is no test infrastructure.
