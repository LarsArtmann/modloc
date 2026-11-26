# Contributing

Thanks for your interest in contributing!

## Development setup

```console
$ nix develop
```

The shell carries the Go toolchain, `golangci-lint`, `govulncheck` and
`gopls`.

## Quality gate

```console
$ go build ./... && go vet ./...
$ go test -race ./...
$ buildflow            # lint, format, tests, repairs: the quality gate
$ nix flake check      # the hermetic build, the specs, both Nix modules
```

BuildFlow owns linting, formatting and `go.mod` hygiene. Do not run a
formatter or `go mod tidy` by hand: `buildflow format` and
`buildflow -s gomod-check --fix` are the two commands.

## Tests

Behavior specs with Ginkgo and Gomega, black-box (`package <name>_test`):
one spec per observable behavior. A spec whose outcome depends on the
sandbox (a directory it cannot read, `scc` being installed) reports VACUOUS
and skips instead of passing silently.

`internal/cli/testdata/` holds the goldens: the command contract
(`describe.golden.json`) and the TSV the bash `modloc` printed for five
fixtures (`parity/`). Change a golden only together with a deliberate
change of the contract.

## Docs

`README.md`, `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md`, `CHANGELOG.md`,
`AGENTS.md`, `STACK.md` and `docs/DOMAIN_LANGUAGE.md` are living documents.
Each fact lives in one of them; when a fact changes, update the document
that owns it. `bash scripts/test-count.sh` checks that the spec count in
`FEATURES.md` is the real one, and
`stack-selection/scripts/stack-drift-check.sh .` checks that every direct
first-party dependency has a row in `STACK.md`.

## Reporting issues

Please use the issue tracker. Security problems go the way
[SECURITY.md](SECURITY.md) describes, never through a public issue.
