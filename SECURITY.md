# Security policy

## Supported versions

| Version | Supported |
|---------|-----------|
| 2.0.x   | yes       |
| < 2.0   | no — see [v1-final](https://github.com/samuelpkg/samuel/tree/v1-final) |

## Reporting a vulnerability

Use [GitHub's private security advisories](https://github.com/samuelpkg/samuel/security/advisories/new). Do not open a public issue.

What to include:

- Affected version (`samuel version` output).
- Reproduction steps. Minimal, with the exact commands you ran.
- Expected vs observed behavior.
- Impact assessment if you have one (data exposure, code execution, capability escape, etc.).

We aim to acknowledge within 5 business days and to triage within 14.

## Scope

In scope:

- The framework (`github.com/samuelpkg/samuel`).
- Built-in commands (`samuel install`, `samuel run`, etc.).
- The capability model, the plugin verifier, the WASM sandbox.
- The reusable plugin release workflow (`samuelpkg/samuel-plugin-release`).

Out of scope (report to the plugin's own repo):

- Third-party plugins under `github.com/samuelpkg/samuel-<name>` or any community plugin.
- Vulnerabilities in upstream dependencies — those go to the upstream.
- Issues caused by running `--allow-unsigned` or granting risky capabilities the framework already warned about.

## Coordinated disclosure

Default: 90-day disclosure window once a fix is available. We can adjust on request. Credits in the release notes unless you ask otherwise.
