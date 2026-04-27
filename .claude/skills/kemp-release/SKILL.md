---
name: kemp-release
description: Cut a new tagged release of terraform-provider-kemp. Runs the full preflight (clean tree, tests, vet), creates an annotated semver tag, pushes it, watches the GitHub Actions Release workflow until goreleaser signs and uploads the artifacts. Use when the user says "release v0.X.Y", "ship a new version", "tag and publish", or similar.
---

# Cutting a release

`terraform-provider-kemp` ships via a `v*` git tag → `Release` workflow →
goreleaser → signed GitHub release → Terraform Registry auto-ingest.
This skill walks through the safe sequence.

## Preflight (always run before tagging)

```bash
# Working tree must be clean and synced with origin/main.
git status
git log --oneline origin/main..HEAD       # should be empty

# All tests green, vet clean, build clean.
make test
make build
go vet ./...

# If you regenerated docs in this session, make sure they're committed.
make generate
git diff --quiet docs/ || echo "uncommitted doc changes — commit before tagging"
```

If any of those fail, fix and commit before continuing. The `Release`
workflow runs goreleaser fresh from the tag, but Registry rendering pulls
docs from the same commit — you don't want a tag pointing at code that
doesn't compile or docs that don't reflect the schema.

## Choose a version

Strict semver `vMAJOR.MINOR.PATCH`:

- **Patch** (`v0.1.0` → `v0.1.1`): doc updates, bug fixes, no schema
  changes that break existing state.
- **Minor** (`v0.1.X` → `v0.2.0`): new resources, new optional
  attributes on existing resources, anything additive.
- **Major** (`v0.X.X` → `v1.0.0` or `v1.X.X` → `v2.0.0`): renames,
  removed attributes, attribute type changes, anything that breaks
  existing user state. Don't do this casually pre-1.0.

## Create and push the tag

```bash
git tag -a vX.Y.Z -m "$(cat <<'EOF'
One-line summary of the release.

- Bullet list of user-visible changes
- Group by resource where it makes sense
- Reference issue/PR numbers if applicable
EOF
)"
git push origin vX.Y.Z
```

The `Release` workflow fires automatically on the push.

## Watch the workflow

```bash
# Find the run that just started
gh run list -R pier62350/terraform-provider-kemp --workflow=Release -L 3

# Watch it to completion (5-7 minutes typically)
gh run watch <RUN_ID> -R pier62350/terraform-provider-kemp --exit-status
```

Or run it as a background command and let the notification surface when
it lands — don't block the session waiting.

## Verify the GitHub release

```bash
gh release view vX.Y.Z -R pier62350/terraform-provider-kemp --json assets --jq '.assets[].name'
```

Expect **16 assets**: 13 zips (10 platforms after the `darwin/386` and
`windows/arm` ignore rules), `_SHA256SUMS`, `_SHA256SUMS.sig`, and
`_manifest.json`. If anything is missing, the Registry ingestion will
fail — investigate before declaring done.

## Registry ingestion

For the **first** release of a brand-new provider, you need to manually
connect the repo at https://registry.terraform.io → "Publish" → "Provider"
→ select the repo. Subsequent tags are auto-ingested within a few minutes
as long as the GPG signature on `SHA256SUMS.sig` validates against the
key registered at https://registry.terraform.io/settings/gpg-keys.

The provider page is `registry.terraform.io/providers/pier62350/kemp/X.Y.Z`.

## When something goes wrong

- **Workflow fails on GPG import** → the `GPG_PRIVATE_KEY` secret is wrong
  or the `PASSPHRASE` doesn't match. Re-export from the local keyring and
  set the secret again with `gh secret set`.
- **Workflow fails on goreleaser** → check the run logs for the exact
  step. Most common: a deprecation that became an error after a
  goreleaser version bump. Update `.goreleaser.yml` and cut a patch tag.
- **Registry rejects the release** → almost always a signature
  verification problem. Confirm the public key registered at the Registry
  matches the private key in `GPG_PRIVATE_KEY`. Both should resolve to
  the same fingerprint via `gpg --fingerprint`.
- **Tag points at the wrong commit** → delete locally and remotely with
  `git tag -d vX.Y.Z && git push origin :refs/tags/vX.Y.Z`, then retag.
  This is hostile if the release was already publicly fetched — prefer
  cutting the next patch version forward instead.

## Safety checklist before pushing the tag

- [ ] Working tree clean (`git status`)
- [ ] Origin/main up to date (`git log origin/main..HEAD` empty)
- [ ] `make test` green
- [ ] `docs/` regenerated and committed if any schema changed
- [ ] Version bump matches the change type (patch/minor/major)
- [ ] Annotated tag with a meaningful message (use `-a`, not just `git tag vX.Y.Z`)
