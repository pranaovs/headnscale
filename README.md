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
| `HEADNSCALE_REFRESH_SECONDS` | No | `60` | How often to scan for changes |
| `HEADNSCALE_PORT` | No | `8080` | Port for internal HTTP server (health checks, hosts.txt) |

#### Docker Source (optional)

Set `HEADNSCALE_DOCKER_ENABLED` to enable.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_DOCKER_ENABLED` | No | - | Set to any value to enable Docker source |
| `HEADNSCALE_LABEL_KEY` | No | `headnscale.subdomain` | Docker label key to look for |
| `HEADNSCALE_NODE_HOSTNAME` | Yes* | - | Hostname of the node running the containers |
| `HEADNSCALE_NODE_IP` | Yes* | - | IPv4 address of the node |
| `HEADNSCALE_NODE_IP6` | No | - | IPv6 address of the node |
| `DOCKER_HOST` | No | `unix:///var/run/docker.sock` | Docker host socket path |
| `DOCKER_CONTEXT` | No | - | Docker Context |

*Required when Docker source is enabled.

#### Tailscale Source (optional)

Set `HEADNSCALE_TS_ENABLED` to enable.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_TS_AUTHKEY` | No | - | Tailscale auth key (enables Tailscale source when set) |
| `HEADNSCALE_TS_LOGIN_SERVER` | No | - | Tailscale/Headscale login server URL |
| `HEADNSCALE_TS_HOSTNAME` | No | `headnscale` | Hostname for this instance on the Tailscale network |

#### Sinks (at least one required)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HEADNSCALE_JSON_PATH` | No* | - | Path to write the extra_records.json file (enables Headscale sink) |
| `HEADNSCALE_HOSTS_PATH` | No* | - | Path to write the hosts file (enables Hosts sink) |

*At least one sink must be enabled.

### Deployment Examples

#### Docker Source Only

```yaml
services:
  headnscale:
    image: ghcr.io/pranaovs/headnscale:latest
    container_name: headnscale
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /var/lib/headscale:/data
    ports:
      - 8080:8080
    environment:
      - HEADNSCALE_DOCKER_ENABLED=true
      - HEADNSCALE_JSON_PATH=/data/extra_records.json
      - HEADNSCALE_NODE_HOSTNAME=<Tailscale Hostname>
      - HEADNSCALE_NODE_IP=<Tailscale IPv4>
      - HEADNSCALE_BASE_DOMAIN=ts.net
      - HEADNSCALE_NO_BASE_DOMAIN=true
```

#### Tailscale Source Only

```yaml
services:
  headnscale:
    image: ghcr.io/pranaovs/headnscale:latest
    container_name: headnscale
    restart: unless-stopped
    volumes:
      - /var/lib/headscale:/data
      - ./tsnet-state:/root/.config/tsnet-headnscale
    ports:
      - 8080:8080
    environment:
      - HEADNSCALE_TS_AUTHKEY=<your-auth-key>
      - HEADNSCALE_TS_LOGIN_SERVER=https://headscale.example.com
      - HEADNSCALE_TS_HOSTNAME=headnscale
      - HEADNSCALE_JSON_PATH=/data/extra_records.json
      - HEADNSCALE_HOSTS_PATH=/data/hosts.txt
```

#### Both Sources

```yaml
services:
  headnscale:
    image: ghcr.io/pranaovs/headnscale:latest
    container_name: headnscale
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /var/lib/headscale:/data
      - ./tsnet-state:/root/.config/tsnet-headnscale
    ports:
      - 8080:8080
    environment:
      # Docker source
      - HEADNSCALE_DOCKER_ENABLED=true
      - HEADNSCALE_NODE_HOSTNAME=<Tailscale Hostname>
      - HEADNSCALE_NODE_IP=<Tailscale IPv4>
      # Tailscale source
      - HEADNSCALE_TS_AUTHKEY=<your-auth-key>
      - HEADNSCALE_TS_LOGIN_SERVER=https://headscale.example.com
      # Sinks
      - HEADNSCALE_JSON_PATH=/data/extra_records.json
      - HEADNSCALE_HOSTS_PATH=/data/hosts.txt
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
      - "traefik.http.routers.myapp.rule=Host(`myapp.your-node-hostname.ts.net`) || Host(`myapp.your-node-hostname`)"
      - "headnscale.subdomain=myapp"
```

A DNS record will be created for `myapp.your-node-hostname.ts.net` -> `HEADNSCALE_NODE_IP`.

#### Tailscale Source

All peers on your Tailscale/Headscale network will automatically have DNS records created based on their hostname and IP addresses.

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
