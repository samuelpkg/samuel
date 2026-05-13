# Samuel v2 — release-day runbook

Manual operations that ship v2. Cannot be automated end-to-end:
each step is destructive or external, and benefits from a human
making the last call.

Run `scripts/release-checklist.sh --candidate <stage>` before
each tag.

## Stage 0 — Pre-flight (anytime before rc.2)

- [ ] `bash scripts/release-checklist.sh --candidate rc.2` exits 0.
- [ ] Working tree clean.
- [ ] `docs/getting-started/migration-v1.md` published in repo.
- [ ] `scripts/v1-deprecation/Formula-samuel.rb` reviewed (drop into
      tap repo on stage 4, not now).

## Stage 1 — v2.0.0-rc.2

PRD Tasks 7.0.

```sh
# 1. Final commit with CHANGELOG ## [v2.0.0-rc.2] entry
git add CHANGELOG.md
git commit -m "chore(release): cut v2.0.0-rc.2 (Polish + Launch)"

# 2. Tag (signed)
git tag -s v2.0.0-rc.2 -m "v2.0.0-rc.2"

# 3. Push
git push origin main
git push origin v2.0.0-rc.2
# CI runs goreleaser; signed artifacts publish to Releases.

# 4. Announce rc.2 — collect feedback for 1 week.
```

## Stage 2 — v2.0.0-rc.3 (after ~1 week soak)

PRD Tasks 8.0.

- [ ] Triage rc.2 feedback. Categorize: blocker / serious / nice-to-have.
- [ ] Fix blockers + serious. Defer nice-to-have to v2.0.x patches.
- [ ] Update CHANGELOG `## [v2.0.0-rc.3]` with the fixes.

```sh
git tag -s v2.0.0-rc.3 -m "v2.0.0-rc.3"
git push origin v2.0.0-rc.3
```

## Stage 3 — Final soak (~1 week)

PRD Tasks 9.0.

- [ ] Triage rc.3 feedback. Apply final fixes.
- [ ] Consolidate the CHANGELOG: collapse rc.2 + rc.3 + final
      fixes into the single `## [v2.0.0]` entry (the rc entries can
      stay as historical anchors).

## Stage 4 — v2.0.0 GA

PRD Tasks 10.0 + 11.0.

```sh
# 1. Make sure rc.3 + your latest fixes are on main.
git checkout main
git pull

# 2. Run preflight one more time.
bash scripts/release-checklist.sh --candidate ga
# Must be green.

# 3. Tag v1-final at the LAST v1 commit BEFORE the v2 force-push.
#    Get the last v1 commit hash from samuel_v1 / GitHub history.
git -C ../samuel_v1 log -1 --format=%H
# Use that hash; for example:
git tag -s v1-final <v1-final-sha> -m "v1 final — preserved for archaeology"
git push origin v1-final

# 4. Tag v2.0.0.
git tag -s v2.0.0 -m "v2.0.0"
git push origin v2.0.0
# CI publishes signed artifacts; goreleaser writes the Homebrew formula
# to samuelpkg/homebrew-tap with this version's SHAs.

# 5. Force-push main to v2 (this is the destructive step that
#    replaces v1's main with v2's). DO NOT skip the v1-final tag.
git push origin main --force
# (or `--force-with-lease` for safety)

# 6. Replace the repo README from the v1-deprecation drop-in.
#    The page that lands on github.com/samuelpkg/samuel needs to be
#    v2's README (already in this repo as README.md after the force-push).
#    The v1-final tag preserves v1's README at the tag URL.

# 7. Update install.sh on samuelpkg/samuel:main — already done
#    (this repo's install.sh is v2's). Verify the curl path works.

# 8. Verify on clean macOS + Linux containers:
docker run --rm alpine sh -c 'apk add curl && \
  curl -sSL https://raw.githubusercontent.com/samuelpkg/samuel/main/install.sh | sh && \
  samuel version'

# 9. brew install (tap is ar4mirez/homebrew-tap — carried over from v1
#    so existing brew users transparently upgrade):
brew update
brew install ar4mirez/tap/samuel
samuel version    # should print v2.0.0
```

## Stage 5 — Docs deploy

PRD Tasks 11.6.

```sh
mkdocs gh-deploy --strict --force
# or the GitHub Action equivalent.
```

Verify [samuelpkg.github.io/samuel](https://samuelpkg.github.io/samuel/) shows v2 docs.

## Stage 6 — Announce

PRD Tasks 12.0.

- [ ] Publish `docs/blog/samuel-v2-launch.md` to chosen channel.
- [ ] (Optional) Post on HN, lobste.rs, relevant Slack/Discord
      communities.
- [ ] Pin the announcement issue on the repo.

## Stage 7 — Post-launch

PRD Tasks 13.0.

- [ ] Open v2.1 issues: `samuel plugin new/build/publish`,
      additional translator plugins (cursor / continue / aider),
      Sigstore signed-default, hot-reload design.
- [ ] Set monthly v2.0.x patch cadence.
- [ ] Watch community plugin authoring; help where useful.

## Safety

The release-checklist script is read-only. The destructive steps are
deliberately not automated:

- `git tag -s` — needs a GPG key the operator owns.
- `git push --force origin main` — replaces v1 history pointer.
  Tag `v1-final` first.
- `brew tap update` — operator-driven.

If the rc soak finds a blocker that requires schema changes, the
right call is to ship `v2.1.0-rc.1` (additive change), not to
re-cut `v2.0.0-rc.4` (semver discipline).
