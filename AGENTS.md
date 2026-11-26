# modloc AGENTS.md

Go CLI (cobra/cmdguard v5 + go-output) built with BuildFlow. Check `go.mod`
for the current dependency set and `cmd/` + `internal/` for structure.

## Remotes (since 2026-11-26)

Forgejo-canonical (fleet pattern): `origin` =
`forgejo@evo-x2.home.lan:lars/modloc.git`, `github` =
`git@github.com:LarsArtmann/modloc.git` is a PUBLIC Forgejo push mirror (8h +
sync_on_commit, hand-attached at creation; owned idempotently by
`forgejo.canonicalRepos` in SystemNix from the next deploy). Both repos
PUBLIC. Push to `origin` first; the mirror follows.
