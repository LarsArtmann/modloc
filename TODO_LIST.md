# TODO List

Short- and mid-term work for modloc. Status markers: `[ ]` open, `[~]` in
progress, `[x]` done. Done rows are deleted on sight: completed work lives
in `CHANGELOG.md`. A row whose fix lands in another repository carries
`(repo: <name>)`.

**Created:** 2026-11-27 · Crush (glm-5.4) docs-health audit · backlog basis:
the two modloc status reports in nixos-modules (`docs/status/`, 06:05 and
17:53), each item verified against the tree before it was kept.

## Owner decisions

- [ ] **License.** `LICENSE` is the proprietary text the tooling writes for
      a new repository, while both remotes are public. Decide: MIT, like
      most public libraries of the fleet, or keep it proprietary.
- [ ] **Cut `v0.1.0`** (go-release skill): annotated tag, link references in
      `CHANGELOG.md`, then `go install github.com/larsartmann/modloc/cmd/modloc@latest`
      works.
- [ ] **SystemNix cutover** (repo: SystemNix). `systems/evo-x2.nix` still
      imports `inputs.nixos-modules.nixosModules.modloc`. Add the `modloc`
      flake input, lock it (`nix flake update modloc`, plus a `follows`
      line for every sub-input the lock audit rejects), switch to
      `inputs.modloc.nixosModules.modloc` (`programs.modloc.enable = true;`
      stays), deploy, and prove `modloc --version` on the host. The ordered
      steps are in the nixos-modules repository:
      `modules/modloc/README.md`, "Successor and cutover".
- [ ] **Retire the bash module** (repo: nixos-modules) once the cutover is
      live (same runbook, steps 2 and 3). Retiring it first turns every
      consumer's evaluation red.
- [ ] **Dependabot on the mirror.** `.github/dependabot.yml` opens its
      update pull requests on GitHub, which is a push mirror of Forgejo: a
      merge there is overwritten by the next sync. Delete the file, or
      accept that it only produces pull requests nobody can land.

## Work

Ordered by impact; every row was open in the tree on 2026-11-27.

- [ ] **First CI run.** `.forgejo/workflows/ci.yml` is hermetic Nix and its
      commits are on `origin`; no run has been seen. Confirm on Forgejo that
      the runs on the native runner pass.
- [ ] **Bite-proof harness, one stale case.** `tests/negative/run.sh` has 14
      cases; `cli/paths_checked_one_by_one` reports ANCHOR on 2026-11-27
      (run with `KEEP=1` to read its log), and the two `SLOW=1` cases
      (`package_without_the_scc_wrapper`, `package_without_its_version_stamp`)
      have not been run.
- [ ] **`rm -rf` in the harness.** `tests/negative/run.sh` removes its
      scratch copy with `rm -rf`; the house rule is `trash`.
- [ ] **The unreadable-subtree leg through the built binary.**
      `tests/package-runtime.sh` has a leg for a directory the user cannot
      read that reports VACUOUS when the sandbox cannot provide one;
      confirm whether it runs or skips inside `nix build`.
- [ ] **Home Manager module against the real Home Manager.**
      `tests/modules.nix` evaluates `homeManagerModules.modloc` against a
      stub that declares `home.packages` only; Home Manager is not a flake
      input.
- [ ] **Review `.golangci.yml` setting by setting.** The file is inherited
      from google-domain-detector (funlen, gosec excludes, `varnamelen`,
      `tagalign`); lint passes, the fit for modloc was never checked.
- [ ] **`internal/cli` coverage gap.** 88.7 %; the uncovered lines are the
      construction-error branch of `Run` (`internal/cli/cli.go`). Cover it
      or record why not.
- [ ] **Coverage floor in the verdict.** `scripts/coverage-floor.sh` runs
      in CI only; wire it into BuildFlow once it can run repo scripts, so
      the local verdict carries it too.
- [ ] **GO-2026-6859.** `govulncheck` reports it against the Go 1.27.3
      standard library (`net/http` server code modloc never runs); Go 1.27.4
      fixes it. Re-run `buildflow -s govulncheck` once nixpkgs ships 1.27.4
      and drop the note from `AGENTS.md`.
- [ ] **README sample.** Once `v0.1.0` exists, show a versioned install
      (`nix run github:LarsArtmann/modloc/v0.1.0`); today the example
      tracks `master`.
- [ ] **The harness in CI.** `tests/negative/run.sh` runs locally only;
      add its fast cases to `.forgejo/workflows/ci.yml` once the stale case
      is fixed and the run time is known.
- [ ] **A status legend in `FEATURES.md`.** `docs-gates` skips its
      vocabulary gate for FEATURES.md ("no status legend section, not
      judged"); a legend section lets the gate check the status words.
- [ ] **Annotate the reports this work resolves** (repo: nixos-modules):
      `docs/status/2026-11-26_17-53_modloc-*.md`, section f, rows 1, 2, 3,
      25 and the modloc half of 31.

## File upstream

Workarounds live in `AGENTS.md`; the causes belong in the libraries. Each
row was verified in modloc's tree; none is in the other repository's
`TODO_LIST.md`. Run the verify-before-filing and github-voice skills first.

- [ ] (repo: cmdguard) A root command that runs itself bypasses the
      middleware chain: `WithMiddleware` does not reach `RootCommand().RunE`.
- [ ] (repo: cmdguard) Help hides the allowed values of a `string` flag with
      a `values` tag, while the enum error's hint points at that help.
- [ ] (repo: cmdguard) Unknown-flag and missing-value errors leave
      `ExecuteWithArgs` without a sentinel, so consumers exit 1 for a usage
      error unless they install a `FlagErrorFunc`.
- [ ] (repo: cmdguard) `testutil.AssertDescribeContract` is not in the
      published v5.1.0 module, while the cmdguard skill tells consumers it
      is.
- [ ] (repo: BuildFlow) The result cache replays a finding whose cause is
      fixed (`securitymd`), and `buildflow --fix` runs the dependency update.
- [ ] (repo: go-output) The Markdown table separator row is one character
      narrower than its columns.
- [ ] (repo: crush-config) Add the lessons to `references/lessons.md`: the
      auto-commit daemon's hook reformats between read and write; a spec
      that depends on an external tool's order is flaky; a root command
      outside the middleware chain.
- [ ] (repo: docs-gates) A gate for a paragraph that appears twice in a
      living doc: a multi-edit session duplicated the CI paragraph of
      `AGENTS.md` on 2026-11-27 and every gate stayed green.
