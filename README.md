# Vigil: Unraid Energy Saver

Vigil is a standalone power management system for your home server (Unraid). It consists of two parts:
1.  **Vigil Controller**: An always-on web server (running on a Pi, VPS, or Router) acting as the "Brain" and Telegram Bot.
2.  **Vigil Agent**: A lightweight binary running on your Unraid server that reports status and performs shutdowns.

## Features
-   **Telegram Control**: Chat with your server to `/status`, `/wake`, `/shutdown`, or `/keepawake`.
-   **Prometheus Metrics**: Export load and status metrics at `/metrics`.
-   **Smart Shutdown**: Agent checks system load and asks Controller for permission to sleep.
-   **Direct Mode**: Simple `shutdown -h now` execution.
-   **Robust Config**: Supports `.env`, flags, and environment variables (via Cobra/Viper).

## Prerequisites
-   Go 1.23+
-   Access to your Unraid terminal (to run the Agent).
-   An always-on device for the Controller.

## Configuration (.env)
Create a `.env` file in the root directory:

```bash
# Controller
PORT=8080
MAC_ADDRESS=XX:XX:XX:XX:XX:XX # Unraid MAC for WOL
TELEGRAM_BOT_TOKEN=your-token
TELEGRAM_CHAT_ID=your-chat-id
LOAD_THRESHOLD=1.5 # CPU Load avg (15 min) to consider "Idle"

# Agent
CONTROLLER_URL=http://<CONTROLLER_IP>:8080
```

## Vigil Controller
The Controller runs on your always-on device.

```bash
# Run
go run cmd/vigil/main.go controller

# Help
go run cmd/vigil/main.go controller --help
```

### Endpoints
-   **Telegram Bot**: Interact via your Telegram app.
-   **Metrics**: `http://localhost:8080/metrics`

## Vigil Agent
The Agent runs on your Unraid server.

```bash
# Run (Dry Run Mode - Safe for testing)
go run cmd/vigil/main.go agent --dry-run
# or
./vigil agent --dry-run

# Run (Production)
go run cmd/vigil/main.go agent
```

## Telegram Commands
-   `/status`: Check if server is online and view current load.
-   `/wake`: Send Wake-on-LAN packet.
-   `/shutdown`: Request a forced shutdown (Agent will pick this up on next heartbeat).
-   `/keepawake [on|off]`: Prevent the server from sleeping regardless of load.

## Deployment (Docker / GHCR)

This project includes a multi-stage `Dockerfile` and a GitHub Actions pipeline to automatically publish images to GHCR.

### Pulling from GHCR
```bash
docker pull ghcr.io/<your-username>/vigil-controller:latest
docker pull ghcr.io/<your-username>/vigil-agent:latest
```

### Building Locally
```bash
# Build Unified Binary
go build -o vigil cmd/vigil/main.go

# Build Controller Image
docker build --target controller -t vigil-controller:latest .

# Build Agent Image
docker build --target agent -t vigil-agent:latest .
```

### Manual Binary Build
To build tiny binaries:

```bash
# Build Controller
CGO_ENABLED=0 go build -o vigil-controller cmd/controller/main.go

# Build Agent (for Linux/Amd64 Unraid)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o vigil-agent cmd/agent/main.go
```
