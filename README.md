# Proxyglass

Proxyglass is a read-only dashboard for one or more HAProxy stats endpoints. It reads HAProxy CSV stats and presents frontends, backends and backend servers. There are no controls that change HAProxy state.

## Features

- Config-file-based HAProxy targets with per-target HTTP Basic Auth.
- Filters for HAProxy instances, frontends, backends and servers.
- Sort by server name, status, current sessions or bytes received.
- Configurable default auto-refresh; 0 disables it. Users can select Off, 2, 5, 10 or 30 seconds from custom styled menus. Preferences persist in a browser cookie.
- Frontend sessions/rate/peak, totals, bytes and denials/errors; backend/server sessions, bytes, status and health-check status/code/duration/last-check detail.
- Native Go binary and systemd installer, DEB/RPM packages, plus optional container deployment.
- OIDC login with ID-token signature verification and optional group allow-list; local bearer tokens also supported.
- Browser assets are bundled locally. The UI makes no external CDN/font/JavaScript requests and works air-gapped. App outbound connections can use `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY`.
- Standalone SVG brand assets live in [`assets/brand/`](assets/brand/) for reuse on a future website; `make assets` syncs them into the embedded dashboard assets.

## Build and run

Requires Go 1.25+.

```sh
cp config.example.yaml config.yaml
# Configure targets and credentials in config.yaml.
make build-native
./bin/proxyglass --config ./config.yaml
```

Authentication is enabled in the example config. Generate a bearer token and sign in at `/login`:

```sh
./bin/proxyglass --generate-token --config ./config.yaml
```

Token commands:

```sh
proxyglass --generate-token --config /etc/proxyglass/config.yml
proxyglass --generate-token --quiet --config /etc/proxyglass/config.yml
proxyglass --revoke-token TOKEN --config /etc/proxyglass/config.yml
```

`make build` targets the current host. `make build-linux-amd64`, `make build-linux-arm64` and `make build-all` cross-compile static Linux binaries. Set build version using `VERSION=...`.

## Configuration

See [`config.example.yaml`](config.example.yaml) and [`packaging/default-config.yml`](packaging/default-config.yml).

```yaml
server:
  listen: "127.0.0.1:8080" # nginx can proxy to this local listener
  default_refresh: 5        # seconds; 0 disables auto-refresh

require_auth: true
tokens_file: "/etc/proxyglass/tokens.json"
token_ttl_days: 0           # 0 means no expiry

oidc:
  enabled: false
  issuer: "https://sso.example.net/realms/operations"
  client_id: "proxyglass"
  client_secret: "configure-me"
  session_ttl_hours: 8
  redirect_uri: "https://proxyglass.example.net/ui/oidc/callback"
  allowed_groups: []

targets:
  - name: edge-01
    url: "http://haproxy.example.net:8404/stats;csv"
    username: stats-reader
    password: secret
    verify_tls: true # true by default; false only for self-signed TLS
```

`PROXYGLASS_CONFIG` selects the config file. For OIDC behind nginx, preserve the original `Host` and set `X-Forwarded-Proto`; register `/ui/oidc/callback` with the identity provider. Proxyglass does not terminate TLS.

## Native install and systemd

Build a Linux binary or install a DEB/RPM from a tagged release. A standalone Linux binary can install its service layout on a systemd host:

```sh
sudo ./proxyglass --install
sudoedit /etc/proxyglass/config.yml
sudo proxyglass --generate-token --config /etc/proxyglass/config.yml
sudo systemctl enable --now proxyglass
```

The installer creates the `proxyglass` system user, `/etc/proxyglass/`, `/var/lib/proxyglass/`, `/var/log/proxyglass/`, `/etc/default/proxyglass` and systemd unit. Existing config/environment files are preserved. `proxyglass --uninstall` retains configuration and data.

Optional outbound proxy settings go in `/etc/default/proxyglass`; restart the service after changes:

```sh
HTTPS_PROXY=http://proxy.corp.example:3128
HTTP_PROXY=http://proxy.corp.example:3128
NO_PROXY=localhost,127.0.0.1,.internal
```

These standard environment variables also work for direct binary and container deployments. Use `NO_PROXY` for private HAProxy endpoints as needed.

## Container and Nomad

The container is optional; the native binary/systemd install is the primary deployment. Build with `docker build -f docker/Dockerfile -t proxyglass .`, mount config at `/etc/proxyglass/config.yml` and expose port 8080. See [`deploy/proxyglass.nomad.hcl`](deploy/proxyglass.nomad.hcl) for a Nomad example.

## Product website

The Astro marketing site for **proxyglass.app** is in [`site/`](site/). It builds to static HTML and deploys as Cloudflare Worker assets; see [`site/README.md`](site/README.md) for local development and Worker setup. No external browser CDNs or Easter eggs are included.

## HAProxy test lab

See [`docs/TESTING.md`](docs/TESTING.md). `test/scripts/setup.sh` starts three HAProxy containers and test backends, enables stats with test-only Basic Auth and sends requests to create nonzero counters. One intentionally unavailable server demonstrates down/check-error states. Test credentials are `admin / changeme`; never reuse them.

## Development and CI

```sh
make test
make lint
make build-native
make build-all
make css # npm required; generated CSS is committed for offline builds
```

GitHub Actions runs Go tests/vet, cross-builds Linux binaries, verifies generated CSS, and publishes amd64/arm64 binaries plus DEB/RPM packages for `v*` tags.
