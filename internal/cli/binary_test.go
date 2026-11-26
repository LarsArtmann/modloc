package cli_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// built is the binary, compiled once per test process.
//
//nolint:gochecknoglobals // compiled once per test process.
var built struct {
	sync.Once
	path string
	err  error
}

// binary compiles cmd/modloc and returns the executable. Building needs the
// Go toolchain, which a deployed host does not have; callers skip first.
func binary() string {
	built.Do(func() {
		dir, err := os.MkdirTemp("", "modloc-binary-")
		if err != nil {
			built.err = err

			return
		}

		DeferCleanup(func() { _ = os.RemoveAll(dir) })

		built.path = filepath.Join(dir, "modloc")

		out, err := exec.Command("go", "build", "-o", built.path, "../../cmd/modloc").CombinedOutput() //nolint:gosec // the toolchain is resolved through PATH
		if err != nil {
			built.err = fmt.Errorf("go build: %w\n%s", err, out)
		}
	})

	Expect(built.err).NotTo(HaveOccurred())

	return built.path
}

// invocation is one run of the binary and what it produced.
type invocation struct {
	code   int
	stdout string
	stderr string
}

// invoke runs the binary in dir and returns its exit code and streams.
func invoke(bin string, dir string, args ...string) invocation {
	cmd := exec.Command(bin, args...) //nolint:gosec // the freshly built binary
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := cmd.Run()

	code := 0
	if err != nil {
		exitErr, exited := errors.AsType[*exec.ExitError](err)
		if !exited {
			Fail("could not run " + bin + ": " + err.Error())
		}

		code = exitErr.ExitCode()
	}

	return invocation{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

var _ = Describe("the built binary", Serial, func() {
	var bin string

	BeforeEach(func() {
		if _, err := exec.LookPath("go"); err != nil {
			Skip("VACUOUS: no Go toolchain here, so the binary cannot be built")
		}

		realSCC()
		bin = binary()
	})

	Describe("parity with the bash modloc", func() {
		DescribeTable("prints its TSV byte for byte",
			func(golden, name string, args []string) {
				root := tree(name)

				run := invoke(bin, root, args...)

				Expect(run.stderr).To(BeEmpty())
				Expect(run.stdout).To(Equal(goldenText(golden)))
			},
			Entry("for Go modules, run bare", "go", "contract", []string{}),
			Entry("for Rust crates", "rust", "contract", []string{"-l", "rust", "."}),
			Entry("for files outside every module (OTHER)", "other", "stray", []string{}),
			Entry("for modules of equal size, by path", "tie", "tie", []string{}),
			Entry("for a path that starts with a dash, after --", "dash", "dash", []string{"--", "-dash"}),
		)
	})

	DescribeTable("exits the way the contract says",
		func(name string, args []string, code int, stdout, stderr string) {
			run := invoke(bin, tree(name), args...)

			Expect(run.code).To(Equal(code), run.stderr)
			Expect(run.stdout).To(Equal(stdout))
			Expect(run.stderr).To(ContainSubstring(stderr))
		},
		Entry("help on stdout", "contract", []string{"--help"}, 0, "", ""),
		Entry("version on stdout", "contract", []string{"--version"}, 0, "", ""),
		Entry("doctor", "contract", []string{"doctor"}, 0, "", ""),
		Entry("a usage error", "contract", []string{"--bogus"}, 2, "", "unknown flag: --bogus"),
		Entry("a missing path", "contract", []string{"/no/such/dir"}, 1, "", "No such directory: /no/such/dir"),
		Entry("an empty scan", "stray", []string{"-l", "rust"}, 0,
			"repo\tmodule\tfiles\tcode\tcomment\tblank\tlines\n", "modloc: no .rs files under '.'"),
	)

	It("scans a repository through the real scc, comments and blank lines included", func() {
		root := tree("contract")
		write(filepath.Join(root, "internal", "x", "doc.go"), "// Package doc documents.\n\npackage doc\n")

		run := invoke(bin, root)

		Expect(run.code).To(Equal(0))
		Expect(run.stdout).To(Equal(tsv(".", ".\t3\t3\t1\t1\t5", "internal/y\t2\t2\t0\t0\t2")))
	})

	It("prints nothing but the report on stdout", func() {
		run := invoke(bin, tree("stray"))

		Expect(strings.HasSuffix(run.stdout, "\n")).To(BeTrue(), "every line ends with a newline")
		Expect(run.stdout).NotTo(ContainSubstring("\x1b["), "a piped report carries no escape codes")
	})
})

// goldenText is the bash modloc's output captured in testdata/parity.
func goldenText(name string) string {
	golden, err := os.ReadFile(filepath.Join("testdata", "parity", name+".tsv"))
	Expect(err).NotTo(HaveOccurred())

	return string(golden)
}
