package cli_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/scc"
)

var _ = Describe("modloc", func() {
	var modloc *harness

	BeforeEach(func() {
		modloc = newHarness()
	})

	Describe("doctor", func() {
		It("names the scc binary and its version", func() {
			Expect(modloc.run("doctor")).To(Equal(0))
			Expect(modloc.report.String()).To(Equal("scc  ok  /usr/bin/scc (scc version 4.1.0)\n"))
			Expect(modloc.errors.String()).To(BeEmpty())
		})

		It("fails with the reason when scc is missing", func() {
			modloc.probe = func(context.Context) (string, string, error) {
				return "", "", fmt.Errorf("%w: install scc", scc.ErrNotFound)
			}

			Expect(modloc.run("doctor")).To(Equal(1))
			Expect(modloc.report.String()).To(BeEmpty())
			Expect(strings.ToLower(modloc.errors.String())).To(ContainSubstring("scc not found on path"), modloc.errors.String())
		})
	})
})
