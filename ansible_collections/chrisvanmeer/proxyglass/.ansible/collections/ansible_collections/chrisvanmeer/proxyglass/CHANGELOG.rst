===============================
chrisvanmeer.proxyglass Release Notes
===============================

.. contents:: Topics


v1.0.0
======

Release Summary
---------------

Initial release of the chrisvanmeer.proxyglass collection with the install role.

Major Changes
-------------

- Add ``install`` role to download, install and manage Proxyglass from GitHub releases
- Support version pinning or ``latest`` (resolved via GitHub API)
- Support configurable temp directory with noexec mount detection and optional fallback
- Idempotent token generation
- Molecule-based testing with login verification
- CI workflow for testing and publishing to Ansible Galaxy
