package render

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
	"github.com/larsartmann/go-output"
	_ "github.com/larsartmann/go-output/delimited" // registers csv and tsv
	_ "github.com/larsartmann/go-output/markdown"  // registers markdown
	_ "github.com/larsartmann/go-output/table"     // registers the terminal table
	"github.com/larsartmann/modloc/internal/loc"
)

// firstNumberColumn is where the numeric columns begin: repo and module are
// names, everything after them is a count.
const firstNumberColumn = 2

// Columns are the table headers in order; the JSON objects use the same names.
func Columns() []string {
	return []string{"repo", "module", "files", "code", "comment", "blank", "lines"}
}

// Write renders the reports in the given output. The rows of every report
// follow each other, repositories in the order they were given and each
// one's modules by size. color only affects the table output.
func Write(w io.Writer, reports []loc.Report, out Output, color output.ColorMode) error {
	switch out {
	case OutputJSON:
		return writeJSON(w, rows(reports))
	case OutputJSONL:
		return writeJSONL(w, rows(reports))
	case OutputTable, OutputTSV, OutputCSV, OutputMarkdown:
		return output.RenderTable(
			table(reports, out),
			out.wireFormat(),
			output.RenderOptions{Writer: w, ColorMode: color},
		)
	default:
		return fmt.Errorf("%w: %q", ErrUnknownFormat, out)
	}
}

// table builds the go-output table for an output: plain numbers and no
// summary for the data formats, grouped numbers and a TOTAL row for people.
func table(reports []loc.Report, out Output) *output.Table {
	data := output.NewTable(Columns())

	human := out.isHuman()
	cellRows := cells(reports, human)

	if human && len(cellRows) > 0 {
		cellRows = append(cellRows, footer(reports))
	}

	if out == OutputTable {
		cellRows = alignNumbers(cellRows)
	}

	body, total := splitFooter(cellRows, human)

	for _, cellRow := range body {
		data.AddRow(cellRow)
	}

	if total != nil {
		data.SetFooter(total)
	}

	return data
}

// splitFooter separates the last row of a human table as its TOTAL row. Data
// formats have no footer, and neither has an empty table.
func splitFooter(cellRows [][]string, human bool) (body [][]string, total []string) {
	if !human || len(cellRows) == 0 {
		return cellRows, nil
	}

	return cellRows[:len(cellRows)-1], cellRows[len(cellRows)-1]
}

// cells renders every module as one row of strings, grouping thousands for
// the human formats.
func cells(reports []loc.Report, grouped bool) [][]string {
	out := make([][]string, 0)

	for _, r := range rows(reports) {
		out = append(out, []string{
			r.Repo,
			r.Module,
			number(r.Files, grouped),
			number(r.Code, grouped),
			number(r.Comment, grouped),
			number(r.Blank, grouped),
			number(r.Lines, grouped),
		})
	}

	return out
}

// number writes a count the way its reader wants it: exact for data,
// grouped for people.
func number(value int, grouped bool) string {
	if grouped {
		return humanize.Comma(int64(value))
	}

	return strconv.Itoa(value)
}

// footer is the TOTAL row of the human formats: every report summed.
func footer(reports []loc.Report) []string {
	var total loc.Totals

	for _, report := range reports {
		total = total.Add(report.Total())
	}

	return []string{
		"TOTAL",
		english.PluralWord(total.Modules, "module", "modules"),
		humanize.Comma(int64(total.Files)),
		humanize.Comma(int64(total.Code)),
		humanize.Comma(int64(total.Comment)),
		humanize.Comma(int64(total.Blank)),
		humanize.Comma(int64(total.Lines)),
	}
}

// alignNumbers left-pads every numeric column to its widest cell. go-output
// tables carry strings only, so padding is what makes counts line up on
// their right edge in a terminal.
func alignNumbers(cellRows [][]string) [][]string {
	widths := make([]int, len(Columns()))

	for _, cellRow := range cellRows {
		for column := firstNumberColumn; column < len(cellRow); column++ {
			widths[column] = max(widths[column], len(cellRow[column]))
		}
	}

	out := make([][]string, 0, len(cellRows))

	for _, cellRow := range cellRows {
		padded := slices.Clone(cellRow)
		for column := firstNumberColumn; column < len(padded); column++ {
			padded[column] = strings.Repeat(" ", widths[column]-len(padded[column])) + padded[column]
		}

		out = append(out, padded)
	}

	return out
}
