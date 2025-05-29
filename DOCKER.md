# Docker Deployment Instructions

This document provides instructions for deploying the Frigate Alerter using Docker.

## Prerequisites

- Docker and Docker Compose installed on your system
- A running Frigate CCTV server that's accessible from the Docker host
- A Discord bot token and channel ID

## Quick Start

1. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/frigate_alerter.git
   cd frigate_alerter
   ```

2. Create your environment file:
   ```bash
   cp .env.example .env
   ```

3. Edit the `.env` file with your specific configuration:
   ```bash
   # Edit with your favorite text editor
   nano .env
   ```

4. Create a data directory for persistent storage:
   ```bash
   mkdir -p data
   ```

5. Build and start the container:
   ```bash
   docker-compose up -d
   ```

6. Access the web UI at `http://your-docker-host:5555`

## Configuration Options

### Environment Variables

All configuration can be set using environment variables in the `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| FRIGATE_SERVER | Hostname/IP of your Frigate server | 192.168.200.1 |
| FRIGATE_PORT | Port of your Frigate server | 5000 |
| MQTT_SERVER | MQTT broker URL (usually the same as Frigate) | tcp://192.168.200.1:1883 |
| DISCORD_TOKEN | Your Discord bot token | - |
| DISCORD_CHANNEL_ID | Your Discord channel ID | - |
| TIME_ZONE | Timezone for timestamps | Asia/Kolkata |
| SERVER_PORT | Port for the web UI | 5555 |

### Volumes

The Docker Compose setup includes the following volumes:

- `./data:/app/data` - Persistent storage for the SQLite database
- `./config.json:/app/config.json:ro` - Optional configuration file (read-only)

## Updating

To update to a new version:

1. Pull the latest code:
   ```bash
   git pull
   ```

2. Rebuild and restart the container:
   ```bash
   docker-compose down
   docker-compose up -d --build
   ```

## Troubleshooting

### View Logs

To see logs from the container:

```bash
docker-compose logs -f
```

### Debug Mode

For more detailed logs, modify the `docker-compose.yml` file to add:

```yaml
environment:
  - LOG_LEVEL=debug
```

### Common Issues

1. **Cannot connect to Frigate/MQTT server**:
   - Ensure the IP addresses in your `.env` file are correct
   - Make sure the Docker host can reach your Frigate server
   - Check if any firewalls are blocking the connections

2. **Discord alerts not working**:
   - Verify your Discord bot token and channel ID
   - Ensure the bot has the necessary permissions in your server

3. **Database errors**:
   - Check if the `data` directory has proper permissions
