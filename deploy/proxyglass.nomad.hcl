job "proxyglass" {
  datacenters = ["dc1"]
  type        = "service"

  group "dashboard" {
    count          = 1
    shutdown_delay = "5s"

    # Register a writable host volume named "proxyglass-data" on eligible nodes
    # to keep generated bearer tokens across task/container restarts.
    volume "proxyglass-data" {
      type      = "host"
      source    = "proxyglass-data"
      read_only = false
    }

    network {
      port "http" {
        to = 8080
      }
    }

    service {
      name     = "proxyglass"
      port     = "http"
      provider = "nomad"

      check {
        name     = "health"
        type     = "http"
        path     = "/healthz"
        interval = "15s"
        timeout  = "3s"
      }
    }

    task "proxyglass" {
      driver = "docker"

      config {
        image = "ghcr.io/chrisvanmeer/proxyglass:latest"
        ports = ["http"]
        args  = ["--config", "local/config.yml"]
      }

      volume_mount {
        volume      = "proxyglass-data"
        destination = "/var/lib/proxyglass"
        read_only   = false
      }

      env {
        HTTP_PROXY  = ""
        HTTPS_PROXY = ""
        NO_PROXY    = "localhost,127.0.0.1"
      }

      template {
        destination = "local/config.yml"
        change_mode = "restart"
        data = <<-EOF
server:
  listen: ":8080"
  default_refresh: 5
require_auth: true
tokens_file: "/var/lib/proxyglass/tokens.json"
token_ttl_days: 0
log_path: ""
oidc:
  enabled: false
targets:
  - name: edge-01
    url: "http://haproxy.service.consul:8404/stats;csv"
    username: "{{ with nomadVar "nomad/jobs/proxyglass" }}{{ .stats_user }}{{ end }}"
    password: "{{ with nomadVar "nomad/jobs/proxyglass" }}{{ .stats_password }}{{ end }}"
EOF
      }

      resources {
        cpu    = 200
        memory = 128
      }
    }
  }
}
