# DOMAIN_LANGUAGE.md — modloc's ubiquitous language

One glossary, one meaning per term. When code, docs and reports disagree,
this file wins and the code gets fixed. Sources: the types in `internal/loc`,
the formats in `internal/render`, the command line in `internal/cli`.

## The domain in one paragraph

A **scan** counts every **file** of one **repository**, attributes each file
of a **language** to its nearest **logical module**, and returns a **report**
of one **row** per module. A person reads the report as a **human format**,
a program as a **data format**.

## Terms

| Term               | Meaning                                                                                                                                                           |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Repository         | One directory tree given as a path; with no path, the current directory. The path is used exactly as given in the `repo` column                                   |
| Logical module     | The directory holding a language's **marker**. What modloc counts by: the unit a `go.mod` or `Cargo.toml` declares                                                |
| Marker             | The file that declares a module: `go.mod` for Go, `Cargo.toml` for Rust                                                                                           |
| Language           | Chooses both halves of attribution: the marker (`Marker`) and the file extension that makes a file count (`Extension`). `go` or `rust`                            |
| File               | One counted source file: its path relative to the repository and its lines, code, comment and blank counts, as `scc` counted them                                 |
| Attribution        | The rule that gives each file its module: the nearest marker directory above it. Files in a nested module never count toward the module that encloses them        |
| Root module (`.`)  | The repository root, when it holds a marker itself                                                                                                                |
| `OTHER`            | The row for files of the language that have no marker above them in a repository whose root is not a module. (`Unattributed` in the domain package)               |
| Pruned directories | `vendor`, `target`, `node_modules`, `.git`: never modules, so a marker inside them claims nothing. The repository root itself is never pruned                     |
| Counter            | Whatever counts files (`loc.FileCounter`). Production uses `scc` through `internal/scc`; tests substitute a canned one                                            |
| Report             | The scan result of one repository: its rows, ordered by code lines descending then by path, and the directories it skipped                                        |
| Row                | One module in a report: `repo`, `module`, `files`, `code`, `comment`, `blank`, `lines`                                                                            |
| Skipped            | Directories that could not be read while looking for markers; their subtrees may hide modules. Named on stderr, never fatal                                       |
| Notice             | A line on stderr that the table cannot say: an empty scan, a skipped directory. stdout stays pure data                                                            |
| Format             | What `--format` accepts: `auto`, `table`, `tsv`, `csv`, `json`, `jsonl`, `markdown`. `auto` is resolved before anything is written                                |
| Output             | A concrete format. `auto` resolves to `table` when stdout is a terminal and to `tsv` otherwise                                                                    |
| Human format       | `table` and `markdown`: numbers grouped by thousands, a `TOTAL` row at the end                                                                                    |
| Data format        | `tsv`, `csv`, `json`, `jsonl`: exact numbers, no summary row, no escape codes                                                                                     |
| TSV contract       | The exact shape of the TSV output, which the bash predecessor printed and tooling reads: one header, one row per module, nothing after the rows                   |
| Parity golden      | A TSV the bash `modloc` printed for a fixture, stored in `internal/cli/testdata/parity/`; the built binary must reproduce it byte for byte                        |
| Contract           | The machine-readable description of the command line (`modloc describe`): every flag, its default, its allowed values and every command. Golden-pinned            |
| Usage error        | A command line modloc cannot act on: a wrong value, an unknown flag, a missing value. Exits `2` before anything is scanned                                        |
| Stamp              | What a build knows about itself: the version and the commit `--version` prints                                                                                    |
| VACUOUS            | A spec that cannot decide in this sandbox (a directory the process can read anyway, no `scc` installed) and skips, naming the reason, instead of passing silently |

## Bound to one side

| Context               | Term that shifts meaning                                                                                  |
| --------------------- | --------------------------------------------------------------------------------------------------------- |
| count vs report       | `scc` counts files; modloc attributes them. A file's numbers are `scc`'s, a module's numbers are modloc's |
| terminal vs pipe      | The same command is a picture on a terminal and data in a pipe (`auto`), never a mixture                  |
| before vs after print | Every path is validated before the header is printed; a counter failure after that is exit `1`            |
