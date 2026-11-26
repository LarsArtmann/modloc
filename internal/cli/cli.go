// Package cli wires the modloc command line on cmdguard v5: typed flags,
// help, version, doctor and the describe contract. modloc is one command, so
// the root itself runs the scan. Everything it needs from its Environment is
// injected, which keeps the whole command testable against buffers.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	v5 "github.com/larsartmann/cmdguard/v5/pkg/cmdguard/v5"
	"github.com/spf13/cobra"

	"github.com/larsartmann/modloc/internal/loc"
	"github.com/larsartmann/modloc/internal/scc"
)

// Name is the binary and command name.
const Name = "modloc"

// exitCodeUsage marks a bad invocation; every other failure exits 1, so a
// script can tell "called it wrong" from "the scan went wrong".
const exitCodeUsage = 2

var (
	// errNoConfig reports a handler that ran without cmdguard's flag parsing.
	errNoConfig = errors.New("flags were not parsed into the command")
	// errFlagUsage marks an error from the command line itself: an unknown
	// flag, a missing value, bad syntax.
	errFlagUsage = errors.New("invalid usage")
)

// Config carries the root command's flags. Each help names its allowed
// values: cmdguard lists them for Enum-typed flags only, and a reader of
// --help should never have to guess.
//

type Config struct {
	Language string `default:"go"   flag:"language" help:"Language to count (go, rust): its files, and the modules its marker declares (go.mod, Cargo.toml)"            short:"l" values:"go,rust"`
	Format   string `default:"auto" flag:"format"   help:"Output format (auto, table, tsv, csv, json, jsonl, markdown); auto is table on a terminal and tsv when piped" short:"f" values:"auto,table,tsv,csv,json,jsonl,markdown"`
	Color    string `default:"auto" flag:"color"    help:"Colors of the table output (auto, always, never)"                                                                       values:"auto,always,never"`
}

// Probe reports which scc will run: its path and its version line.
type Probe func(ctx context.Context) (path string, version string, err error)

// Environment is what the commands need from their environment: the streams
// they write to, whether stdout is a terminal, the counter that knows line
// counts, the probe that names the counter's binary, and the version stamps.
// Production wires the process itself; tests substitute buffers.
type Environment struct {
	// Stdout receives the report.
	Stdout io.Writer
	// Stderr receives notices and failed writes.
	Stderr io.Writer
	// StdoutTTY says whether Stdout is a terminal; it resolves the auto
	// format.
	StdoutTTY bool
	// Counter counts the files of a repository.
	Counter loc.FileCounter
	// Scc names the binary that Counter runs.
	Scc Probe
	// Version and Commit are the build stamps shown by --version.
	Version string
	Commit  string
}

// Main runs modloc against the process's streams and returns its exit code:
// 0 on success, 1 for a failed scan or path, 2 for a usage error.
func Main(ctx context.Context, args []string, version, commit string) int {
	return v5.ExitCode(Run(ctx, args, version, commit))
}

// Run builds the CLI against the process's streams and scc and executes args.
// A construction error is printed here; the execution error is printed by
// cmdguard exactly once and returned unprinted.
func Run(ctx context.Context, args []string, version, commit string) error {
	env := Environment{
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		StdoutTTY: isTerminal(os.Stdout),
		Counter:   scc.New(),
		Scc:       scc.New().Probe,
		Version:   version,
		Commit:    commit,
	}

	cli, err := Build(env) //nolint:contextcheck // the handler must use cobra's signal-cancelled context
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		return err
	}

	return Execute(cli, ctx, args)
}

// Build assembles the command tree: the scanning root and the doctor command.
func Build(env Environment) (*v5.CLI[Config], error) {
	cli, err := v5.NewCLI(Name, long, Config{}, options(env.Version, env.Commit)...)
	if err != nil {
		return nil, err
	}

	wireRoot(cli, env)

	doctorCmd, doctorErr := v5.NewCommand(
		"doctor",
		v5.NoFlags{},
		func(ctx context.Context, _ *Config, _ v5.NoFlags) error { return runDoctor(ctx, env) },
		v5.WithShort("Check that modloc can count: locate scc and print its version"),
	)

	b := v5.NewBuilder(cli)
	v5.Try(b, doctorCmd, doctorErr)

	if err := b.Build(); err != nil {
		return nil, err
	}

	return cli, nil
}

// long is the root command's description in --help.
const long = "Count lines of code per logical module instead of per repository.\n" +
	"\n" +
	"A logical module is the nearest directory holding go.mod (Go) or Cargo.toml\n" +
	"(Rust). Every file counts toward the closest module above it, so nested\n" +
	"modules never add to the module that encloses them. With no <repo>, the\n" +
	"current directory is scanned."

// examples are the root command's examples in --help.
const examples = "  # Scan the current directory\n" +
	"  modloc\n\n" +
	"  # Scan other repositories\n" +
	"  modloc ~/projects/some-monorepo\n\n" +
	"  # Rust crates, as a Markdown table\n" +
	"  modloc -l rust --format markdown ~/projects/collector-utils"

// wireRoot makes the root command the scan itself.
func wireRoot(cli *v5.CLI[Config], env Environment) {
	root := cli.RootCommand()
	root.Use = Name + " [flags] [repo...]"
	root.Example = examples
	// The root has subcommands (doctor, version, ...), so cobra would treat a
	// bare path as an unknown command without an explicit args validator.
	root.Args = cobra.ArbitraryArgs
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errFlagUsage, err)
	})
	root.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, ok := v5.ConfigFromContext[Config](cmd.Context())
		if !ok {
			return errNoConfig
		}

		return scan(cmd.Context(), cfg, args, env)
	}
}

// options are the root options: version stamps and panic recovery, so a bug
// reports an error instead of crashing the process.
func options(version, commit string) []v5.CLIOption {
	opts := []v5.CLIOption{
		v5.WithCLIVersion(version),
		v5.WithMiddleware(v5.RecoveryMiddleware[Config]()),
	}

	if commit != "" {
		opts = append(opts, v5.WithCLICommit(commit))
	}

	return opts
}

// Execute runs the CLI through cmdguard's exit-code contract: a malformed
// command line is re-tagged as a usage error, so the process exits 2.
func Execute(cli *v5.CLI[Config], ctx context.Context, args []string) error {
	err := cli.ExecuteWithArgs(ctx, args)
	if isUsageError(err) {
		return usageError(err)
	}

	return err
}

// isUsageError reports whether err means the command was called wrong: a
// malformed command line, or cmdguard's binding layer rejecting a flag value
// (a bad enum, a number out of bounds). Anything else is a real failure.
func isUsageError(err error) bool {
	if errors.Is(err, errFlagUsage) {
		return true
	}

	for _, sentinel := range []error{
		v5.ErrFlagParseFailed,
		v5.ErrInvalidEnum,
		v5.ErrTypeConversion,
		v5.ErrValueOutOfRange,
		v5.ErrValueTooShort,
		v5.ErrValueTooLong,
		v5.ErrValueTooSmall,
		v5.ErrValueTooLarge,
		v5.ErrValuePatternMismatch,
		v5.ErrValueEmpty,
	} {
		if errors.Is(err, sentinel) {
			return true
		}
	}

	return false
}

// usageError wraps err so the process exits 2 instead of the default 1.
func usageError(err error) error {
	exitErr, invalid := v5.NewExitError(exitCodeUsage, err)
	if invalid != nil {
		return err
	}

	return exitErr
}

// isTerminal reports whether the file is attached to a terminal.
func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
