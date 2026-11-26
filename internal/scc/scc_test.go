package scc_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/loc"
	"github.com/larsartmann/modloc/internal/scc"
)

var _ = Describe("Counter", func() {
	var dir string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
	})

	Describe("CountFiles", func() {
		It("returns every file scc reports, relative to the repository", func() {
			counter, _ := counter(fakeSCC("cat " + goldenPath() + "\n"))

			files, err := counter.CountFiles(context.Background(), dir)

			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(10))
			Expect(files[0]).To(Equal(sccFile("main.go", 1)))
			Expect(filePaths(files)[0:4]).To(Equal(goFiles()))
		})

		It("runs scc inside the repository, once, for its whole tree", func() {
			record := filepath.Join(GinkgoT().TempDir(), "record")
			counter, _ := counter(
				fakeSCC("printf '%s|%s\\n' \"$PWD\" \"$*\" >> " + record + "\ncat " + goldenPath() + "\n"),
			)

			_, err := counter.CountFiles(context.Background(), dir)

			Expect(err).NotTo(HaveOccurred())
			Expect(os.ReadFile(record)).To(Equal([]byte(dir + "|--by-file --format json .\n")))
		})

		It("reports an scc that is not installed, and how to get it", func() {
			counter, _ := counter(filepath.Join(dir, "missing", "scc"))

			_, err := counter.CountFiles(context.Background(), dir)

			Expect(errors.Is(err, scc.ErrNotFound)).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring("go install github.com/boyter/scc/v4")))
		})

		It("reports a failing scc with the lines it wrote", func() {
			counter, diagnostics := counter(fakeSCC("echo 'cannot walk: permission denied' >&2\nexit 3\n"))

			_, err := counter.CountFiles(context.Background(), dir)

			Expect(
				err,
			).To(MatchError(And(ContainSubstring("exit status 3"), ContainSubstring("cannot walk: permission denied"))))
			Expect(diagnostics.String()).To(BeEmpty(), "the lines are part of the error, not printed twice")
		})

		It("keeps only the last lines of a long failure", func() {
			counter, _ := counter(fakeSCC("for n in 1 2 3 4 5 6 7 8; do echo \"line $n\" >&2; done\nexit 1\n"))

			_, err := counter.CountFiles(context.Background(), dir)

			Expect(
				err,
			).To(MatchError(And(ContainSubstring("…"), ContainSubstring("line 4\n"), ContainSubstring("line 8"))))
			Expect(err).NotTo(MatchError(ContainSubstring("line 3")))
		})

		It("reports output that is not JSON", func() {
			counter, _ := counter(fakeSCC("echo 'not json'\n"))

			_, err := counter.CountFiles(context.Background(), dir)

			Expect(err).To(MatchError(And(ContainSubstring("scc output"), ContainSubstring("not json"))))
		})

		It("reports an scc that wrote nothing", func() {
			counter, _ := counter(fakeSCC("exit 0\n"))

			_, err := counter.CountFiles(context.Background(), dir)

			Expect(err).To(MatchError(ContainSubstring("no JSON written")))
		})

		It("lets scc's own diagnostics through on success", func() {
			counter, diagnostics := counter(fakeSCC("echo 'scc: noting something' >&2\ncat " + goldenPath() + "\n"))

			_, err := counter.CountFiles(context.Background(), dir)

			Expect(err).NotTo(HaveOccurred())
			Expect(diagnostics.String()).To(Equal("scc: noting something\n"))
		})
	})

	Describe("Probe", func() {
		It("names the binary and its version", func() {
			bin := fakeSCC("echo 'scc version 4.1.0'\n")
			counter, _ := counter(bin)

			path, version, err := counter.Probe(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(bin))
			Expect(version).To(Equal("scc version 4.1.0"))
		})

		It("fails when the binary cannot run", func() {
			counter, _ := counter(fakeSCC("exit 1\n"))

			_, _, err := counter.Probe(context.Background())

			Expect(err).To(MatchError(ContainSubstring("--version failed")))
		})

		It("reports an scc that is not installed", func() {
			counter, _ := counter(filepath.Join(dir, "missing"))

			_, _, err := counter.Probe(context.Background())

			Expect(errors.Is(err, scc.ErrNotFound)).To(BeTrue())
		})
	})

	Describe("cancellation", func() {
		It("stops when its context is cancelled", func() {
			ready := filepath.Join(GinkgoT().TempDir(), "ready")
			counter, _ := counter(fakeSCC("touch " + ready + "\nsleep 30\n"))
			ctx, cancel := context.WithCancel(context.Background())
			DeferCleanup(cancel)

			done := make(chan error, 1)

			go func() {
				_, err := counter.CountFiles(ctx, dir)
				done <- err
			}()

			Eventually(
				ready,
				10*time.Second,
			).Should(BeAnExistingFile(), "the fake scc must be running before it is cancelled")
			cancel()

			var err error

			Eventually(done, 10*time.Second).Should(Receive(&err))
			Expect(errors.Is(err, context.Canceled)).To(BeTrue())
		})
	})

	Describe("with the real scc", func() {
		BeforeEach(realSCC)

		It("counts the files of a repository, comments and blank lines included", func() {
			write(filepath.Join(dir, "go.mod"), "module example.com/root\n")
			write(filepath.Join(dir, "a.go"), "package main\n\n// a comment\nfunc main() {}\n")
			write(filepath.Join(dir, "sub", "b.go"), "package sub\n")

			files, err := scc.New(scc.WithStderr(io.Discard)).CountFiles(context.Background(), dir)

			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(Equal([]loc.File{
				{Path: "a.go", Lines: 4, Code: 2, Comment: 1, Blank: 1},
				{Path: "sub/b.go", Lines: 1, Code: 1},
			}))
		})

		It("counts lines so that code, comments and blanks add up to lines", func() {
			write(filepath.Join(dir, "a.go"), "package main\n\n// one\n// two\nfunc main() {\n\t// three\n}\n")
			write(filepath.Join(dir, "b.rs"), "fn main() {\n    // four\n}\n\n")
			write(filepath.Join(dir, "c.go"), strings.Repeat("x\n", 300))

			files, err := scc.New(scc.WithStderr(io.Discard)).CountFiles(context.Background(), dir)

			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(3))

			for _, file := range files {
				Expect(file.Code+file.Comment+file.Blank).To(Equal(file.Lines), file.Path)
			}
		})
	})
})
