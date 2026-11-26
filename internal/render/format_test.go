package render_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/render"
)

var _ = Describe("Format", func() {
	Describe("Parse", func() {
		It("accepts every format it advertises", func() {
			for _, format := range render.AllFormats() {
				parsed, err := render.Parse(format.String())

				Expect(err).NotTo(HaveOccurred(), format.String())
				Expect(parsed).To(Equal(format))
			}
		})

		It("rejects an unknown format and names every allowed one", func() {
			_, err := render.Parse("yaml")

			Expect(err).To(MatchError(render.ErrUnknownFormat))
			Expect(err).To(MatchError(ContainSubstring(render.Names())))
		})
	})

	Describe("Names", func() {
		It("lists the formats as they are typed, separated by commas", func() {
			Expect(render.Names()).To(Equal("auto, table, tsv, csv, json, jsonl, markdown"))
		})
	})

	Describe("Resolve", func() {
		It("is the bordered table when the destination is a terminal", func() {
			Expect(render.FormatAuto.Resolve(true)).To(Equal(render.OutputTable))
		})

		It("is TSV when the output is piped or redirected", func() {
			Expect(render.FormatAuto.Resolve(false)).To(Equal(render.OutputTSV))
		})

		It("leaves every explicit format alone, whatever the destination", func() {
			for _, terminal := range []bool{true, false} {
				Expect(render.FormatJSON.Resolve(terminal)).To(Equal(render.OutputJSON))
				Expect(render.FormatTSV.Resolve(terminal)).To(Equal(render.OutputTSV))
				Expect(render.FormatTable.Resolve(terminal)).To(Equal(render.OutputTable))
			}
		})
	})
})
