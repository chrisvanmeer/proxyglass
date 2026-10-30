# Proxyglass role

This role installs and manages [Proxyglass](https://github.com/chrisvanmeer/proxyglass).

## Requirements

- Linux target with systemd
- Network access to GitHub Releases (or a configured proxy)

## Role variables

See `defaults/main.yml` for the full list. Key variables:

- `proxyglass_version`: `latest` or a tag such as `v0.1.6`
- `proxyglass_temp_dir`: temporary download location
- `proxyglass_install_path`: binary destination (default `/usr/local/bin/proxyglass`)
- `proxyglass_config_path`: config destination (default `/etc/proxyglass/config.yml`)
- `proxyglass_service_name`: systemd unit name (default `proxyglass`)
- `proxyglass_force_noexec_check`: fail when the temp mount is `noexec`
- `proxyglass_allow_noexec_fallback_dir`: optional executable fallback temp directory
- `proxyglass_http_proxy`, `proxyglass_https_proxy`, `proxyglass_no_proxy`: proxy settings for downloads
- `proxyglass_generate_token_if_missing`: generate an admin token when missing

### Managing the config file

By default the role enforces the config file on every run (`proxyglass_manage_config: true`).
Set to `false` to leave an existing config untouched.

The role manages `/etc/proxyglass/config.yml`. You can either let the role
render it from individual variables, or provide the full file content.

**Individual sections** (used when `proxyglass_config_content` is empty):

- `proxyglass_listen`, `proxyglass_default_refresh`
- `proxyglass_require_auth`, `proxyglass_token_ttl_days`
- `proxyglass_oidc_enabled`, `proxyglass_oidc_issuer`, `proxyglass_oidc_client_id`
- `proxyglass_oidc_client_secret`, `proxyglass_oidc_session_ttl_hours`, `proxyglass_oidc_allowed_groups`
- `proxyglass_targets`: list of HAProxy target dicts
- `proxyglass_topology`: list of topology route dicts

**Full override:**

```yaml
proxyglass_config_content: |
  server:
    listen: ":8080"
  require_auth: true
  targets:
    - name: lb-01
      url: http://lb:8404/stats;csv
```

## Example playbook

```yaml
- hosts: proxyglass
  become: true
  roles:
    - role: chrisvanmeer.proxyglass.proxyglass
```
