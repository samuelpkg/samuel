---
title: internal/github/client.go
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-github-ui]
tags: [v1, github, refactor]
---

# github.Client

Simple HTTP wrapper around GitHub API + raw.githubusercontent.com. 295 lines.

## Capabilities

| Method | Purpose |
|---|---|
| `GetLatestRelease()` | `GET /repos/{owner}/{repo}/releases/latest`. Returns nil + nil err when no releases. |
| `GetLatestVersionOrBranch()` | Latest release version (v-prefix stripped) OR `("dev", true)` fallback. |
| `GetTags()` | `GET /repos/{owner}/{repo}/tags`. |
| `DownloadArchive(version)` | `<owner>/<repo>/archive/refs/tags/v<version>.tar.gz`. Returns ReadCloser + ContentLength. |
| `DownloadBranchArchive(branch)` | `<owner>/<repo>/archive/refs/heads/<branch>.tar.gz`. |
| `DownloadFile(version, path)` | `raw.githubusercontent.com/<owner>/<repo>/v<version>/<path>`. Limit-reader at 10 MB. |
| `CheckForUpdates(current)` | Returns `VersionInfo{Current, Latest, UpdateNeeded, ReleaseNotes}`. |

## Defaults

- HTTP timeout: **30 seconds**.
- Max single-file download: **10 MB**.
- Headers on API: `Accept: application/vnd.github.v3+json`.
- Headers on every request: `User-Agent: samuel-cli`.

## v2 implications #refactor

### Drop the v1 model

In v1, this client is tied to the Samuel repo (default `ar4mirez/samuel`). Every install pulls a Samuel tarball, then extracts subsets of `template/`. This couples the registry to the framework — to add a skill, you bump Samuel.

v2 inverts: every plugin is its own repo. The plugin loader instantiates a `github.Client{owner, repo}` per plugin source.

### What ports

- The shape: `Client{httpClient, owner, repo}` + timeout + User-Agent + size cap + limit-reader.
- The URL templates: archive, branch archive, raw file.
- The `nil release` → `("dev", true)` fallback pattern is right for prerelease projects.
- v-prefix stripping.

### What extends

- **Authentication**: `Authorization: Bearer <token>` from `SAMUEL_GITHUB_TOKEN` or `GH_TOKEN`. Required for private plugin repos.
- **Sigstore fetch**: pull `.sig` and `.crt` alongside archive for verification.
- **OCI sibling**: `oci.Client` for the OCI tier of [[concepts/plugin-format]] — likely via `oras-project/oras-go`.

### Don't reinvent

GitHub doesn't have a great Go SDK for this narrow use case (`google/go-github` is too much). A 200-line custom client is the right cut. Keep it small.
