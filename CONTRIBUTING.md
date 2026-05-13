# Contributing to Samuel

Thanks for being here. A few things before you open a PR.

## Where work belongs

| You want to … | Open it here |
|---|---|
| Report a framework bug | [Bug report](https://github.com/samuelpkg/samuel/issues/new?template=bug_report.yml) on this repo |
| Propose a small change | [Feature request](https://github.com/samuelpkg/samuel/issues/new?template=feature_request.yml) on this repo |
| Propose a big design change (new tier, schema change, cross-cutting concern) | [Open a discussion](https://github.com/samuelpkg/samuel/discussions) first; an RFD usually follows |
| Fix a docs page | [Docs issue](https://github.com/samuelpkg/samuel/issues/new?template=docs.yml) or PR direct |
| File a plugin bug | The plugin's own repo (`github.com/samuelpkg/samuel-<name>`), not this one |
| Report a vulnerability | [Private security advisory](https://github.com/samuelpkg/samuel/security/advisories/new) — see [SECURITY.md](SECURITY.md) |
| Ask a question | [Discussions](https://github.com/samuelpkg/samuel/discussions) |

## Local development

Go 1.24+, no other prerequisites.

```sh
git clone git@github.com:samuelpkg/samuel.git
cd samuel
go build ./...
go test -race ./...
golangci-lint run ./...
```

For docs work:

```sh
pip install -r requirements-docs.txt
mkdocs serve              # local preview at http://127.0.0.1:8000/
mkdocs build --strict     # what CI runs on every push
```

## Conventions

- **Commits**: conventional commits (`feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, etc.). One logical change per commit.
- **Function ≤ 50 lines, file ≤ 300 lines.** Same guardrails Samuel itself enforces on projects it manages.
- **Tests carry the feature.** New code lands with tests; bug fixes land with a regression test.
- **Errors are structured.** Use `internal/errors` to build errors with `Fix:` and `DocsURL:` fields; do not return bare `fmt.Errorf`.
- **No emojis in code, docs, or commit messages.**
- **AGENTS.md is canonical.** The framework does not write or read `CLAUDE.md`, `.claude/`, `.cursor/`, or `.codex/` files — those belong in translator plugins. The CI gate at `.github/workflows/agnostic-check.yml` enforces this.
- **AGENTS.md template ≤ 150 lines** (source and rendered). CI gate at `.github/workflows/agents-md-check.yml`.

## Before you open the PR

Run the preflight script:

```sh
bash scripts/release-checklist.sh
```

It runs the build, tests, lint, AGENTS.md budget, agnostic grep, RFD index freshness, CHANGELOG check, and the goreleaser config validator in one pass.

## RFDs

Substantial design changes go through an RFD (Request for Discussion) under [docs/rfd/](docs/rfd/). The format and the source-of-truth index live in [rfd-index.toml](rfd-index.toml). Read a few of the existing RFDs (0001–0008) for the shape; the structure is Summary → Problem → Background → Options → Decision → Implementation → Outcome.

## Code of conduct

Participation is governed by the [Contributor Covenant](CODE_OF_CONDUCT.md). Reports go to [angel@cuemby.com](mailto:angel@cuemby.com) or via private security advisory.
