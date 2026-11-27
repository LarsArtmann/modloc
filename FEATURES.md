# Features

Honest inventory of what modloc does, by status. The source of truth is the
code; every row says where it lives and what proves it. There is no
BROKEN, PARTIALLY_FUNCTIONAL or DISABLED row: every shipped feature passed
its specs and its Nix checks when this list was refreshed.

**Last refreshed:** 2026-11-27 · Crush (glm-5.4) docs-health audit · basis:
`go test ./...` with coverage, `nix flake check`, and the built binary run in
every format.

Test count: 157 specs (Ginkgo, black-box), covering 94.4 % of the
statements. `bash scripts/test-count.sh` checks the first number.

## FULLY_FUNCTIONAL

### Counting

- Lines of code per logical module: every file of the language joins the
  nearest `go.mod` (Go) or `Cargo.toml` (Rust) above it; files outside every
  module land on `OTHER`, or on the root module `.` when the root is a
  module; `vendor`, `target`, `node_modules` and `.git` are never modules;
  rows by code lines, then by path. `internal/loc` (42 specs, 95.9 %).
- `scc` as the counting engine, run once per repository as a process:
  `scc --by-file --format json .`. A missing, failing or silent `scc` is a
  named error; a cancelled scan stops at once. `internal/scc` (15 specs,
  95.2 %, one golden captured from the real `scc` 4.1.0).
- Byte parity with the bash `modloc` it replaces: five TSV goldens (Go,
  Rust, `OTHER`, equal-sized modules, a path that starts with a dash) that
  the built binary must reproduce. `internal/cli/testdata/parity/`.

### Command line

- A bare `modloc` scans the current directory; paths go anywhere on the
  line, `--` ends the options, and a path with spaces or a leading dash is
  a path.
- Typed flags with allowed values: `-l/--language`, `-f/--format`,
  `--color`. A wrong value, an unknown flag or a missing value exits `2`
  before anything is scanned.
- Every path is checked before anything is printed: a missing path or a file
  exits `1` and is named.
- `--help` on stdout; `--version` and the `version` command with the commit;
  `doctor` naming the `scc` that will run; `completion` for bash, zsh and
  fish; `--no-color`; the hidden machine-readable `describe` contract,
  golden-pinned.
- Notices on stderr, never on stdout: an empty scan (with the command that
  finds the other language) and unreadable directories.
- A crash inside the counter is reported as a failure, not a stack trace.
- `internal/cli` (64 specs, 88.7 %), including specs that build the binary
  and run it as a process with the real `scc`.

### Output

- `auto`: a bordered table on a terminal, TSV when piped or redirected.
- `table` and `markdown`: numbers grouped by thousands, numeric columns
  right-aligned, a `TOTAL` row.
- `tsv`, `csv`, `json`, `jsonl`: exact numbers, no summary row, no escape
  codes; JSON counts are numbers.
- `internal/render` (36 specs, 100 %).

### Packaging

- A Nix flake on `go-nix-helpers.flakeModules.go-standard`: the package
  wraps `scc` onto the binary's `PATH` and installs shell completions.
- `nixosModules.modloc` and `homeManagerModules.modloc`, both
  `programs.modloc.enable` with `package` defaulting to this flake's build.
- Checks: the package build runs the specs with the real `scc`;
  `package-runtime` runs the wrapped binary with an empty environment and
  asserts its three completion files (`modloc.bash`, `_modloc`,
  `modloc.fish`); the module checks evaluate both modules enabled and
  disabled (`tests/modules.nix`): the package is on the list exactly once
  when enabled and absent when disabled.

## DEMAND_GATED

- More languages (Python, TypeScript): the attribution is language-agnostic;
  a language needs a marker, an extension and two spec fixtures.
- `-l all` in one pass, `--sort`, `--top N`, `--exclude-dir`, a fleet sweep:
  ideas in [ROADMAP](ROADMAP.md), none requested yet.
