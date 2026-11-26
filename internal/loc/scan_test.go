package loc_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/loc"
)

// stubCounter answers every directory with the same canned result, so a scan
// can be driven without a counting engine.
type stubCounter struct {
	files []loc.File
	err   error
}

// CountFiles returns the canned result.
func (s stubCounter) CountFiles(context.Context, string) ([]loc.File, error) {
	return s.files, s.err
}

var _ = Describe("Scan", Serial, func() {
	var (
		root    string
		counter stubCounter
	)

	BeforeEach(func() {
		root = repository()
		touched(filepath.Join(root, "go.mod"))
		touched(filepath.Join(root, "usermgmt", "go.mod"))
		counter = stubCounter{files: []loc.File{
			goFile("main.go", 1),
			goFile("usermgmt/user.go", 5),
			goFile("stray/tool.go", 2),
		}}
	})

	It("attributes the counted files to the modules found on disk", func() {
		report, err := loc.Scan(context.Background(), root, loc.Go, counter)

		Expect(err).NotTo(HaveOccurred())
		Expect(paths(report.Modules)).To(Equal([]string{"usermgmt", "."}))
		Expect(report.Skipped).To(BeEmpty())
	})

	It("labels the report with the path as given", func() {
		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(root)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		report, err := loc.Scan(context.Background(), ".", loc.Go, counter)

		Expect(err).NotTo(HaveOccurred())
		Expect(report.Repo).To(Equal("."))
	})

	It("rejects a language that does not exist", func() {
		_, err := loc.Scan(context.Background(), root, loc.Language("cpp"), counter)

		Expect(errors.Is(err, loc.ErrUnknownLanguage)).To(BeTrue())
	})

	It("fails when the counter fails", func() {
		boom := errors.New("scc exploded")

		_, err := loc.Scan(context.Background(), root, loc.Go, stubCounter{err: boom})

		Expect(errors.Is(err, boom)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("count files")))
	})

	It("fails when the root cannot be walked", func() {
		missing := filepath.Join(root, "missing")

		_, err := loc.Scan(context.Background(), missing, loc.Go, counter)

		Expect(err).To(MatchError(ContainSubstring("missing")))
	})

	It("lists an unreadable directory beside the modules", func() {
		secret := filepath.Join(root, "secret")
		touched(filepath.Join(secret, "go.mod"))

		Expect(os.Chmod(secret, 0o000)).To(Succeed())
		DeferCleanup(func() { _ = os.Chmod(secret, 0o700) })

		if _, err := os.ReadDir(secret); err == nil {
			Skip("VACUOUS: this process can read a mode-000 directory (running as root)")
		}

		report, err := loc.Scan(context.Background(), root, loc.Go, counter)

		Expect(err).NotTo(HaveOccurred())
		Expect(report.Skipped).To(Equal([]string{"secret"}))
	})
})
