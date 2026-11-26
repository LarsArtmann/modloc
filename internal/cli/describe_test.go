package cli_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v5 "github.com/larsartmann/cmdguard/v5/pkg/cmdguard/v5"
	"github.com/larsartmann/go-output"

	"github.com/larsartmann/modloc/internal/cli"
	"github.com/larsartmann/modloc/internal/loc"
	"github.com/larsartmann/modloc/internal/render"
)

// describeGolden pins the machine-readable contract: every flag, its
// default and allowed values, and every command.
var describeGolden = filepath.Join("testdata", "describe.golden.json")

var _ = Describe("the command contract", func() {
	contract := func() *v5.CLIDescription {
		c, err := cli.Build(cli.Environment{})
		Expect(err).NotTo(HaveOccurred())

		return c.Describe()
	}

	It("pins the whole contract: every flag, its default and values, and every command", func() {
		var buf bytes.Buffer

		c, err := cli.Build(cli.Environment{})
		Expect(err).NotTo(HaveOccurred())
		Expect(c.WriteDescription(&buf)).To(Succeed())

		if os.Getenv("MODLOC_UPDATE_GOLDEN") == "1" {
			Expect(os.WriteFile(describeGolden, buf.Bytes(), 0o640)).To(Succeed())
		}

		golden, err := os.ReadFile(describeGolden)
		Expect(err).NotTo(HaveOccurred(), "run with MODLOC_UPDATE_GOLDEN=1 to capture the contract")
		Expect(buf.String()).To(Equal(string(golden)))
	})

	DescribeTable("offers exactly the values the domain knows",
		func(flag string, allowed []string) {
			var values []string

			for _, f := range contract().Flags {
				if f.Name == flag {
					values = f.Values
				}
			}

			Expect(values).To(Equal(allowed))
		},
		Entry("--language", "language", names(loc.AllLanguages())),
		Entry("--format", "format", names(render.AllFormats())),
		Entry("--color", "color", colorNames()),
	)

	It("names every allowed value in its flag's help", func() {
		for _, f := range contract().Flags {
			for _, value := range f.Values {
				Expect(f.Usage).To(ContainSubstring(value), "--"+f.Name)
			}
		}
	})

	It("starts every flag at its documented default", func() {
		defaults := map[string]string{}

		for _, f := range contract().Flags {
			defaults["--"+f.Name] = f.Default
		}

		Expect(defaults).To(Equal(map[string]string{
			"--language": loc.Go.String(),
			"--format":   render.FormatAuto.String(),
			"--color":    output.ColorModeAuto.String(),
			"--help":     "false",
			"--no-color": "false",
			"--version":  "false",
		}))
	})
})

// names renders enums the way they are typed on the command line.
func names[T ~string](values []T) []string {
	out := make([]string, 0, len(values))

	for _, value := range values {
		out = append(out, string(value))
	}

	return out
}

// colorNames are the color modes go-output accepts.
func colorNames() []string {
	return names(output.AllColorModes)
}
