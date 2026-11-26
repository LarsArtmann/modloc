package render_test

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-output"

	"github.com/larsartmann/modloc/internal/loc"
	"github.com/larsartmann/modloc/internal/render"
)

var _ = Describe("Write", func() {
	Describe("tsv", func() {
		It("prints the header and one row per module, exact numbers, nothing after the rows", func() {
			Expect(written([]loc.Report{service()}, render.OutputTSV)).To(Equal(lines(
				"repo\tmodule\tfiles\tcode\tcomment\tblank\tlines",
				"svc\t.\t2\t1200\t10\t30\t1240",
				"svc\tinternal/api\t1\t15\t0\t1\t16",
			)))
		})

		It("keeps the rows of several repositories apart, in the order they were given", func() {
			Expect(written([]loc.Report{service(), library()}, render.OutputTSV)).To(Equal(lines(
				"repo\tmodule\tfiles\tcode\tcomment\tblank\tlines",
				"svc\t.\t2\t1200\t10\t30\t1240",
				"svc\tinternal/api\t1\t15\t0\t1\t16",
				"lib\tauth\t4\t900\t0\t0\t900",
			)))
		})

		It("prints only the header for a repository without modules", func() {
			Expect(written([]loc.Report{bare("empty")}, render.OutputTSV)).To(Equal(lines(
				"repo\tmodule\tfiles\tcode\tcomment\tblank\tlines",
			)))
		})
	})

	Describe("csv", func() {
		It("separates the same rows with commas", func() {
			Expect(written([]loc.Report{service()}, render.OutputCSV)).To(Equal(lines(
				"repo,module,files,code,comment,blank,lines",
				"svc,.,2,1200,10,30,1240",
				"svc,internal/api,1,15,0,1,16",
			)))
		})

		It("quotes a module path that contains a comma", func() {
			comma := loc.Report{
				Repo:    "svc",
				Modules: []loc.Module{{Path: "a,b", Files: 1, Code: 1, Lines: 1}},
			}

			Expect(written([]loc.Report{comma}, render.OutputCSV)).To(Equal(lines(
				"repo,module,files,code,comment,blank,lines",
				`svc,"a,b",1,1,0,0,1`,
			)))
		})
	})

	Describe("json", func() {
		It("writes one indented array of module objects with numeric counts", func() {
			Expect(written([]loc.Report{service()}, render.OutputJSON)).To(Equal(lines(
				"[",
				"  {",
				`    "repo": "svc",`,
				`    "module": ".",`,
				`    "files": 2,`,
				`    "code": 1200,`,
				`    "comment": 10,`,
				`    "blank": 30,`,
				`    "lines": 1240`,
				"  },",
				"  {",
				`    "repo": "svc",`,
				`    "module": "internal/api",`,
				`    "files": 1,`,
				`    "code": 15,`,
				`    "comment": 0,`,
				`    "blank": 1,`,
				`    "lines": 16`,
				"  }",
				"]",
			)))
		})

		It("writes an empty array for a repository without modules", func() {
			Expect(written([]loc.Report{bare("empty")}, render.OutputJSON)).To(Equal("[]\n"))
		})
	})

	Describe("jsonl", func() {
		It("writes one module object per line", func() {
			Expect(written([]loc.Report{service()}, render.OutputJSONL)).To(Equal(lines(
				`{"repo":"svc","module":".","files":2,"code":1200,"comment":10,"blank":30,"lines":1240}`,
				`{"repo":"svc","module":"internal/api","files":1,"code":15,"comment":0,"blank":1,"lines":16}`,
			)))
		})

		It("writes nothing for a repository without modules", func() {
			Expect(written([]loc.Report{bare("empty")}, render.OutputJSONL)).To(BeEmpty())
		})
	})

	Describe("markdown", func() {
		It("writes a table with grouped, right-aligned numbers that ends in a TOTAL row", func() {
			Expect(written([]loc.Report{service()}, render.OutputMarkdown)).To(Equal(lines(
				"| repo  | module       | files | code  | comment | blank | lines |",
				"|------|-------------|------:|------:|--------:|------:|------:|",
				"| svc   | .            |     2 | 1,200 |      10 |    30 | 1,240 |",
				"| svc   | internal/api |     1 |    15 |       0 |     1 |    16 |",
				"|------|-------------|------:|------:|--------:|------:|------:|",
				"| TOTAL | 2 modules    |     3 | 1,215 |      10 |    31 | 1,256 |",
			)))
		})

		It("sums every repository into the one TOTAL row", func() {
			Expect(written([]loc.Report{service(), library()}, render.OutputMarkdown)).To(Equal(lines(
				"| repo  | module       | files | code  | comment | blank | lines |",
				"|------|-------------|------:|------:|--------:|------:|------:|",
				"| svc   | .            |     2 | 1,200 |      10 |    30 | 1,240 |",
				"| svc   | internal/api |     1 |    15 |       0 |     1 |    16 |",
				"| lib   | auth         |     4 |   900 |       0 |     0 |   900 |",
				"|------|-------------|------:|------:|--------:|------:|------:|",
				"| TOTAL | 3 modules    |     7 | 2,115 |      10 |    31 | 2,156 |",
			)))
		})

		It("writes only the header for a repository without modules, with no TOTAL row", func() {
			Expect(written([]loc.Report{bare("empty")}, render.OutputMarkdown)).To(Equal(lines(
				"| repo | module | files | code | comment | blank | lines |",
				"|-----|-------|------:|-----:|--------:|------:|------:|",
			)))
		})

		It("stays plain Markdown when the output is piped", func() {
			var buf bytes.Buffer

			Expect(render.Write(&buf, []loc.Report{service()}, render.OutputMarkdown, output.ColorModeAuto)).To(Succeed())
			Expect(buf.String()).NotTo(ContainSubstring("\x1b["), "auto colors off the terminal would corrupt the Markdown")
		})

		It("emphasizes its header and TOTAL row when colors are forced", func() {
			var buf bytes.Buffer

			Expect(render.Write(&buf, []loc.Report{service()}, render.OutputMarkdown, output.ColorModeAlways)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring("\x1b[1m"))
		})
	})

	Describe("table", func() {
		It("draws a bordered table with right-aligned numbers and a TOTAL footer", func() {
			Expect(block(written([]loc.Report{service()}, render.OutputTable))).To(Equal(rowBlock(
				"╭───────┬──────────────┬───────┬───────┬─────────┬───────┬───────╮",
				"│ repo  │ module       │ files │ code  │ comment │ blank │ lines │",
				"├───────┼──────────────┼───────┼───────┼─────────┼───────┼───────┤",
				"│ svc   │ .            │     2 │ 1,200 │      10 │   30 │ 1,240 │",
				"│ svc   │ internal/api │     1 │    15 │       0 │     1 │    16 │",
				"│ TOTAL │ 2 modules    │     3 │ 1,215 │      10 │   31 │ 1,256 │",
				"╰───────┴──────────────┴───────┴───────┴─────────┴───────┴───────╯",
			)))
		})

		It("draws only the header for a repository without modules", func() {
			Expect(block(written([]loc.Report{bare("empty")}, render.OutputTable))).To(Equal(rowBlock(
				"╭──────┬────────┬───────┬──────┬─────────┬───────┬───────╮",
				"│ repo │ module │ files │ code │ comment │ blank │ lines │",
				"╰──────┴────────┴───────┴──────┴─────────┴───────┴───────╯",
			)))
		})

		It("carries no escape codes while colors are off", func() {
			Expect(written([]loc.Report{service()}, render.OutputTable)).NotTo(ContainSubstring("\x1b["))
		})

		It("is styled when colors are forced", func() {
			var buf bytes.Buffer

			Expect(render.Write(&buf, []loc.Report{service()}, render.OutputTable, output.ColorModeAlways)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring("\x1b["))
		})
	})

	Describe("a failing destination", func() {
		DescribeTable("reports the failure and keeps its cause",
			func(out render.Output) {
				err := render.Write(brokenWriter{}, []loc.Report{service()}, out, output.ColorModeNever)

				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, errBroken)).To(BeTrue(), "the caller must be able to recognize its own error")
			},
			Entry("tsv", render.OutputTSV),
			Entry("csv", render.OutputCSV),
			Entry("json", render.OutputJSON),
			Entry("jsonl", render.OutputJSONL),
			Entry("markdown", render.OutputMarkdown),
			Entry("table", render.OutputTable),
		)
	})

	Describe("an output that cannot be written", func() {
		It("rejects a name that is no format", func() {
			err := render.Write(&strings.Builder{}, nil, render.Output("yaml"), output.ColorModeNever)

			Expect(err).To(MatchError(render.ErrUnknownFormat))
		})

		It("rejects the auto format: it must be resolved for a destination first", func() {
			err := render.Write(&strings.Builder{}, nil, render.Output(render.FormatAuto), output.ColorModeNever)

			Expect(err).To(MatchError(render.ErrUnknownFormat))
		})
	})

	Describe("colors", func() {
		DescribeTable("never reach a data format",
			func(out render.Output) {
				var buf bytes.Buffer

				Expect(render.Write(&buf, []loc.Report{service()}, out, output.ColorModeAlways)).To(Succeed())
				Expect(buf.String()).NotTo(ContainSubstring("\x1b["))
			},
			Entry("tsv", render.OutputTSV),
			Entry("csv", render.OutputCSV),
			Entry("json", render.OutputJSON),
			Entry("jsonl", render.OutputJSONL),
		)
	})
})
