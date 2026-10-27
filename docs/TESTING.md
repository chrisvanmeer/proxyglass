# Testing Proxyglass with HAProxy

The test lab uses Docker Compose to start three HAProxy instances, each protected with Basic Auth on its stats page, plus local nginx containers as backend servers.

## Start the lab

From the repository root:

```sh
./test/scripts/setup.sh
```

This builds/starts the Proxyglass test container and HAProxy lab, waits for each CSV stats page, and sends 150 requests through each public frontend. This generates genuine HAProxy session and byte counters rather than static/faked CSV values. HAProxy 1 has one intentionally unavailable backend server so the dashboard also shows a down state and check error.

Open <http://localhost:18000>. The setup script creates and prints a disposable bearer token; paste it into the login page. The token is stored only in the test container's temporary filesystem. In a separate local/native run, copy `test/config.docker.yaml` and change target URLs to `http://localhost:18404`, `28404` and `38404`.

Stats pages (Basic Auth `admin` / `changeme`):

- <http://localhost:18404/stats>
- <http://localhost:28404/stats>
- <http://localhost:38404/stats>

To generate another traffic burst:

```sh
./test/scripts/traffic.sh 250
```

To inspect CSV directly:

```sh
curl -u admin:changeme 'http://localhost:18404/stats;csv'
```

## Stop and cleanup

```sh
docker compose -f docker/docker-compose.test.yaml down
```

The lab publishes host ports 18000, 18080, 18404, 28080, 28404, 38080 and 38404. Check these are free before starting. Data is ephemeral and lives in containers only.

## Configuration details

HAProxy stats are enabled on port 8404 with `stats uri /stats` and `stats auth admin:changeme`. The config files under `test/haproxy/` define multiple frontends/backends and checked nginx servers. Stats CSV is available by appending `;csv` to the stats URI. Do not expose this test setup publicly or reuse its credentials.
