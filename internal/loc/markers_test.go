package loc_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/loc"
)

var _ = Describe("FindMarkers", func() {
	var root string

	BeforeEach(func() {
		root = repository()
	})

	It("lists the root when it holds a marker", func() {
		touched(filepath.Join(root, "go.mod"))

		markers, skipped, err := loc.FindMarkers(root, loc.Go)

		Expect(err).NotTo(HaveOccurred())
		Expect(skipped).To(BeEmpty())
		Expect(markers).To(Equal([]string{"."}))
	})

	It("lists nested marker directories relative to the root, deepest included", func() {
		touched(filepath.Join(root, "go.mod"))
		touched(filepath.Join(root, "usermgmt", "go.mod"))
		touched(filepath.Join(root, "usermgmt", "oauth2", "go.mod"))

		markers, _, err := loc.FindMarkers(root, loc.Go)

		Expect(err).NotTo(HaveOccurred())
		Expect(markers).To(ConsistOf(".", "usermgmt", "usermgmt/oauth2"))
	})

	It("finds only the language's own marker", func() {
		touched(filepath.Join(root, "Cargo.toml"))
		touched(filepath.Join(root, "crates", "Cargo.toml"))

		_, _, err := loc.FindMarkers(root, loc.Go)
		Expect(err).NotTo(HaveOccurred())

		markers, _, err := loc.FindMarkers(root, loc.Rust)
		Expect(err).NotTo(HaveOccurred())
		Expect(markers).To(ConsistOf(".", "crates"))
	})

	DescribeTable("does not descend into",
		func(name string) {
			touched(filepath.Join(root, "go.mod"))
			touched(filepath.Join(root, name, "nested", "go.mod"))

			markers, skipped, err := loc.FindMarkers(root, loc.Go)

			Expect(err).NotTo(HaveOccurred())
			Expect(skipped).To(BeEmpty(), "a pruned directory is a choice, not a gap")
			Expect(markers).To(Equal([]string{"."}))
		},
		Entry("vendor", "vendor"),
		Entry("target", "target"),
		Entry("node_modules", "node_modules"),
		Entry(".git", ".git"),
	)

	DescribeTable("never prunes the root, whatever it is called",
		func(name string) {
			named := filepath.Join(root, name)
			touched(filepath.Join(named, "go.mod"))

			markers, _, err := loc.FindMarkers(named, loc.Go)

			Expect(err).NotTo(HaveOccurred())
			Expect(markers).To(Equal([]string{"."}))
		},
		Entry("vendor", "vendor"),
		Entry("target", "target"),
		Entry("node_modules", "node_modules"),
	)

	It("returns no markers for a repository without any", func() {
		markers, skipped, err := loc.FindMarkers(root, loc.Go)

		Expect(err).NotTo(HaveOccurred())
		Expect(skipped).To(BeEmpty())
		Expect(markers).To(BeNil())
	})

	It("fails for a root that does not exist", func() {
		_, _, err := loc.FindMarkers(filepath.Join(root, "missing"), loc.Go)

		Expect(err).To(MatchError(ContainSubstring("missing")))
	})

	It("skips an unreadable directory and reports it, keeping the readable tree", func() {
		touched(filepath.Join(root, "go.mod"))
		touched(filepath.Join(root, "secret", "go.mod"))
		touched(filepath.Join(root, "open", "go.mod"))

		secret := filepath.Join(root, "secret")
		Expect(os.Chmod(secret, 0o000)).To(Succeed())
		DeferCleanup(func() { _ = os.Chmod(secret, 0o700) })

		if _, err := os.ReadDir(secret); err == nil {
			Skip("VACUOUS: this process can read a mode-000 directory (running as root); nothing is skipped")
		}

		markers, skipped, err := loc.FindMarkers(root, loc.Go)

		Expect(err).NotTo(HaveOccurred())
		Expect(skipped).To(Equal([]string{"secret"}))
		Expect(markers).To(Equal([]string{".", "open"}))
	})
})

var _ = Describe("Language", func() {
	DescribeTable("parses",
		func(typed string, want loc.Language) {
			language, err := loc.Parse(typed)

			Expect(err).NotTo(HaveOccurred())
			Expect(language).To(Equal(want))
		},
		Entry("go", "go", loc.Go),
		Entry("rust", "rust", loc.Rust),
	)

	It("rejects an unknown language and names the allowed ones", func() {
		_, err := loc.Parse("cpp")

		Expect(errors.Is(err, loc.ErrUnknownLanguage)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("cpp")))
		Expect(err).To(MatchError(ContainSubstring("go, rust")))
	})

	DescribeTable("pairs a marker with an extension",
		func(language loc.Language, marker, extension string) {
			Expect(language.Marker()).To(Equal(marker))
			Expect(language.Extension()).To(Equal(extension))
		},
		Entry("Go", loc.Go, "go.mod", ".go"),
		Entry("Rust", loc.Rust, "Cargo.toml", ".rs"),
	)

	It("points at the other language for a re-run", func() {
		Expect(loc.Go.Others()).To(Equal([]loc.Language{loc.Rust}))
		Expect(loc.Rust.Others()).To(Equal([]loc.Language{loc.Go}))
	})

	It("lists go first and rust second", func() {
		Expect(loc.AllLanguages()).To(Equal([]loc.Language{loc.Go, loc.Rust}))
	})
})
