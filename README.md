# Overwatch

**Infrastructure monitoring** from the command line: know when services, endpoints, certificates, and scheduled jobs fail—without living in a browser.

Overwatch ships as a **single Go binary**. Run `overwatch serve` to start a self-hosted monitoring server with checks and alerts defined in YAML. Use the CLI to manage everything.

---

## Install

### macOS (Homebrew)

```bash
brew install processfoundry/tap/overwatch
```

### Linux

```bash
curl -sLO "https://github.com/processfoundry/overwatch/releases/latest/download/overwatch_linux_amd64.tar.gz"
tar xzf overwatch_linux_amd64.tar.gz
sudo mv overwatch /usr/local/bin/
```

For ARM64:

```bash
curl -sLO "https://github.com/processfoundry/overwatch/releases/latest/download/overwatch_linux_arm64.tar.gz"
tar xzf overwatch_linux_arm64.tar.gz
sudo mv overwatch /usr/local/bin/
```

### Windows

```powershell
Invoke-WebRequest "https://github.com/processfoundry/overwatch/releases/latest/download/overwatch_windows_amd64.tar.gz" -OutFile overwatch.tar.gz
tar xzf overwatch.tar.gz
New-Item -ItemType Directory -Force -Path "C:\overwatch" | Out-Null
Move-Item overwatch.exe "C:\overwatch\overwatch.exe" -Force
# Add C:\overwatch to your PATH if not already present:
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\overwatch", "User")
```

### From source

```bash
go install github.com/christianmscott/overwatch/cmd/overwatch@latest
```

Or clone and build:

```bash
git clone https://github.com/christianmscott/overwatch.git
cd overwatch
go build -o overwatch ./cmd/overwatch
```

---

## Quick start

### 1. Start the server

```bash
overwatch serve
```

This creates a starter `overwatch.yaml` if one doesn't exist, generates a **join token**, and starts the API + scheduler on `127.0.0.1:3030`.

Use flags to customize:

```bash
overwatch serve --bind-address 0.0.0.0 --bind-port 3030
```

The join token is printed to stderr on startup — copy it for the next step.

### With Docker Compose

```yaml
services:
  overwatch:
    build:
      context: .
      dockerfile: packaging/docker/Dockerfile
    ports:
      - "3030:3030"
    volumes:
      - ./overwatch.yaml:/overwatch.yaml
    command: ["--bind-address", "0.0.0.0"]
```

```bash
docker compose up -d
```

### With systemd (Linux)

Copy the provided unit file and config:

```bash
sudo cp packaging/systemd/overwatch.service /etc/systemd/system/
sudo mkdir -p /etc/overwatch
sudo cp overwatch.yaml /etc/overwatch/overwatch.yaml
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now overwatch
sudo journalctl -u overwatch -f   # follow logs
```

Reload config without restarting:

```bash
sudo systemctl reload overwatch
```

### Run as a Windows service

The simplest approach is [NSSM](https://nssm.cc/) (the Non-Sucking Service Manager). Download `nssm.exe` and place it somewhere on your PATH, then:

```powershell
# Install the service (adjust paths as needed)
nssm install Overwatch "C:\overwatch\overwatch.exe" "serve --bind-address 0.0.0.0 --config C:\overwatch\overwatch.yaml"
nssm set Overwatch AppDirectory "C:\overwatch"
nssm set Overwatch DisplayName "Overwatch Monitoring Server"
nssm set Overwatch Start SERVICE_AUTO_START
nssm set Overwatch AppStdout "C:\overwatch\overwatch.log"
nssm set Overwatch AppStderr "C:\overwatch\overwatch.log"

# Start the service
nssm start Overwatch
```

Manage with standard Windows service commands:

```powershell
nssm status Overwatch    # check status
nssm restart Overwatch   # restart (reloads config)
nssm stop Overwatch      # stop
nssm remove Overwatch    # uninstall
```

### 2. Connect a client

On any machine that should manage this server:

```bash
overwatch init
```

Select **option 2** ("Setup a client"), paste the join token from the server. This generates a keypair under `~/.overwatch/`, registers it with the server, and saves the connection config.

### 3. Use the CLI

```bash
overwatch status                # full status: checks, alerts, results
overwatch check list            # list checks
overwatch check add my-api \
  --type http \
  --target https://api.example.com \
  --interval 30s
overwatch alert add slack \
  --url https://hooks.slack.com/services/T.../B.../xxx
overwatch token                 # print the join token (to share with colleagues)
```

---

## Features

| Capability | Description |
|------------|-------------|
| **CLI** | `overwatch check` and `overwatch alert` (`add`, `list`, `remove`, `update`, `test`) plus `overwatch status` for verbose, tabular config—ideal for scripts and automation. |
| **Self-hosted server** | `overwatch serve` runs the API and scheduler. Monitor definitions live in **YAML** as the **source of truth** (edits via CLI or on disk; reload via SIGHUP or `POST /api/reload`). |
| **Check types** | HTTP/HTTPS, TCP, TLS certificate expiry, DNS, and **scheduled-job check-in** (webhook endpoint your jobs `curl` on success, with missed-window alerting and optional failure signaling). |
| **Alerts** | Outbound **webhooks** (Slack, Teams, PagerDuty, etc.) and **SMTP** (your relay). |
| **Auth** | Ed25519 client signatures. Clients join with a token; all subsequent requests are signed. |
| **Config** | `overwatch config init`, `overwatch config validate`, and `overwatch version`. |

---

## Commands

```text
overwatch                         # show help / setup prompt
overwatch init                    # interactive setup (server, client, or cloud)
overwatch serve                   # start the self-hosted server
overwatch status                  # verbose table of all checks & alerts + live results
overwatch check list|add|remove|update|test
overwatch alert list|add|remove|update|test
overwatch token                   # print the server's join token (authenticated)
overwatch config init             # generate a starter YAML config
overwatch config validate         # validate the config file
overwatch version                 # build/version metadata
```

Use `--help` on any command for flags and examples.

---

## Server configuration (YAML)

The YAML config file is the source of truth for the server. Example:

```yaml
server:
  bind_address: 127.0.0.1
  bind_port: 3030
  external_address: overwatch.example.com
  concurrency: 4

checks:
  - name: my-api
    type: http
    target: https://api.example.com
    interval: 30s
    timeout: 10s
    expected_status: 200
    alerts: [slack]

  - name: db
    type: tcp
    target: localhost:5432
    interval: 30s
    timeout: 5s

  - name: cert
    type: tls
    target: example.com:443
    interval: 1h
    timeout: 10s

  - name: nightly-backup
    type: checkin
    max_silence: 25h
    interval: 1m
    timeout: 5s

alerts:
  webhooks:
    - name: slack
      url: https://hooks.slack.com/services/...
      method: POST
      timeout: 10s
```

Edit the file and send SIGHUP or `POST /api/reload` to reload without restarting.

---

## Repository layout

- `cmd/overwatch` — main entrypoint
- `internal/` — implementation (CLI, server, checks, alerts, auth, …)
- `pkg/spec` — shared config and API types
- `packaging/` — Docker, systemd, launchd assets
- `examples/` — example configs

---

## Contributing

Issues and pull requests are welcome. Read [`CONTRIBUTING.md`](./CONTRIBUTING.md) first.

---

## License

[MIT License](./LICENSE)
