package cli_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/loc"
)

var _ = Describe("modloc", func() {
	var h *harness

	BeforeEach(func() {
		h = newHarness()
	})

	Describe("the scan", func() {
		Context("given a repository path", func() {
			It("prints one TSV row per logical module, largest first", func() {
				Expect(h.run(h.root)).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv(h.root, goRows...)))
				Expect(h.notices.String()).To(BeEmpty())
			})

			It("labels the rows with the path exactly as it was given", func() {
				enter(filepath.Dir(h.root))

				Expect(h.run("./"+filepath.Base(h.root))).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv("./"+filepath.Base(h.root), goRows...)))
			})

			It("accepts a path with spaces", func() {
				spaced := filepath.Join(GinkgoT().TempDir(), "my repo")
				write(filepath.Join(spaced, "go.mod"), "module example.com/spaced\n")
				h.counts(spaced, loc.File{Path: "a.go", Lines: 1, Code: 1})

				Expect(h.run(spaced)).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv(spaced, ".\t1\t1\t0\t0\t1")))
			})
		})

		Context("given no path", func() {
			It("scans the current directory and labels its rows with a dot", func() {
				enter(h.root)

				Expect(h.run()).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv(".", goRows...)))
			})
		})

		Context("given several paths", func() {
			It("prints every repository's rows, one after the other", func() {
				enter(h.root)

				Expect(h.run(h.root, ".")).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv(h.root, goRows...) + tsv(".", goRows...)))
			})
		})

		Context("given a path that is no directory", func() {
			It("exits 1, prints no table, and names the missing path", func() {
				Expect(h.run(h.root, "/no/such/dir")).To(Equal(1))
				Expect(h.report.String()).To(BeEmpty())
				Expect(h.errors.String()).To(ContainSubstring("No such directory: /no/such/dir"))
			})

			It("names a file that was given as a path", func() {
				file := filepath.Join(h.root, "main.go")

				Expect(h.run(file)).To(Equal(1))
				Expect(h.report.String()).To(BeEmpty())
				Expect(h.errors.String()).To(ContainSubstring("Not a directory: " + file))
			})

			It("checks every path before it scans any of them", func() {
				Expect(h.run(h.root, "/no/such/dir")).To(Equal(1))
				Expect(h.counter.asked).To(BeEmpty())
			})
		})

		Context("given options after the path", func() {
			It("accepts -l after the path", func() {
				Expect(h.run(h.root, "-l", "rust")).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv(h.root, rustRows...)))
			})

			It("accepts --format after the path", func() {
				Expect(h.run(h.root, "--format", "csv")).To(Equal(0))
				Expect(h.report.String()).To(Equal(lines(
					"repo,module,files,code,comment,blank,lines",
					h.root+",.,2,3,0,0,3",
					h.root+",internal/y,2,2,0,0,2",
				)))
			})
		})

		Context("given --", func() {
			It("scans a path that starts with a dash", func() {
				dash := tree("dash")
				enter(dash)
				h.counts("-dash", loc.File{Path: "d.go", Lines: 1, Code: 1})

				Expect(h.run("--", "-dash")).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv("-dash", ".\t1\t1\t0\t0\t1")))
			})
		})

		Context("when the scan matches no file", func() {
			It("exits 0, prints only the header, and says so on stderr", func() {
				empty := GinkgoT().TempDir()

				Expect(h.run(empty)).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv(empty)))
				Expect(h.notices.String()).To(Equal("modloc: no .go files under '" + empty + "'\n"))
			})

			It("points at the language that would find files", func() {
				empty := GinkgoT().TempDir()

				Expect(h.run(empty, "-l", "rust")).To(Equal(0))
				Expect(h.notices.String()).To(Equal("modloc: no .rs files under '" + empty + "'; for go run: modloc -l go '" + empty + "'\n"))
			})
		})

		Context("when a directory cannot be read", func() {
			It("says which directory it skipped and counts the rest", func() {
				secret := filepath.Join(h.root, "secret")
				write(filepath.Join(secret, "go.mod"), "module example.com/root/secret\n")

				Expect(os.Chmod(secret, 0o000)).To(Succeed())
				DeferCleanup(func() { _ = os.Chmod(secret, 0o700) })

				if _, err := os.ReadDir(secret); err == nil {
					Skip("VACUOUS: this process reads a mode-000 directory (running as root)")
				}

				Expect(h.run(h.root)).To(Equal(0))
				Expect(h.report.String()).To(Equal(tsv(h.root, goRows...)))
				Expect(h.notices.String()).To(Equal("modloc: skipped 1 unreadable directory under '" + h.root + "': secret\n"))
			})
		})

		Context("when the counter fails", func() {
			It("exits 1, prints no table, and reports the failure", func() {
				h.counter.err = errors.New("scc exploded")

				Expect(h.run(h.root)).To(Equal(1))
				Expect(h.report.String()).To(BeEmpty())
				Expect(h.errors.String()).To(ContainSubstring("scc exploded"))
			})

			It("reports a crash in the counter as a failure, not as a panic", func() {
				h.counter.panics = true

				Expect(h.run(h.root)).To(Equal(1))
				Expect(h.report.String()).To(BeEmpty())
			})
		})

		Context("when the report cannot be written", func() {
			It("exits 1 and reports the failed write", func() {
				h.sink = brokenWriter{}

				Expect(h.run(h.root)).To(Equal(1))
				Expect(h.errors.String()).To(ContainSubstring("write report"))
			})
		})
	})

	Describe("--format", func() {
		It("defaults to TSV when stdout is piped", func() {
			Expect(h.run(h.root)).To(Equal(0))
			Expect(h.report.String()).To(Equal(tsv(h.root, goRows...)))
		})

		It("defaults to the bordered table when stdout is a terminal", func() {
			h.terminal = true

			Expect(h.run(h.root)).To(Equal(0))
			Expect(h.report.String()).To(ContainSubstring("╭"))
			Expect(h.report.String()).To(ContainSubstring("TOTAL"))
		})

		It("writes csv", func() {
			Expect(h.run(h.root, "-f", "csv")).To(Equal(0))
			Expect(h.report.String()).To(Equal(lines(
				"repo,module,files,code,comment,blank,lines",
				h.root+",.,2,3,0,0,3",
				h.root+",internal/y,2,2,0,0,2",
			)))
		})

		It("writes json", func() {
			Expect(h.run(h.root, "-f", "json")).To(Equal(0))
			Expect(h.report.String()).To(Equal(lines(
				"[",
				"  {",
				`    "repo": "`+h.root+`",`,
				`    "module": ".",`,
				`    "files": 2,`,
				`    "code": 3,`,
				`    "comment": 0,`,
				`    "blank": 0,`,
				`    "lines": 3`,
				"  },",
				"  {",
				`    "repo": "`+h.root+`",`,
				`    "module": "internal/y",`,
				`    "files": 2,`,
				`    "code": 2,`,
				`    "comment": 0,`,
				`    "blank": 0,`,
				`    "lines": 2`,
				"  }",
				"]",
			)))
		})

		It("writes jsonl", func() {
			Expect(h.run(h.root, "-f", "jsonl")).To(Equal(0))
			Expect(h.report.String()).To(Equal(lines(
				`{"repo":"`+h.root+`","module":".","files":2,"code":3,"comment":0,"blank":0,"lines":3}`,
				`{"repo":"`+h.root+`","module":"internal/y","files":2,"code":2,"comment":0,"blank":0,"lines":2}`,
			)))
		})

		It("writes a table with a TOTAL row", func() {
			Expect(h.run(h.root, "-f", "table")).To(Equal(0))
			Expect(h.report.String()).To(ContainSubstring("TOTAL"))
			Expect(h.report.String()).To(ContainSubstring("2 modules"))
		})

		It("writes Markdown with right-aligned numbers and a TOTAL row", func() {
			Expect(h.run(h.root, "-f", "markdown")).To(Equal(0))
			Expect(h.report.String()).To(ContainSubstring("|------:|"))
			Expect(h.report.String()).To(ContainSubstring("| TOTAL  | 2 modules"))
		})
	})

})
