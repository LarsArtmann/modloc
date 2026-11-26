# modloc AGENTS.md

modloc counts lines of code per **logical module** (the nearest `go.mod` or
`Cargo.toml`), not per repository. Go CLI on cmdguard v5 and go-output,
built and checked with BuildFlow. `go.mod` has the dependency set and
`STACK.md` the reason for each; `FEATURES.md` the inventory.

## Remotes (since 2026-11-26)

Forgejo-canonical (fleet pattern): `origin` =
`forgejo@evo-x2.home.lan:lars/modloc.git`, `github` =
`git@github.com:LarsArtmann/modloc.git` is a PUBLIC Forgejo push mirror (8h +
sync_on_commit, hand-attached at creation; owned idempotently by
`forgejo.canonicalRepos` in SystemNix from the next deploy). Both repos
PUBLIC. Push to `origin` first; the mirror follows.

CI therefore lives in `.forgejo/workflows/ci.yml` (runner label `native`),
never in `.github/workflows/`: the token of the mirror has no `workflow`
scope (SystemNix, `platforms/nixos/system/configuration.nix`, the entry for
`modloc`), so a commit that adds a workflow file there is refused by GitHub
and the mirror stalls.

## Commands

```bash
go build ./...    # build
go test ./...     # behavior specs (Ginkgo); specs needing scc skip without it
buildflow         # the quality gate: lint, format, tests, repairs
nix build         # hermetic build; the package wraps scc onto its PATH
nix flake check   # package, specs with the real scc, both Nix modules, format
nix run . -- -l rust ~/projects/collector-utils
```

Nix sees only git-tracked files (`src = lib.fileset.gitTracked ./.`):
`git add` a new file before `nix build` or `nix flake check`, or the flake
silently builds without it.

## Architecture map

| Package           | Owns                                                                                       |
| ----------------- | ------------------------------------------------------------------------------------------ |
| `cmd/modloc`      | the entry point and the two stamps the linker sets (`version`, `commit`)                   |
| `internal/loc`    | the domain: what a logical module is, how files are attributed, `Scan`. Depends on nothing |
| `internal/scc`    | runs the `scc` binary and reads its JSON. The only place that knows `scc`                  |
| `internal/render` | the seven output formats. The only place that knows go-output                              |
| `internal/cli`    | cmdguard wiring, path validation, notices, exit codes. Composes the three above            |

A change to how files are attributed belongs in `internal/loc`, a change to a
flag or an exit code in `internal/cli`, a change to a format in
`internal/render`. `internal/cli` has no business logic of its own.

## Invariants (deliberate, do not "fix")

- **The TSV contract.** `repo\tmodule\tfiles\tcode\tcomment\tblank\tlines`,
  one header for every repository, one row per module, nothing after the
  rows. `internal/cli/testdata/parity/` holds what the bash `modloc` printed
  for five fixtures; the built binary must reproduce it byte for byte. A new
  feature may add formats; it never changes TSV or the exit codes (`0` ok,
  `1` a path or the count failed, `2` the command line is wrong).
- **`scc` is a process, never a library.** Its module pulls json-iterator and
  zerolog (both banned by how-to-golang) and its package-global API is under
  change. Running the binary keeps language coverage on the nixpkgs channel.
- **Every path is checked before anything is printed**, and no path means the
  current directory: a bare `modloc` is the common invocation, and the first
  live run of the bash predecessor printed a bare header and exited 0.
- **stdout is the report and nothing else.** Notices go to stderr: an empty
  scan (with the command that finds the other language) and every directory
  that was skipped.
- **Root-module strays are not `OTHER`.** A file with no marker above it is
  `OTHER` only when the repository root is not a module; otherwise it is the
  root module `.`.
- **The unreadable-directory notice is not byte-equal to the bash one.** Go
  names each skipped directory ("skipped 1 unreadable directory under 'x':
  secret"); the bash line said only that some were skipped.
- **Markdown numbers are aligned by markers** (`---:`), table numbers by
  padding: go-output's terminal table exposes no alignment.

## Tool findings that are deliberate non-fixes

BuildFlow's rule is fixed, filed, or recorded; these are recorded.

- **erraudit (45 warnings: "project enforces samber/oops", "returns bare
  error interface").** modloc follows the how-to-golang rule for errors at
  this size: package-level sentinels and `fmt.Errorf` with `%w`. `oops`
  would be a dependency for error classes nothing here needs.
- **go-auto-upgrade (11 suggestions: `samber/lo`).** The loops are three to
  five lines each, two sit in test helpers; a new dependency with its own
  graph is not worth them (see `STACK.md`).
- **library-policy (1: cobra, "USE charm.land/fang/v2 with cobra").** Every
  help and error screen is fang's: cmdguard renders through fang by
  default. The direct cobra import exists for the two lines of the root
  command cmdguard does not wrap.
- **govulncheck (1: GO-2026-6859).** A `net/http` server race in the Go
  1.27.3 standard library, fixed in 1.27.4. modloc starts no HTTP server;
  the only trace govulncheck prints runs from `exec.Cmd.Start`. Goes away
  when nixpkgs ships 1.27.4.
- **golangci-lint-auto-configure (1: "gohumanize is recommended, not in
  golangci-lint v2.15.0").** BuildFlow runs it as its own step,
  `go-humanize-linter`; it passes.

## Gotchas

- **cmdguard's middleware chain does not reach a root that runs itself.**
  `WithMiddleware` applies to commands registered through cmdguard;
  `RootCommand().RunE` bypasses it. `wireRoot` applies the recovery
  middleware by hand, or a panic in the scan crashes the process.
- **cmdguard sends unknown-flag and missing-value errors out without a
  sentinel**, and cobra prints them as a plain failure (exit 1). The root's
  `SetFlagErrorFunc` tags them, so they exit 2 like every other usage error.
- **cmdguard's help lists allowed values only for fields typed as an
  `Enum`.** A `string` flag with a `values` tag is validated but its values
  are not shown, so every help text repeats them, and `describe_test.go`
  pins the `values` tags, the help texts and the domain lists together.
- **fang capitalizes the first letter of an error** ("Scc not found") and
  abbreviates the commit to seven characters. Specs that match error text
  do so case-insensitively.
- **cmdguard's `testutil.AssertDescribeContract` is not in the published
  v5.1.0 module** (it is in the repository, after the tag). The contract
  golden in `internal/cli/testdata/describe.golden.json` is compared by
  hand; `MODLOC_UPDATE_GOLDEN=1 go test ./internal/cli` rewrites it.
- **go-output's root module registers no format.** The blank imports in
  `internal/render` are load-bearing. Its Markdown separator row is one
  character narrower than its columns; valid Markdown, cosmetic.
- **`encoding/json/v2` has no `json.NewDecoder`.** Decoding from a reader is
  `json.UnmarshalDecode(jsontext.NewDecoder(r), &v)`.
- **The auto-commit daemon** commits whatever the working tree holds, and
  its pre-commit hook reformats files (golines) between a read and a write.
  Edit with the edit tool on a fresh read, never with a search-and-replace
  over a file read earlier. Throwaway files belong in `/tmp`, never in this
  repository.
- **`buildflow --fix` also runs the dependency update** ("Updating
  module"). Take `buildflow format` for formatting, `buildflow` for the
  verdict, and `-s <tool> --fix` for one tool.
- **BuildFlow's format step does not run golines or gofumpt**; the lint
  step reports both, and its repair step applies them.
