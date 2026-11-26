package render_test

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-output"

	"github.com/larsartmann/modloc/internal/loc"
	"github.com/larsartmann/modloc/internal/render"
)

// service is a repository with two modules: its root, with 1,200 code lines,
// and a nested one with 15.
func service() loc.Report {
	return loc.Report{
		Repo: "svc",
		Modules: []loc.Module{
			{Path: loc.RootModule, Files: 2, Code: 1200, Comment: 10, Blank: 30, Lines: 1240},
			{Path: "internal/api", Files: 1, Code: 15, Blank: 1, Lines: 16},
		},
	}
}

// library is a second repository with a single module.
func library() loc.Report {
	return loc.Report{
		Repo: "lib",
		Modules: []loc.Module{
			{Path: "auth", Files: 4, Code: 900, Lines: 900},
		},
	}
}

// bare is a report for a repository where the scan matched nothing.
func bare() loc.Report {
	return loc.Report{Repo: "empty"}
}

// written renders the reports in out, without color, and returns the text.
func written(reports []loc.Report, out render.Output) string {
	var buf bytes.Buffer

	Expect(render.Write(&buf, reports, out, output.ColorModeNever)).To(Succeed())

	return buf.String()
}

// lines joins the lines of an output and ends it with a newline, the way
// every text format ends.
func lines(lns ...string) string {
	return strings.Join(lns, "\n") + "\n"
}

// block splits an output into its lines, the last of which is empty after
// the final newline.
func block(out string) []string {
	return strings.Split(out, "\n")
}

// rowBlock is the line-split form of lines: what block returns for that
// output. Comparing a table line by line names the line that differs.
func rowBlock(lns ...string) []string {
	return append(lns, "")
}

// errBroken is what a broken destination reports.
var errBroken = errors.New("broken pipe")

// brokenWriter fails every write.
type brokenWriter struct{}

// Write fails.
func (brokenWriter) Write([]byte) (int, error) {
	return 0, errBroken
}
