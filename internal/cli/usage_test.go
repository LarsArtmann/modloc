package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("modloc", func() {
	var modloc *harness

	BeforeEach(func() {
		modloc = newHarness()
	})

	Describe("help", func() {
		DescribeTable("answers a help flag with the usage on stdout and exit 0",
			func(flag string) {
				Expect(modloc.run(flag)).To(Equal(0))
				Expect(modloc.report.String()).To(BeEmpty(), "help is not a report")
				Expect(modloc.errors.String()).To(BeEmpty())
				Expect(modloc.screen.String()).To(ContainSubstring("USAGE"))
				Expect(modloc.screen.String()).To(ContainSubstring("current directory is scanned"))
			},
			Entry("-h", "-h"),
			Entry("--help", "--help"),
		)

		It("prints the same help after a repository path", func() {
			Expect(modloc.run(modloc.root, "-h")).To(Equal(0))
			first := modloc.screen.String()

			modloc.screen.Reset()
			Expect(modloc.run("-h")).To(Equal(0))
			Expect(modloc.screen.String()).To(Equal(first))
		})

		It("does not scan anything for help", func() {
			Expect(modloc.run(modloc.root, "--help")).To(Equal(0))
			Expect(modloc.counter.asked).To(BeEmpty())
		})

		It("names the allowed values of every flag", func() {
			Expect(modloc.run("--help")).To(Equal(0))
			help := modloc.screen.String()

			Expect(help).To(ContainSubstring("(go, rust)"))
			Expect(help).To(ContainSubstring("(auto, table, tsv, csv, json, jsonl, markdown)"))
			Expect(help).To(ContainSubstring("auto, always, never"))
		})

		It("lists its commands", func() {
			Expect(modloc.run("--help")).To(Equal(0))

			Expect(modloc.screen.String()).To(ContainSubstring("doctor"))
			Expect(modloc.screen.String()).To(ContainSubstring("version"))
		})
	})

	Describe("usage errors", func() {
		DescribeTable("exit 2, print no table, and name the problem",
			func(args []string, problem string) {
				Expect(modloc.run(args...)).To(Equal(2))
				Expect(modloc.report.String()).To(BeEmpty())
				Expect(modloc.screen.String()).To(BeEmpty())
				Expect(modloc.errors.String()).To(ContainSubstring(problem))
			},
			Entry("-l without a value", []string{"-l"}, "needs an argument"),
			Entry("an unknown language", []string{"-l", "bogus"}, `"bogus"`),
			Entry("an unknown format", []string{"--format", "bogus"}, `"bogus"`),
			Entry("an unknown color mode", []string{"--color", "bogus"}, `"bogus"`),
			Entry("an unknown flag", []string{"--bogus"}, "unknown flag: --bogus"),
		)

		It("does not scan anything on a usage error", func() {
			Expect(modloc.run(modloc.root, "-l", "bogus")).To(Equal(2))
			Expect(modloc.counter.asked).To(BeEmpty())
		})
	})

	Describe("version", func() {
		It("prints the version for --version and exits 0", func() {
			Expect(modloc.run("--version")).To(Equal(0))
			Expect(modloc.report.String()).To(BeEmpty())
			Expect(modloc.screen.String()).To(ContainSubstring("1.2.3"))
		})

		It("prints the version for the version command and exits 0", func() {
			Expect(modloc.run("version")).To(Equal(0))
			Expect(modloc.report.String()).To(BeEmpty())
			Expect(modloc.screen.String()).To(ContainSubstring("1.2.3"))
		})

		It("names the commit it was built from", func() {
			Expect(modloc.run("--version")).To(Equal(0))
			Expect(modloc.screen.String()).To(ContainSubstring("modloc version 1.2.3 (abc1234)"))
		})
	})
})
