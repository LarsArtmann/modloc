# STACK — modloc

> The deliberate foundation of this project: every significant dependency was
> chosen, not accumulated. Written 2026-11-26 with the stack-selection skill
> (profile: Go CLI tool). This is a LIVING document: update the row when the
> reason changes, delete the row when the dependency goes.
> Enforced by `stack-selection/scripts/stack-drift-check.sh <project-dir>`:
> every direct `github.com/larsartmann/*` require must have a Chosen
> foundation row. Exit 0 = consistent, exit 1 = undeclared or contradicted
> dependency, exit 2 = this file is missing.

## Profile

| Axis         | Answer                                                                                           |
| ------------ | ------------------------------------------------------------------------------------------------ |
| Shape        | CLI: the root command counts, plus `doctor` and `version`                                        |
| Domain       | size of source code: lines of code per logical module                                            |
| Persistence  | none                                                                                             |
| Interfaces   | terminal: paths in; table, Markdown, TSV, CSV, JSON and JSONL out                                 |
| Scale target  | repositories of up to about a hundred thousand files, one `scc` scan each                       |
| Constraints  | single binary, nix build, public dependencies only, tests offline, `scc` on `PATH` at run time   |

## Chosen foundation

| Category            | Choice (first-party = ★)                                     | Why this, not that                                                                                                                                             | Owning skill  | Revisit trigger                              |
| ------------------- | ------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------- | -------------------------------------------- |
| CLI framework       | ★ cmdguard v5                                                | typed flags with allowed values, help, version, doctor and the describe contract. Rejected a hand-written parser (three argument-parsing bug classes shipped in the bash predecessor within two days) and bare fang (no typed flags) | cmdguard      | cmdguard offers a root-run API without cobra types |
| Output rendering    | ★ go-output (root, `delimited`, `markdown`, `table`)         | one table feeds five formats through registered renderers. Rejected hand-rolled padding and a direct `encoding/csv` writer                                     | go-output README | a format go-output cannot render          |
| Counting engine     | `scc` binary, run as a process                               | its language table decides what code, comment and blank lines are, and coverage follows the installed version. Rejected `scc` as a library: its module pulls json-iterator and zerolog (both banned by how-to-golang) and its package-global API is under change | how-to-golang | `scc` ships a stable `--by-module`            |
| Number formatting   | `github.com/dustin/go-humanize`                              | `humanize.Comma` and `english.Plural`. The fleet's `go-humanize-linter` flags hand-rolled copies of both                                                         | how-to-golang | a stdlib thousands separator                 |
| CLI types           | `github.com/spf13/cobra`                                     | only for the two lines of the root command that cmdguard does not wrap: its arguments validator and its flag-error hook. Rejected a root registered as a subcommand (it would change the command line) | cmdguard      | the same trigger as the cmdguard row         |
| Tests               | Ginkgo v2 + Gomega                                           | the how-to-golang policy: testify is banned; specs describe behavior                                                                                           | bdd-testing    |                                              |
| Nix packaging       | go-nix-helpers `flakeModules.go-standard`                    | the fleet pattern: one block for package, checks, devShells and formatter. Rejected a hand-rolled `buildGoModule`                                             | nix-review    |                                              |
| Quality pipeline    | BuildFlow                                                     | lint, format, tests and repairs in one loop. This repository ships only a comment-only `.buildflow.yml`                                                            | buildflow      |                                              |

Rules for this table:

- One row per concern the profile actually has.
- "Why this, not that" names the rejected alternative.
- First-party libraries are marked ★ and were verified to exist (README,
  `go.mod`, tagged version) before this table was written.

## Deliberately excluded

| Candidate                | Reason                                                                                                                      |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `samber/lo` + `samber/mo` | seven short loops would gain little; a new dependency with its own graph is not worth them (go-auto-upgrade suggests it) |
| `samber/oops`            | sentinels and `fmt.Errorf` with `%w` are the how-to-golang rule for errors at this size                                     |
| `cockroachdb/errors`     | nothing here needs stack traces or error classes                                                                             |
| `go-snaps`               | the goldens are five small TSV files and one JSON file; a plain byte comparison reads clearer                                |
| go-appkit                | the fleet default for an HTTP service or daemon; modloc is neither                                                          |

## Wiring order

1. **root**: `go get github.com/larsartmann/cmdguard/v5@v5.1.0` and
   `github.com/larsartmann/go-output@v0.38.7` with its `delimited`,
   `markdown` and `table` sub-modules.
2. `cmdguard` per the cmdguard skill: `NewCLI` with the root flags, the root
   command running the scan, the battery on.
3. `go-output`: the blank imports in `internal/render` register the renderers.
4. `go-nix-helpers` `go-standard` in `flake.nix`; public dependencies only,
   so no `nix-private-go-repos` machinery.
5. BuildFlow: `.buildflow.yml` marker, then `buildflow precommit install`.
5. Registered with `meta`: `meta set --project modloc --tags cli,go,tool,developer-tools,static-analysis`.

## Decision log

- 2026-11-26 — Initial selection (stack-selection skill), in the session
  that moved modloc out of the nixos-modules collection. The predecessor
  was a bash script with `jq`; the counting engine stayed `scc`, as a
  process.
- 2026-11-26 — `scc` as a library rejected on evidence read from its module:
  json-iterator and zerolog are direct dependencies of `github.com/boyter/scc/v4`.
- 2026-11-26 — `samber/lo` not adopted: go-auto-upgrade's eleven suggestions
  are map and reduce loops of three to five lines, two of them in test
  helpers.
