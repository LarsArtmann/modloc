package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("modloc", func() {
	var h *harness

	BeforeEach(func() {
		h = newHarness()
	})

	Describe("help", func() {
		DescribeTable("answers a help flag with the usage on stdout and exit 0",
			func(flag string) {
				Expect(h.run(flag)).To(Equal(0))
				Expect(h.report.String()).To(BeEmpty(), "help is not a report")
				Expect(h.errors.String()).To(BeEmpty())
				Expect(h.screen.String()).To(ContainSubstring("USAGE"))
				Expect(h.screen.String()).To(ContainSubstring("current directory is scanned"))
			},
			Entry("-h", "-h"),
			Entry("--help", "--help"),
		)

		It("prints the same help after a repository path", func() {
			Expect(h.run(h.root, "-h")).To(Equal(0))
			first := h.screen.String()

			h.screen.Reset()
			Expect(h.run("-h")).To(Equal(0))
			Expect(h.screen.String()).To(Equal(first))
		})

		It("does not scan anything for help", func() {
			Expect(h.run(h.root, "--help")).To(Equal(0))
			Expect(h.counter.asked).To(BeEmpty())
		})

		It("names the allowed values of every flag", func() {
			Expect(h.run("--help")).To(Equal(0))
			help := h.screen.String()

			Expect(help).To(ContainSubstring("(go, rust)"))
			Expect(help).To(ContainSubstring("(auto, table, tsv, csv, json, jsonl, markdown)"))
			Expect(help).To(ContainSubstring("auto, always, never"))
		})

		It("lists its commands", func() {
			Expect(h.run("--help")).To(Equal(0))

			Expect(h.screen.String()).To(ContainSubstring("doctor"))
			Expect(h.screen.String()).To(ContainSubstring("version"))
		})
	})

	Describe("usage errors", func() {
		DescribeTable("exit 2, print no table, and name the problem",
			func(args []string, problem string) {
				Expect(h.run(args...)).To(Equal(2))
				Expect(h.report.String()).To(BeEmpty())
				Expect(h.screen.String()).To(BeEmpty())
				Expect(h.errors.String()).To(ContainSubstring(problem))
			},
			Entry("-l without a value", []string{"-l"}, "needs an argument"),
			Entry("an unknown language", []string{"-l", "bogus"}, `"bogus"`),
			Entry("an unknown format", []string{"--format", "bogus"}, `"bogus"`),
			Entry("an unknown color mode", []string{"--color", "bogus"}, `"bogus"`),
			Entry("an unknown flag", []string{"--bogus"}, "unknown flag: --bogus"),
		)

		It("does not scan anything on a usage error", func() {
			Expect(h.run(h.root, "-l", "bogus")).To(Equal(2))
			Expect(h.counter.asked).To(BeEmpty())
		})
	})

	Describe("version", func() {
		It("prints the version for --version and exits 0", func() {
			Expect(h.run("--version")).To(Equal(0))
			Expect(h.report.String()).To(BeEmpty())
			Expect(h.screen.String()).To(ContainSubstring("1.2.3"))
		})

		It("prints the version for the version command and exits 0", func() {
			Expect(h.run("version")).To(Equal(0))
			Expect(h.report.String()).To(BeEmpty())
			Expect(h.screen.String()).To(ContainSubstring("1.2.3"))
		})

		It("names the commit it was built from", func() {
			Expect(h.run("--version")).To(Equal(0))
			Expect(h.screen.String()).To(ContainSubstring("modloc version 1.2.3 (abc1234)"))
		})
	})
})
