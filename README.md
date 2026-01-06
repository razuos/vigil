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
# Run (Dry Run)
./vigil agent --dry-run

# Run (Production - Sleep Mode [Default])
./vigil agent --shutdown-mode sleep

# Run (Production - Full Power Off)
./vigil agent --shutdown-mode off
```

## Telegram Commands
-   `/status`: Check if server is online and view current load.
-   `/wake`: Send Wake-on-LAN packet.
-   `/shutdown`: Request a forced shutdown (Agent will pick this up on next heartbeat).
-   `/keepawake [on|off]`: Prevent the server from sleeping regardless of load.

## Deployment (Docker / GHCR)

### Pulling from GHCR
```bash
docker pull ghcr.io/<your-username>/vigil:latest
```

### Running with Docker

#### Standard (Generic)
```bash
# Controller
docker run -d --restart always -p 8080:8080 --env-file .env vigil:latest controller

# Agent (Sleep Mode)
docker run -d --restart always --network host --privileged vigil:latest agent

# Agent (Power Off Mode)
docker run -d --restart always --network host --privileged vigil:latest agent --shutdown-mode off
```

#### Unraid (Agent)
Run this command in the Unraid terminal to start the Agent (using Sleep mode):
```bash
docker run -d \
  --name vigil-agent \
  --restart always \
  --privileged \
  --net=host \
  -e CONTROLLER_URL="http://<CONTROLLER_IP>:8080" \
  -e SHUTDOWN_MODE="sleep" \
  ghcr.io/<your-username>/vigil:latest agent
```
*Replace `<CONTROLLER_IP>` with the IP where your Controller is running.*
*Set `SHUTDOWN_MODE` to `off` if you want a full shutdown instead of sleep.*

### Building Locally
```bash
docker build -t vigil:latest .
```

### Manual Binary Build
To build a tiny unified binary:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o vigil cmd/vigil/main.go
```
