// Package scc counts file lines by running the scc binary.
//
// scc is the counting engine: its language table decides what a code, a
// comment and a blank line is, for every language it knows. modloc runs it
// once per repository with --by-file --format json and keeps the per-file
// counts, so language coverage follows the installed scc.
package scc

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/larsartmann/modloc/internal/loc"
)

// Binary is the executable modloc runs.
const Binary = "scc"

var (
	// ErrNotFound reports that no scc executable is on PATH.
	ErrNotFound = errors.New("scc not found on PATH")
	// errEmptyOutput reports an scc run that printed nothing.
	errEmptyOutput = errors.New("scc wrote no output")
	// errInvalidJSON reports an scc run whose output is not the JSON it promises.
	errInvalidJSON = errors.New("scc wrote invalid JSON")
)

// pipeCloseGrace bounds how long CountFiles waits for scc's output pipes
// after scc itself has ended. A process scc left behind can keep them open;
// without a bound, a cancelled scan would hang until that process exits.
const pipeCloseGrace = 2 * time.Second

// rawOutputHint is the command that shows what scc really printed, for an
// output that modloc could not read.
const rawOutputHint = "; run 'scc --by-file --format json .' there to see what it printed"

// installHint names the ways to get scc; the Nix package built from this
// repository already wraps scc, so only other installs need it.
const installHint = "install scc, for example `go install github.com/boyter/scc/v4@latest` " +
	"or `brew install scc` (the Nix package from the modloc repository wraps scc already)"

// countArgs make scc print one JSON object per file.
func countArgs() []string {
	return []string{"--by-file", "--format", "json", "."}
}

// Counter is a loc.FileCounter backed by the scc binary.
type Counter struct {
	// lookPath resolves the scc executable.
	lookPath func(file string) (string, error)
	// stderr receives scc's own diagnostics.
	stderr io.Writer
}

// Option configures a Counter.
type Option func(*Counter)

// WithBinary pins the scc executable instead of resolving scc on PATH.
func WithBinary(bin string) Option {
	return func(c *Counter) {
		c.lookPath = func(string) (string, error) { return exec.LookPath(bin) }
	}
}

// WithStderr sends scc's diagnostics to w instead of the process's stderr.
func WithStderr(w io.Writer) Option {
	return func(c *Counter) { c.stderr = w }
}

// New returns a Counter that runs the scc found on PATH and lets scc's
// diagnostics through to the process's stderr.
func New(opts ...Option) Counter {
	counter := Counter{lookPath: exec.LookPath, stderr: os.Stderr}

	for _, opt := range opts {
		opt(&counter)
	}

	return counter
}

// CountFiles runs scc in dir and returns the per-file counts, with paths
// relative to dir. A missing scc, a failing scc and unreadable output are
// errors; scc's diagnostics pass through in every case.
func (c Counter) CountFiles(ctx context.Context, dir string) ([]loc.File, error) {
	bin, err := c.lookup()
	if err != nil {
		return nil, err
	}

	// scc prints one JSON object per file of the directory it is started in.
	cmd := exec.CommandContext(ctx, bin, countArgs()...) //nolint:gosec // scc is resolved through PATH
	cmd.Dir = dir
	cmd.WaitDelay = pipeCloseGrace

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("pipe %s output: %w", Binary, err)
	}

	var diagnostics bytes.Buffer

	cmd.Stderr = &diagnostics

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s in %s: %w", bin, dir, err)
	}

	files, decodeErr := decodeFiles(dir, stdout)
	waitErr := cmd.Wait()

	return c.settle(ctx, dir, files, decodeErr, waitErr, diagnostics.String())
}

// Probe resolves the scc executable and asks it for its version. It backs the
// doctor command: the result names what will run, so a wrong scc on PATH is
// visible.
func (c Counter) Probe(ctx context.Context) (string, string, error) {
	path, err := c.lookup()
	if err != nil {
		return "", "", err
	}

	out, err := exec.CommandContext(ctx, path, "--version").Output() //nolint:gosec // scc is resolved through PATH
	if err != nil {
		return "", "", fmt.Errorf("%s --version failed: %w", path, err)
	}

	return path, strings.TrimSpace(string(out)), nil
}

// settle decides the outcome once scc has exited: a killed or failing scc
// wins over a decode error, and diagnostics reach stderr only on success
// (they are part of the error message otherwise).
func (c Counter) settle(
	ctx context.Context,
	dir string,
	files []loc.File,
	decodeErr, waitErr error,
	diagnostics string,
) ([]loc.File, error) {
	switch {
	case ctx.Err() != nil:
		return nil, fmt.Errorf("count %s: %w", dir, ctx.Err())
	case waitErr != nil:
		return nil, fmt.Errorf("scc failed in %s: %w%s", dir, waitErr, detail(diagnostics))
	case decodeErr != nil:
		return nil, decodeErr
	}

	c.emit(diagnostics)

	return files, nil
}

// lookup resolves the scc executable or returns the not-found error with its
// fix attached.
func (c Counter) lookup() (string, error) {
	bin, err := c.lookPath(Binary)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrNotFound, installHint)
	}

	return bin, nil
}

// emit forwards scc's diagnostics, if it wrote any.
func (c Counter) emit(diagnostics string) {
	if diagnostics != "" && c.stderr != nil {
		_, _ = io.WriteString(c.stderr, diagnostics)
	}
}

// detail appends scc's last output lines to an error message.
func detail(diagnostics string) string {
	const tailLines = 5

	if strings.TrimSpace(diagnostics) == "" {
		return ""
	}

	lines := strings.Split(strings.TrimRight(diagnostics, "\n"), "\n")
	if len(lines) > tailLines {
		lines = append([]string{"…"}, lines[len(lines)-tailLines:]...)
	}

	return "\nscc wrote:\n  " + strings.Join(lines, "\n  ")
}

// fileJob is one entry of scc's --by-file JSON. The keys are scc's, in its
// own case; see the tagliatelle override in .golangci.yml.
type fileJob struct {
	Location string `json:"Location"`
	Lines    int    `json:"Lines"`
	Code     int    `json:"Code"`
	Comment  int    `json:"Comment"`
	Blank    int    `json:"Blank"`
}

// languageGroup is one element of scc's JSON array: the files of one language.
type languageGroup struct {
	Files []fileJob `json:"Files"`
}

// decodeFiles reads scc's --by-file JSON into loc files.
func decodeFiles(dir string, r io.Reader) ([]loc.File, error) {
	var groups []languageGroup

	err := json.UnmarshalDecode(jsontext.NewDecoder(r), &groups)

	switch {
	case errors.Is(err, io.EOF):
		return nil, fmt.Errorf("%w in %s", errEmptyOutput, dir)
	case err != nil:
		return nil, fmt.Errorf("%w in %s: %w%s", errInvalidJSON, dir, err, rawOutputHint)
	}

	files := make([]loc.File, 0)

	for _, group := range groups {
		for _, job := range group.Files {
			files = append(files, loc.File{
				Path:    loc.Normalize(job.Location),
				Lines:   job.Lines,
				Code:    job.Code,
				Comment: job.Comment,
				Blank:   job.Blank,
			})
		}
	}

	return files, nil
}
