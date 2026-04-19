# GroupAlarm MQTT Bridge

A lightweight Go service that polls the [GroupAlarm](https://app.groupalarm.com) API every 5 seconds for open alarms and publishes new ones to an MQTT broker. Runs as a plain Docker container **and** as a Home Assistant add-on — same image, no duplication.

## Repository Structure

```
.
├── src/                        # Go source code
│   ├── main.go
│   ├── go.mod
│   └── go.sum
├── docker/                     # Docker files
│   ├── Dockerfile
│   ├── docker-entrypoint.sh    # Handles both Docker (env vars) and HA (/data/options.json)
│   ├── docker-compose.yml
│   └── docker-build.sh
├── homeassistant/              # Home Assistant add-on manifest
│   ├── config.yaml
│   └── DOCS.md
├── .github/workflows/
│   └── docker-build.yml        # CI: builds & pushes on release
└── repository.yaml             # HA add-on repository descriptor
```

## How It Works

1. For each configured organisation ID, fetches open events from the GroupAlarm API
2. For each event, fetches associated alarms
3. Filters out already-seen alarm IDs (tracked in memory)
4. Publishes new alarm messages to the MQTT topic (replacing `{org}` with the organisation ID)

## Running with Docker Compose

Copy `docker/docker-compose.yml`, fill in the environment variables, then:

```bash
docker compose -f docker/docker-compose.yml up -d
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GROUPALARM_APIKEY` | — | GroupAlarm Personal Access Token |
| `GROUPALARM_ORGS` | — | Comma-separated organisation IDs |
| `MQTT_HOST` | — | MQTT broker hostname |
| `MQTT_PORT` | `1883` | MQTT broker port |
| `MQTT_USER` | — | MQTT username |
| `MQTT_PASSWORD` | — | MQTT password |
| `MQTT_TOPIC` | `pager/groupalarm/{org}` | Topic template; `{org}` is replaced per organisation |

## Building Locally

```bash
# Run directly (requires config.ini in src/)
cd src && go run main.go

# Build binary
cd src && go build -o ../app ./...

# Build Docker image from repo root
docker build -f docker/Dockerfile -t groupalarm .

# Multi-platform push (edit docker/docker-build.sh first)
./docker/docker-build.sh
```

## CI / Docker Registry

Pushing a GitHub release triggers `.github/workflows/docker-build.yml`, which builds and pushes a multi-arch image (`amd64`, `arm64`, `arm/v7`) to:

```
ghcr.io/Gurkengewuerz/groupalarm
```

## Home Assistant Add-on

The same Docker image works as a Home Assistant add-on. When the container detects `/data/options.json` (present in all HA add-on containers), it reads configuration from there instead of environment variables.

### Setup

1. In HA go to **Settings → Add-ons → Add-on Store → ⋮ → Repositories**
2. Add `https://github.com/Gurkengewuerz/groupalarm`
3. Install **GroupAlarm MQTT Bridge** and configure it via the add-on UI

See `homeassistant/DOCS.md` for the full option reference.

### Publishing

Create a GitHub release — the workflow builds and pushes the image automatically. Then update `homeassistant/config.yaml` and `repository.yaml` with your GitHub username.

## Known Limitations

- Seen alarm IDs are tracked in memory only — restarting the service re-publishes alarms from currently open events
