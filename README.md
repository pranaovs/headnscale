# Headnscale

Automatically generates DNS records for Headscale `extra_records.json` based on Docker container labels and/or Tailscale network peers.
It is meant to supplement Docker label based reverse proxy setups (e.g., Traefik, Nginx Proxy Manager, etc.) when using Headscale as Tailscale control server.

Refer: <https://github.com/juanfont/headscale/blob/main/docs/ref/dns.md>

## Features

- **Docker Source**: Discovers containers with specific labels and creates DNS records
- **Tailscale Source**: Discovers peers on your Tailscale/Headscale network
- **Hosts File Sink**: Writes records to a hosts file format
- **Headscale Sink**: Writes records to Headscale's `extra_records.json` format

All sources and sinks are optional and enabled via environment variables.

## Quick Start

### Using Pre-built Image from GHCR

1. Create a `docker-compose.yml` file (or use the provided one)
2. Update environment variables with your settings
3. Run: `docker compose up -d`

### Environment Variables

#### Common Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_BASE_DOMAIN` | No | `ts.net` | Base domain for DNS records |
| `HEADNSCALE_NO_BASE_DOMAIN` | No | `false` | Create additional records without base domain |
| `HEADNSCALE_REFRESH_SECONDS` | No | `60` | How often to scan for changes (for non-callback based sources) |
| `HEADNSCALE_WILDCARD` | No | `false` | Create wildcard DNS records for subdomains |
| `HEADNSCALE_STATE_DIR` | No | `/var/lib/headnscale` | Persistent state directory |
| `HEADNSCALE_TS_SERVE` | No | - | Set to `true` to serve HTTP and DNS over Tailscale network ([Configuration](#tailscale-source)) |

#### Sources

##### Docker Source

Set `HEADNSCALE_DOCKER_ENABLED` to `true` to enable.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_DOCKER_ENABLED` | No | `false` | Set to `true` to enable Docker source |
| `HEADNSCALE_LABEL_KEY` | No | `headnscale.subdomain` | Docker label key to look for |
| `HEADNSCALE_NODE_HOSTNAME` | Yes | - | Hostname of the node running the containers |
| `HEADNSCALE_NODE_IP` | Yes | - | IPv4 address of the node |
| `HEADNSCALE_NODE_IP6` | No | - | IPv6 address of the node |
| `DOCKER_HOST` | No | `unix:///var/run/docker.sock` | Docker host socket path |
| `DOCKER_CONTEXT` | No | - | Docker context to use |

##### Tailscale Source

Set `HEADNSCALE_TS_ENABLED` to `true` to enable.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_TS_ENABLED` | No | `false` | Set to `true` to enable Tailscale source (read peers from tailnet) |
| `TS_AUTHKEY` | No | - | Tailscale auth key |
| `HEADNSCALE_TS_LOGIN_SERVER` | No | - | Tailscale/Headscale login server URL |
| `HEADNSCALE_TS_HOSTNAME` | No | `headnscale` | Hostname for this instance on the Tailscale network |


#### Sinks

##### Headscale `extra_records.json` Sink

Set `HEADNSCALE_JSON_ENABLED` to `true` to enable.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_JSON_ENABLED` | No | `false` | Set to `true` to enable Headscale sink |
| `HEADNSCALE_JSON_PATH` | Yes | - | Path to write the extra_records.json file |


##### Hosts File Sink

Set `HEADNSCALE_HOSTS_ENABLED` to `true` to enable.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_HOSTS_ENABLED` | No | `false` | Set to `true` to enable Hosts sink |
| `HEADNSCALE_HOSTS_PATH` | Yes | - | Path to write the hosts file |
| `HEADNSCALE_HOSTS_PORT` | No | `80` | Port to serve hosts file under `/hosts` and `/hosts.txt` |
| `HEADNSCALE_HOSTS_TS_PORT` | No | `HEADNSCALE_HOSTS_PORT` | Port on tailnet to serve hosts file under `/hosts` and `/hosts.txt` |


##### DNS Sink

Set `HEADNSCALE_DNS_ENABLED` to `true` to enable.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_DNS_ENABLED` | No | `false` | Set to `true` to enable DNS sink |
| `HEADNSCALE_DNS_PORT` | No | `53` | DNS port for serving DNS queries |
| `HEADNSCALE_DNS_TS_PORT` | No | `HEADNSCALE_DNS_PORT` | DNS port on tailnet for serving DNS queries |


### Deployment

#### Both Sources and Tailscale Serving

```yaml
services:
  headnscale:
    image: ghcr.io/pranaovs/headnscale:latest
    container_name: headnscale
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /var/lib/headscale:/data
      - ./state:/var/lib/headnscale
    ports:
      - 8080:8080
    environment:
      # Docker source
      - HEADNSCALE_DOCKER_ENABLED=true
      - HEADNSCALE_NODE_HOSTNAME=<Tailscale Hostname>
      - HEADNSCALE_NODE_IP=<Tailscale IPv4>
      - HEADNSCALE_NODE_IP6=<Tailscale IPv6>
      # Tailscale source (read peers from tailnet)
      - HEADNSCALE_TS_ENABLED=true
      # Tailscale serve (serve HTTP/DNS over tailnet)
      - HEADNSCALE_TS_SERVE=true
      - TS_AUTHKEY=<your-auth-key>
      - HEADNSCALE_TS_LOGIN_SERVER=https://headscale.example.com
      # Sinks
      - HEADNSCALE_JSON_ENABLED=true
      - HEADNSCALE_JSON_PATH=/data/extra_records.json
      - HEADNSCALE_HOSTS_ENABLED=true
      - HEADNSCALE_HOSTS_PATH=/data/hosts.txt
      - HEADNSCALE_DNS_ENABLED=true
      - HEADNSCALE_DNS_PORT=53
      # Common
      - HEADNSCALE_BASE_DOMAIN=ts.net
      - HEADNSCALE_NO_BASE_DOMAIN=true
```

### Usage

#### Docker Source

Label your containers:

```yaml
services:
  myapp:
    image: myapp
    labels:
      - "traefik.http.routers.myapp.rule=Host(`myapp.your-node-hostname.ts.net`) || Host(`myapp.your-node-hostname`) || Host(`app.your-node-hostname.ts.net`) || Host(`app.your-node-hostname`)""
      - "headnscale.subdomain=myapp|app"
```

A DNS record will be created for `myapp.your-node-hostname.ts.net` -> `HEADNSCALE_NODE_IP`.

#### Tailscale Source

All peers on your Tailscale/Headscale network will automatically have DNS records created based on their hostname and IP addresses when `HEADNSCALE_TS_ENABLED=true`.

#### Tailscale Serving

When `HEADNSCALE_TS_SERVE=true`, the HTTP and DNS servers (if enabled) will also be accessible over your Tailscale network. This allows you to access the services from any device on your tailnet without exposing them to the internet.

## Building from Source

```bash
docker build -t headnscale .
```

## GitHub Actions

This repository automatically builds and pushes Docker images to GitHub Container Registry on:

- Latest tag (tagged as `latest`)
- Every push to `main` branch (tagged as `main`)
- Every tagged release (tagged as version numbers)

The image is available at: `ghcr.io/pranaovs/headnscale:latest`

---

_Disclaimer: README.md, Dockerfile and .github/ created using Claude Sonnet 4.5 (GitHub Copilot). Please report any problems/inconsistencies if found._
