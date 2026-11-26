package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/larsartmann/go-output"

	"github.com/larsartmann/modloc/internal/loc"
	"github.com/larsartmann/modloc/internal/render"
)

var (
	// errNoSuchDirectory names a path that does not exist.
	errNoSuchDirectory = errors.New("no such directory")
	// errNotDirectory names a path that is not a directory.
	errNotDirectory = errors.New("not a directory")
)

// scanOptions are the parsed flags of one scan.
type scanOptions struct {
	language loc.Language
	output   render.Output
	color    output.ColorMode
}

// scan validates its options and every repository, counts each repository
// and writes the reports. Nothing reaches stdout before every path has been
// checked, so a bad path never leaves a half table behind.
func scan(ctx context.Context, cfg *Config, args []string, env Environment) error {
	options, err := parseOptions(cfg, env.StdoutTTY)
	if err != nil {
		return usageError(err)
	}

	repos := repositories(args)

	err = validateRepositories(repos)
	if err != nil {
		return err
	}

	reports := make([]loc.Report, 0, len(repos))

	for _, repo := range repos {
		report, scanErr := loc.Scan(ctx, repo, options.language, env.Counter)
		if scanErr != nil {
			return fmt.Errorf("scan: %w", scanErr)
		}

		notice(env.Stderr, report, options.language)
		reports = append(reports, report)
	}

	err = render.Write(env.Stdout, reports, options.output, options.color)
	if err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	return nil
}

// parseOptions turns the flag strings into validated values and resolves the
// auto format for the destination. cmdguard has already checked the enums,
// so an error here is a flag that bypassed it.
func parseOptions(cfg *Config, stdoutTTY bool) (scanOptions, error) {
	language, err := loc.Parse(cfg.Language)
	if err != nil {
		return scanOptions{}, fmt.Errorf("--language: %w", err)
	}

	format, err := render.Parse(cfg.Format)
	if err != nil {
		return scanOptions{}, fmt.Errorf("--format: %w", err)
	}

	color, err := output.ParseColorMode(cfg.Color)
	if err != nil {
		return scanOptions{}, fmt.Errorf("--color: %w", err)
	}

	return scanOptions{language: language, output: format.Resolve(stdoutTTY), color: color}, nil
}

// repositories are the paths to scan: the arguments, or the current directory
// when there are none. A bare run is the common case.
func repositories(args []string) []string {
	if len(args) == 0 {
		return []string{loc.RootModule}
	}

	return args
}

// validateRepositories checks every path before anything is scanned or
// printed: each one must exist and be a directory.
func validateRepositories(repos []string) error {
	for _, repo := range repos {
		info, err := os.Stat(repo)

		switch {
		case errors.Is(err, os.ErrNotExist):
			return fmt.Errorf("%w: %s", errNoSuchDirectory, repo)
		case err != nil:
			return fmt.Errorf("cannot read %s: %w", repo, err)
		case !info.IsDir():
			return fmt.Errorf("%w: %s", errNotDirectory, repo)
		}
	}

	return nil
}

// notice says on stderr what the table cannot: directories that were skipped
// and an empty scan. stdout stays pure data.
func notice(w io.Writer, report loc.Report, language loc.Language) {
	if len(report.Skipped) > 0 {
		// A notice that cannot be written is not worth failing a finished scan.
		_, _ = fmt.Fprintf(
			w,
			"modloc: skipped %s under '%s': %s\n",
			english.Plural(len(report.Skipped), "unreadable directory", "unreadable directories"),
			report.Repo,
			strings.Join(report.Skipped, ", "),
		)
	}

	if len(report.Modules) == 0 {
		_, _ = fmt.Fprintf(
			w,
			"modloc: no %s files under '%s'%s\n",
			language.Extension(),
			report.Repo,
			rerun(report.Repo, language),
		)
	}
}

// rerun is the hint for an empty scan: the same command for every other
// language, so a Go run inside a Rust repository finds its way at once.
func rerun(repo string, language loc.Language) string {
	hints := make([]string, 0, len(language.Others()))
	for _, other := range language.Others() {
		hints = append(hints, fmt.Sprintf("for %s run: modloc -l %s '%s'", other, other, repo))
	}

	if len(hints) == 0 {
		return ""
	}

	return "; " + strings.Join(hints, "; ")
}
