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

## Example playbook

```yaml
- hosts: proxyglass
  become: true
  roles:
    - role: chrisvanmeer.proxyglass.proxyglass
```
