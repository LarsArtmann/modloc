# Changelog

All notable changes to modloc are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and versions follow
Semantic Versioning.

## [Unreleased]

### Added

- Count lines of code per logical module: every file of the language joins
  the nearest `go.mod` (Go) or `Cargo.toml` (Rust) above it, and a file
  outside every module lands on `OTHER`, or on the root module `.` when the
  root is a module itself. `scc` stays the counting engine, run as a
  process, so language coverage follows the installed `scc`. The TSV output
  is byte-identical to the bash `modloc` of the nixos-modules collection on
  five golden fixtures.
- The command line on cmdguard v5: typed flags `-l/--language` and
  `-f/--format`, a help that names every allowed value, `--version`,
  `doctor`, shell completion and the hidden `describe` contract. Exit
  codes: `0`, `1` (a path or the count failed) and `2` (the command line is
  wrong).
- Output formats: `auto` (the default: `table` on a terminal, `tsv` when
  piped), `table`, `tsv`, `csv`, `json`, `jsonl` and `markdown`, plus
  `--color`.
- A Nix flake: the package wraps `scc` onto its `PATH`;
  `nixosModules.modloc` and `homeManagerModules.modloc`
  (`programs.modloc.enable`); eval and runtime checks.
- 157 behavior specs (Ginkgo) over 94.4 % of the statements.
- Notices on stderr, never on stdout: an empty scan names the command that
  finds the other language, and every directory that could not be read is
  named while the rest of the repository is still counted.
- `vendor`, `target`, `node_modules` and `.git` are never modules, whatever
  marker they hold; the repository root is never pruned.
- A CI workflow on Forgejo Actions: the BuildFlow gate, a secret scan, the
  `STACK.md` drift check, the spec-count gate and a coverage floor of 80 %.
- `modloc completion` for bash, zsh and fish, installed with the Nix
  package.
