# TODO List

Short- and mid-term work for modloc. Status markers: `[ ]` open, `[~]` in
progress, `[x]` done. Done rows are deleted on sight: completed work lives
in `CHANGELOG.md`.

## Owner decisions

- [ ] **License.** `LICENSE` is the proprietary text the tooling writes for
      a new repository, while both remotes are public. Decide: MIT, like
      most public libraries of the fleet, or keep it proprietary.
- [ ] **Cut `v0.1.0`** (go-release skill): annotated tag, link references in
      `CHANGELOG.md`, then `go install github.com/larsartmann/modloc/cmd/modloc@latest`
      works.
- [ ] **SystemNix cutover.** Add the `modloc` flake input next to
      `nixos-modules`, lock it (`nix flake update modloc`, plus a `follows`
      line for every sub-input the lock audit rejects), switch
      `systems/evo-x2.nix` from `inputs.nixos-modules.nixosModules.modloc`
      to `inputs.modloc.nixosModules.modloc` (`programs.modloc.enable =
      true;` stays), deploy, and prove `modloc --version` on the host. The
      ordered steps are in the nixos-modules repository:
      `modules/modloc/README.md`, "Successor and cutover".
- [ ] **Retire the bash module** in nixos-modules once the cutover is live
      (same runbook, step 2 and 3). Retiring it first turns every
      consumer's evaluation red.

## Work

- [ ] **First CI run.** `.forgejo/workflows/ci.yml` is new: confirm that the
      next push to `origin` starts a run on the native runner and that the
      run is green.
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
