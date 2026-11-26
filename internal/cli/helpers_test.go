package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v5 "github.com/larsartmann/cmdguard/v5/pkg/cmdguard/v5"

	"github.com/larsartmann/modloc/internal/cli"
	"github.com/larsartmann/modloc/internal/loc"
)

// header is the first line of every TSV report.
const header = "repo\tmodule\tfiles\tcode\tcomment\tblank\tlines"

// goRows and rustRows are the contract tree's rows, each without the label
// of its repository.
var (
	goRows   = []string{".\t2\t3\t0\t0\t3", "internal/y\t2\t2\t0\t0\t2"}
	rustRows = []string{".\t1\t3\t0\t0\t3", "crates/one\t1\t2\t0\t0\t2", "crates/two\t1\t1\t0\t0\t1"}
)

// fixtureFile is one file of a fixture repository: its path and content.
type fixtureFile struct {
	path    string
	content string
}

// trees are the fixture repositories. Their files are exactly what the
// bash modloc was last run on for the parity goldens in testdata/parity: a
// file changed here without recapturing its golden is a parity failure.
//
var trees = map[string][]fixtureFile{
	"contract": {
		{"go.mod", "module example.com/root\n"},
		{"main.go", "package main\n"},
		{"internal/x/x.go", "package x\nfunc X() {}\n"},
		{"internal/y/go.mod", "module example.com/root/y\n"},
		{"internal/y/y.go", "package y\n"},
		{"internal/y/deep/d.go", "package d\n"},
		{"Cargo.toml", "[workspace]\nmembers = [\"crates/one\", \"crates/two\"]\n"},
		{"crates/one/Cargo.toml", "[package]\nname = \"one\"\n"},
		{"crates/one/src/lib.rs", "fn a() {}\nfn b() {}\n"},
		{"crates/two/Cargo.toml", "[package]\nname = \"two\"\n"},
		{"crates/two/src/lib.rs", "pub fn two() {}\n"},
		{"src/root.rs", "fn r1() {}\nfn r2() {}\nfn r3() {}\n"},
	},
	"stray": {
		{"stray.go", "package stray\nfunc S1() {}\nfunc S2() {}\n"},
		{"sub/go.mod", "module example.com/stray/sub\n"},
		{"sub/b.go", "package sub\nfunc B() {}\n"},
	},
	"tie": {
		{"zeta/go.mod", "module example.com/tie/zeta\n"},
		{"zeta/zeta.go", "package zeta\nfunc F() {}\n"},
		{"alpha/go.mod", "module example.com/tie/alpha\n"},
		{"alpha/alpha.go", "package alpha\nfunc F() {}\n"},
		{"mid/go.mod", "module example.com/tie/mid\n"},
		{"mid/mid.go", "package mid\nfunc F() {}\n"},
	},
	"dash": {
		{"-dash/go.mod", "module example.com/dash\n"},
		{"-dash/d.go", "package dash\n"},
	},
}

// tree writes the named fixture into a fresh directory and returns it.
func tree(name string) string {
	root := GinkgoT().TempDir()

	for _, f := range trees[name] {
		write(filepath.Join(root, f.path), f.content)
	}

	return root
}

// write creates a file and its directories.
func write(path, content string) {
	Expect(os.MkdirAll(filepath.Dir(path), 0o750)).To(Succeed())
	Expect(os.WriteFile(path, []byte(content), 0o640)).To(Succeed())
}

// lines joins the lines of an output and ends it with a newline.
func lines(lns ...string) string {
	return strings.Join(lns, "\n") + "\n"
}

// tsv is a TSV report: the header, then rows, each behind the label of its
// repository.
func tsv(label string, rows ...string) string {
	out := make([]string, 0, 1+len(rows))
	out = append(out, header)

	for _, row := range rows {
		out = append(out, label+"\t"+row)
	}

	return lines(out...)
}

// enter makes dir the working directory until the spec ends.
func enter(dir string) {
	previous, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	Expect(os.Chdir(dir)).To(Succeed())
	DeferCleanup(func() { _ = os.Chdir(previous) })
}

// contractCounts are the files scc reports for the contract tree.
func contractCounts() []loc.File {
	return []loc.File{
		{Path: "main.go", Lines: 1, Code: 1},
		{Path: "internal/x/x.go", Lines: 2, Code: 2},
		{Path: "internal/y/y.go", Lines: 1, Code: 1},
		{Path: "internal/y/deep/d.go", Lines: 1, Code: 1},
		{Path: "src/root.rs", Lines: 3, Code: 3},
		{Path: "crates/one/src/lib.rs", Lines: 2, Code: 2},
		{Path: "crates/two/src/lib.rs", Lines: 1, Code: 1},
	}
}

// fakeCounter stands in for scc: it answers each directory from a canned
// table and records every directory it was asked to count.
type fakeCounter struct {
	files map[string][]loc.File
	err   error
	// panics makes the next count crash instead of returning.
	panics bool
	// asked lists every directory CountFiles was called with.
	asked []string
}

// CountFiles returns the canned files of dir.
func (f *fakeCounter) CountFiles(_ context.Context, dir string) ([]loc.File, error) {
	f.asked = append(f.asked, dir)

	if f.panics {
		panic("scc exploded")
	}

	absolute, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return f.files[absolute], f.err
}

// foundScc answers doctor's probe with a healthy scc.
func foundScc(context.Context) (string, string, error) {
	return "/usr/bin/scc", "scc version 4.1.0", nil
}

// harness runs modloc in-process. The report and its notices go to the
// buffers a script would read; cmdguard's own output (help, version,
// errors) goes to what a person reads on the screen.
type harness struct {
	// report is stdout of the scan.
	report *bytes.Buffer
	// sink, when set, replaces report as the destination of the scan.
	sink io.Writer
	// notices is stderr of the scan.
	notices *bytes.Buffer
	// screen is stdout of help and version banners.
	screen *bytes.Buffer
	// errors is stderr of the error box.
	errors *bytes.Buffer
	// counter stands in for scc.
	counter *fakeCounter
	// probe answers doctor.
	probe cli.Probe
	// terminal says whether stdout is a terminal.
	terminal bool
	// root is the contract tree's directory.
	root string
}

// newHarness builds a harness over a fresh contract tree.
func newHarness() *harness {
	h := &harness{
		report:  &bytes.Buffer{},
		notices: &bytes.Buffer{},
		screen:  &bytes.Buffer{},
		errors:  &bytes.Buffer{},
		counter: &fakeCounter{},
		probe:   foundScc,
	}
	h.root = tree("contract")
	h.counter.files = map[string][]loc.File{h.root: contractCounts()}

	return h
}

// run executes one command line and returns the exit code it maps to.
func (h *harness) run(args ...string) int {
	var out io.Writer = h.report
	if h.sink != nil {
		out = h.sink
	}

	env := cli.Environment{
		Stdout:    out,
		Stderr:    h.notices,
		StdoutTTY: h.terminal,
		Counter:   h.counter,
		Scc:       h.probe,
		Stamp:     cli.Stamp{Version: "1.2.3", Commit: "abc1234"},
	}

	c, err := cli.Build(env)
	Expect(err).NotTo(HaveOccurred())
	c.RootCommand().SetOut(h.screen)
	c.RootCommand().SetErr(h.errors)

	return v5.ExitCode(cli.Execute(c, context.Background(), args))
}

// counts registers the files the fake counter answers for dir.
func (h *harness) counts(dir string, files ...loc.File) {
	if h.counter.files == nil {
		h.counter.files = map[string][]loc.File{}
	}

	h.counter.files[dir] = files
}

// brokenWriter fails every write.
type brokenWriter struct{}

// Write fails.
func (brokenWriter) Write([]byte) (int, error) {
	return 0, errors.New("disk full")
}

// realSCC skips when no scc is on PATH, naming the reason.
func realSCC() {
	if _, err := exec.LookPath("scc"); err != nil {
		Skip("VACUOUS: scc is not on PATH, so the real engine cannot be exercised")
	}
}
