# ROADMAP

Long-term direction and raw ideas that are not refined into tasks. Actionable
short-term work lives in [TODO_LIST](TODO_LIST.md), the feature inventory in
[FEATURES](FEATURES.md).

**Last refreshed:** 2026-11-26 · Crush (glm-5.4) docs-health audit · basis:
the bash roadmap re-homed from nixos-modules, minus what has shipped or
become a task.

## Direction

1. **Stay boring.** One job (lines per logical module), one engine (`scc`),
   a byte-stable TSV. Every idea below must leave the TSV and the exit
   codes (`0`, `1`, `2`) exactly as they are.
2. **Follow nixpkgs.** What counts as code is `scc`'s decision, and the
   flake wraps the channel's `scc`: a new language arrives with the next
   `nix flake update`, without a modloc release.

## Ideas (not refined, not commitments)

- **`--sort code|lines|files|module` and `--top N`.** Today the order is
  fixed: code lines, largest first, then path.
- **`--exclude-dir`.** More directories that are never modules. The default
  list stays `vendor`, `target`, `node_modules`, `.git`.
- **`-l all`.** Every language in one pass: a file joins the nearest marker
  of any language.
- **More languages.** Python (`pyproject.toml`, `.py`), TypeScript and
  JavaScript (`package.json`, `.ts`, `.js`). Each needs a decision about
  which marker counts: `package.json` sits in far more directories than
  `go.mod`, so nearest-marker would split a repository differently.
- **A fleet sweep.** `modloc ~/projects/*` works today (several paths,
  one `TOTAL` row). A dedicated `--fleet` mode needs a real request first.
- **An upstream `scc --by-module`.** The attribution could live in `scc`
  itself. An issue-first pull request was parked while the consumer was a
  bash script; modloc no longer depends on it.
- **A project website** (website-launch skill): a landing page with the
  table and the formats. Not needed for the first release.

## Open questions

- **Does the bash module stay one release as a fallback** after the SystemNix
cutover, or is it retired in the same cycle? (nixos-modules,
  `modules/modloc/README.md`, "Successor and cutover".)
- **When does SystemNix switch:** now, or once `v0.1.0` is tagged?
