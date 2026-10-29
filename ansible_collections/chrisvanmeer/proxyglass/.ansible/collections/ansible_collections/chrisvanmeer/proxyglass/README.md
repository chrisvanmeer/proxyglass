# Ansible Collection - chrisvanmeer.proxyglass

Collection to install and manage [Proxyglass](https://github.com/chrisvanmeer/proxyglass).

## Usage

```yaml
- name: Install Proxyglass
  hosts: proxyglass
  become: true
  gather_facts: true
  roles:
    - role: chrisvanmeer.proxyglass.proxyglass
```

Or use the provided playbook:

```bash
ansible-playbook -i inventory ansible_collections/chrisvanmeer/proxyglass/playbooks/install_proxyglass.yml
```

## Configuration

All variables can be set via `group_vars`/`host_vars`. See `roles/proxyglass/defaults/main.yml` for the full list.

### Key variables

| Variable | Default | Description |
| --- | --- | --- |
| `proxyglass_version` | `latest` | GitHub release tag (e.g. `v0.1.6`) or `latest` |
| `proxyglass_temp_dir` | `/tmp` | Download location on the target. Override per host/group |
| `proxyglass_install_path` | `/usr/local/bin/proxyglass` | Target binary path |
| `proxyglass_config_path` | `/etc/proxyglass/config.yml` | Config file path |
| `proxyglass_log_dir` | `/var/log/proxyglass` | Log directory, owned by the Proxyglass service account |
| `proxyglass_log_file` | `/var/log/proxyglass/proxyglass.log` | Application log file |
| `proxyglass_service_name` | `proxyglass` | Systemd service name |
| `proxyglass_force_noexec_check` | `true` | Fail if the temp mount has `noexec` |
| `proxyglass_allow_noexec_fallback_dir` | `""` | Optional exec fallback temp dir |
| `proxyglass_generate_token_if_missing` | `true` | Generate an auth token if missing (idempotent) |
| `proxyglass_token_file` | `/etc/proxyglass/.proxyglass_auth_token` | Path where the token is stored (if generated) |

## Publishing

The collection version in `galaxy.yml` is independent of Proxyglass application
releases. Update that version and push a matching `collection-v<version>` tag
(for example, `collection-v1.0.1`) to publish to Ansible Galaxy. The
`GALAXY_API_KEY` repository secret is used by the publish workflow. The same
workflow can also be run manually from GitHub Actions.
