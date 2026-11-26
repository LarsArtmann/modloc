package loc_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/loc"
)

var _ = Describe("Attribute", func() {
	var markers []string

	Describe("when the repository root is a module", func() {
		BeforeEach(func() {
			markers = []string{".", "usermgmt", "usermgmt/oauth2"}
		})

		It("gives a file beside the root marker to the root module", func() {
			modules := loc.Attribute([]loc.File{goFile("main.go", 10)}, markers, loc.Go)

			Expect(paths(modules)).To(Equal([]string{"."}))
			Expect(modules[0].Code).To(Equal(10))
		})

		It("gives a file to the nearest marker above it", func() {
			modules := loc.Attribute(
				[]loc.File{goFile("usermgmt/oauth2/token.go", 1), goFile("usermgmt/user.go", 2)},
				markers,
				loc.Go,
			)

			Expect(paths(modules)).To(Equal([]string{"usermgmt", "usermgmt/oauth2"}))
		})

		It("never counts a nested module's lines toward the module that encloses it", func() {
			modules := loc.Attribute(
				[]loc.File{goFile("usermgmt/oauth2/token.go", 100), goFile("usermgmt/user.go", 5)},
				markers,
				loc.Go,
			)

			usermgmt := modules[slices.IndexFunc(modules, func(m loc.Module) bool { return m.Path == "usermgmt" })]
			Expect(usermgmt.Code).To(Equal(5), "the nested module's 100 lines stay with the nested module")
		})

		It("sums files and every kind of line per module", func() {
			files := []loc.File{
				{Path: "usermgmt/a.go", Lines: 11, Code: 7, Comment: 2, Blank: 2},
				{Path: "usermgmt/b.go", Lines: 5, Code: 4, Comment: 1, Blank: 0},
			}

			modules := loc.Attribute(files, markers, loc.Go)

			Expect(modules).To(Equal([]loc.Module{{
				Path: "usermgmt", Files: 2, Lines: 16, Code: 11, Comment: 3, Blank: 2,
			}}))
		})
	})

	Describe("when the repository root is not a module", func() {
		BeforeEach(func() {
			markers = []string{"adminui"}
		})

		It("collects files with no marker above them under OTHER", func() {
			modules := loc.Attribute([]loc.File{goFile("cmd/main.go", 3)}, markers, loc.Go)

			Expect(paths(modules)).To(Equal([]string{loc.Unattributed}))
		})

		It("still attributes the files below a marker", func() {
			modules := loc.Attribute(
				[]loc.File{goFile("cmd/main.go", 3), goFile("adminui/page.go", 4)},
				markers,
				loc.Go,
			)

			Expect(paths(modules)).To(Equal([]string{"adminui", loc.Unattributed}))
		})
	})

	It("ignores files of another language", func() {
		files := []loc.File{goFile("main.go", 10), rustFile("lib.rs", 20), goFile("notes.gotxt", 30)}

		modules := loc.Attribute(files, []string{"."}, loc.Go)

		Expect(modules).To(Equal([]loc.Module{{Path: ".", Files: 1, Lines: 10, Code: 10}}))
	})

	It("attributes Rust files by Cargo.toml", func() {
		files := []loc.File{rustFile("fuzz/src/lib.rs", 8), goFile("main.go", 99)}

		modules := loc.Attribute(files, []string{".", "fuzz"}, loc.Rust)

		Expect(paths(modules)).To(Equal([]string{"fuzz"}))
	})

	It("reads a leading ./ in a path as the repository root", func() {
		modules := loc.Attribute([]loc.File{goFile("./a/b.go", 1)}, []string{".", "a"}, loc.Go)

		Expect(paths(modules)).To(Equal([]string{"a"}))
	})

	It("does not mistake a path prefix for a directory", func() {
		modules := loc.Attribute([]loc.File{goFile("foobar/x.go", 1)}, []string{"foo"}, loc.Go)

		Expect(paths(modules)).To(Equal([]string{loc.Unattributed}))
	})

	Describe("row order", func() {
		It("puts the module with the most code first", func() {
			modules := loc.Attribute(
				[]loc.File{goFile("small/a.go", 1), goFile("big/b.go", 50), goFile("mid/c.go", 10)},
				[]string{"small", "big", "mid"},
				loc.Go,
			)

			Expect(paths(modules)).To(Equal([]string{"big", "mid", "small"}))
		})

		It("orders equal modules by path", func() {
			modules := loc.Attribute(
				[]loc.File{goFile("zeta/a.go", 5), goFile("alpha/b.go", 5), goFile("root.go", 5)},
				[]string{".", "zeta", "alpha"},
				loc.Go,
			)

			Expect(paths(modules)).To(Equal([]string{".", "alpha", "zeta"}))
		})
	})

	It("gives every line of every file to exactly one module", func() {
		files := []loc.File{
			goFile("a/x.go", 3), goFile("a/b/y.go", 4), goFile("a/b/c/z.go", 5), goFile("top.go", 6),
			rustFile("ignored.rs", 100),
		}

		modules := loc.Attribute(files, []string{".", "a", "a/b", "a/b/c"}, loc.Go)

		var fileCount, lineCount int

		for _, module := range modules {
			fileCount += module.Files
			lineCount += module.Lines
		}

		Expect(fileCount).To(Equal(4))
		Expect(lineCount).To(Equal(18))
	})

	It("returns no rows when nothing matches", func() {
		Expect(loc.Attribute(nil, []string{"."}, loc.Go)).To(BeEmpty())
	})
})

var _ = Describe("Report", func() {
	It("totals its modules", func() {
		report := loc.Report{
			Repo: ".",
			Modules: []loc.Module{
				{Path: "a", Files: 2, Lines: 20, Code: 10, Comment: 5, Blank: 5},
				{Path: "b", Files: 1, Lines: 10, Code: 8, Comment: 1, Blank: 1},
			},
		}

		Expect(report.Total()).To(Equal(loc.Totals{
			Modules: 2, Files: 3, Lines: 30, Code: 18, Comment: 6, Blank: 6,
		}))
	})

	It("totals to zero modules when nothing matched", func() {
		Expect(loc.Report{}.Total()).To(Equal(loc.Totals{}))
	})
})
