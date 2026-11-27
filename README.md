# modloc

Lines of code per **logical module**, not per repository.

`scc`, `tokei` and `cloc` answer "how big is this repository?". A monorepo
holds many modules (nested `go.mod` files, Cargo workspace crates), and there
the question is "how much code lives in each module?". `modloc` answers that:
every file counts toward the nearest module marker above it, so the lines of
a nested module never count toward the module that encloses it.

```console
$ cd ~/projects/collector-utils && modloc -l rust
╭───────┬───────────┬───────┬───────┬─────────┬───────┬────────╮
│ repo  │ module    │ files │ code  │ comment │ blank │ lines  │
├───────┼───────────┼───────┼───────┼─────────┼───────┼────────┤
│ .     │ .         │    22 │ 6,841 │   3,010 │   904 │ 10,755 │
│ .     │ fuzz      │     7 │   282 │      77 │    44 │    403 │
│ TOTAL │ 2 modules │    29 │ 7,123 │   3,087 │   948 │ 11,158 │
╰───────┴───────────┴───────┴───────┴─────────┴───────┴────────╯
```

The root module `.` holds the 22 Rust files outside every crate; `fuzz` is a
nested module, so its 7 files count toward `fuzz` alone.

## Quick start

Nothing to install:

```console
$ nix run github:LarsArtmann/modloc -- -l rust ~/projects/collector-utils
```

On a NixOS system, as a system-wide tool:

```nix
{
  inputs.modloc.url = "github:LarsArtmann/modloc";
}
```

```nix
{
  imports = [ inputs.modloc.nixosModules.modloc ];
  programs.modloc.enable = true;
}
```

In a Home Manager profile (any Linux, or one user among several):

```nix
{
  imports = [ inputs.modloc.homeManagerModules.modloc ];
  programs.modloc.enable = true;
}
```

Without Nix: `go install github.com/larsartmann/modloc/cmd/modloc@master`,
plus [scc](https://github.com/boyter/scc) on `PATH`
(`go install github.com/boyter/scc/v4@latest`). `modloc doctor` says whether
it can count.

## Usage

```console
$ modloc                                  # the current directory
$ modloc ~/projects/some-monorepo         # explicit paths, options anywhere
$ modloc -l rust ~/projects/collector-utils
$ modloc -f json ~/projects/a ~/projects/b > modules.json
$ modloc --help
```

| Flag               | Values                                                     | Default | Meaning                                   |
| ------------------ | ---------------------------------------------------------- | ------- | ----------------------------------------- |
| `-l`, `--language` | `go`, `rust`                                               | `go`    | Which files to count, and by which marker |
| `-f`, `--format`   | `auto`, `table`, `tsv`, `csv`, `json`, `jsonl`, `markdown` | `auto`  | Output format                             |
| `--color`          | `auto`, `always`, `never`                                  | `auto`  | Colors of the table and Markdown output   |

`auto` picks `table` when stdout is a terminal and `tsv` when it is piped or
redirected, so `modloc > file` gives data and `modloc` gives a picture.

`modloc doctor` names the `scc` binary that will run. `modloc --version`
prints the version and the commit it was built from, `modloc completion
bash` (or `zsh`, `fish`) prints the shell completion script.

| Command      | Does                                              |
| ------------ | ------------------------------------------------- |
| `modloc`     | count the paths given, or the current directory   |
| `doctor`     | name the `scc` that will run, or say why none can |
| `version`    | the same as `--version`                           |
| `completion` | print the completion script for bash, zsh or fish |

| Exit code | Meaning                                     |
| --------- | ------------------------------------------- |
| `0`       | Counted                                     |
| `1`       | A path is no directory, or the count failed |
| `2`       | The command line is wrong                   |

Every path is checked before anything is printed. A scan that matched no
file still prints the header and says so on stderr; a directory that cannot
be read drops only its own subtree and is named on stderr. stdout carries
the report and nothing else.

## Output formats

| Format     | For          | Numbers          | Ends with              |
| ---------- | ------------ | ---------------- | ---------------------- |
| `table`    | the terminal | grouped, aligned | a `TOTAL` row          |
| `markdown` | documents    | grouped, aligned | a `TOTAL` row          |
| `tsv`      | scripts      | exact            | nothing after the rows |
| `csv`      | spreadsheets | exact            | nothing after the rows |
| `json`     | tooling      | exact, numeric   | nothing after the rows |
| `jsonl`    | streaming    | exact, numeric   | nothing after the rows |

```json
[
  {
    "repo": ".",
    "module": "fuzz",
    "files": 7,
    "code": 282,
    "comment": 77,
    "blank": 44,
    "lines": 403
  }
]
```

## How it counts

1. One `scc --by-file --format json` scan counts every file of the
   repository: `scc`'s language table decides what a code, a comment and a
   blank line is.
2. A **logical module** is a directory holding the language's marker:
   `go.mod` for Go, `Cargo.toml` for Rust.
3. Every file of the language joins the nearest marker directory above it.
   `vendor`, `target`, `node_modules` and `.git` are never modules.
4. A file with no marker above it joins the root module (`.`) when the
   repository root is a module itself, and the row `OTHER` when it is not.
5. Rows are ordered by code lines, largest first; modules of equal size are
   ordered by path.

Columns: `repo`, `module`, `files`, `code`, `comment`, `blank`, `lines`.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md). In short: `nix build`, `nix flake
check`, `buildflow`.

## License

See [LICENSE](LICENSE).
