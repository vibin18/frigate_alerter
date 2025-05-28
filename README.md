# Frigate CCTV Alerter

A Go service that listens to Frigate CCTV MQTT events and sends alerts to a Discord channel when new detections occur. 

## Features

- Listens to Frigate MQTT events
- Sends Discord notifications with camera image when new detections occur
- Stores alerts in SQLite database
- Configurable via environment variables or config.json
- Timezone support for accurate timestamps

## Architecture

This project follows hexagonal architecture (ports and adapters):

- **Domain**: Core business entities and logic
- **Ports**: Interfaces that define boundaries between the application and external systems
- **Adapters**: Implementations of the interfaces to interact with external systems
- **Application**: Business logic orchestration
- **Config**: Configuration management

## Requirements

- Go 1.16 or higher
- Frigate CCTV system
- MQTT broker (the one Frigate publishes to)
- Discord bot with channel access

## Installation

1. Clone the repository
2. Install dependencies:

```bash
go mod tidy
```

3. Configure the application (see Configuration section)
4. Build the application:

```bash
go build -o frigate_alerter ./cmd/frigate_alerter
```

## Configuration

Configuration can be provided via environment variables or a `config.json` file:

```json
{
  "frigate_server": "localhost",
  "frigate_port": "5000",
  "mqtt_server": "tcp://localhost:1883",
  "discord_token": "your-discord-bot-token",
  "discord_channel_id": "your-discord-channel-id",
  "time_zone": "UTC",
  "server_port": "8080"
}
```

### Environment Variables

- `FRIGATE_SERVER`: Frigate server hostname/IP (default: "localhost")
- `FRIGATE_PORT`: Frigate server port (default: "5000")
- `MQTT_SERVER`: MQTT broker URL (default: "tcp://localhost:1883")
- `DISCORD_TOKEN`: Discord bot token
- `DISCORD_CHANNEL_ID`: Discord channel ID for notifications
- `TIME_ZONE`: Timezone for alert timestamps (default: "UTC")
- `SERVER_PORT`: Server port for future HTTP interface (default: "8080")

## Running the Service

```bash
./frigate_alerter
```

The service will:
1. Connect to the MQTT broker
2. Subscribe to Frigate events
3. Process "new" detection events
4. Send alerts to Discord
5. Store alerts in an SQLite database

## Discord Integration

1. Create a Discord bot at https://discord.com/developers/applications
2. Add the bot to your server with appropriate permissions
3. Get the bot token and channel ID
4. Configure these in the application


