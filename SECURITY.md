# Security Policy

## Supported Versions

| Version        | Supported  |
| -------------- | ---------- |
| Latest release | ✅ Current |

Only the latest release receives security fixes. Please update to the latest
version before reporting issues found in older versions.

modloc reads every file of the repositories it is given and runs one other
program, `scc`, on them. It writes nothing and opens no network connection.

## Security Practices

- **Small attack surface.** The standard library and public dependencies
  only; `scc` runs as a separate process, never as a library inside modloc.
- **Checked inputs.** Every path is checked before anything is scanned, and
  a bad command line is rejected before anything runs.
- **Scanned in CI.** Every change runs `govulncheck`, `golangci-lint` and a
  secret scan; the Nix build pins every action of the workflow to a commit.

## Reporting a Vulnerability

Please do **not** open public issues for security problems. Report privately
instead:

- **GitHub Security Advisory**: https://github.com/LarsArtmann/modloc/security/advisories/new
- **Email**: security@artmann.tech

### What to include

- Vulnerability type and severity
- Steps to reproduce
- Potential impact
- Proof of concept (if available)

### Response commitments

| Phase            | Timeline                    |
| ---------------- | --------------------------- |
| Initial response | Within 48 hours             |
| Triage           | Within 72 hours             |
| Disclosure       | Coordinated, within 90 days |
