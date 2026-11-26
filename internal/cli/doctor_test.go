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
	var h *harness

	BeforeEach(func() {
		h = newHarness()
	})

	Describe("doctor", func() {
		It("names the scc binary and its version", func() {
			Expect(h.run("doctor")).To(Equal(0))
			Expect(h.report.String()).To(Equal("scc  ok  /usr/bin/scc (scc version 4.1.0)\n"))
			Expect(h.errors.String()).To(BeEmpty())
		})

		It("fails with the reason when scc is missing", func() {
			h.probe = func(context.Context) (string, string, error) {
				return "", "", fmt.Errorf("%w: install scc", scc.ErrNotFound)
			}

			Expect(h.run("doctor")).To(Equal(1))
			Expect(h.report.String()).To(BeEmpty())
			Expect(strings.ToLower(h.errors.String())).To(ContainSubstring("scc not found on path"), h.errors.String())
		})
	})
})
