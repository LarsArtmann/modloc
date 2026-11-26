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
	"github.com/larsartmann/go-output/markdown"
	_ "github.com/larsartmann/go-output/table" // registers the terminal table
	"github.com/larsartmann/modloc/internal/loc"
)

const (
	// firstNumberColumn is where the numeric columns begin: repo and module
	// are names, everything after them is a count.
	firstNumberColumn = 2
	// columnCount is the number of columns in every row.
	columnCount = 7
)

// Columns are the table headers in order; the JSON objects use the same names.
func Columns() []string {
	columns := headers()

	return slices.Clone(columns[:])
}

// headers are the table headers as a fixed-length array, so a mismatch with
// columnCount cannot compile.
func headers() [columnCount]string {
	return [columnCount]string{"repo", "module", "files", "code", "comment", "blank", "lines"}
}

// Write renders the reports in the given output. The rows of every report
// follow each other, repositories in the order they were given and each
// one's modules by size. color only affects the two formats people read
// (table, markdown); the data formats never carry escape codes.
func Write(w io.Writer, reports []loc.Report, out Output, color output.ColorMode) error {
	switch out {
	case OutputJSON:
		return writeJSON(w, rows(reports))
	case OutputJSONL:
		return writeJSONL(w, rows(reports))
	case OutputMarkdown:
		return writeMarkdown(w, table(reports, out), color)
	case OutputTable, OutputTSV, OutputCSV:
		return output.RenderTable(
			table(reports, out),
			out.wireFormat(),
			output.RenderOptions{Writer: w, ColorMode: color},
		)
	default:
		return fmt.Errorf("%w: %q", ErrUnknownFormat, out)
	}
}

// writeMarkdown writes the table as a GitHub-flavored Markdown table with
// its numeric columns right-aligned, through the alignment markers rather
// than padding, so the alignment survives any renderer and edits.
func writeMarkdown(w io.Writer, data *output.Table, color output.ColorMode) error {
	md := markdown.NewMarkdownTableFromTable(data)
	md.SetColorMode(color)

	for column := firstNumberColumn; column < columnCount; column++ {
		md.SetAlign(column, markdown.AlignRight)
	}

	return output.WriteRenderedRawFrom(w, md.Render, "markdown", "markdown")
}

// table builds the go-output table for an output: plain numbers and no
// summary for the data formats, grouped numbers and a TOTAL row for people.
func table(reports []loc.Report, out Output) *output.Table {
	data := output.NewTable(Columns())

	forPeople := out.isHuman()
	cellRows := cells(reports, forPeople)

	if forPeople && len(cellRows) > 0 {
		cellRows = append(cellRows, footer(reports))
	}

	if out == OutputTable {
		cellRows = alignNumbers(cellRows)
	}

	body, total := splitFooter(cellRows, forPeople)

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
func splitFooter(cellRows [][]string, forPeople bool) ([][]string, []string) {
	if !forPeople {
		return cellRows, nil
	}

	for index, row := range cellRows {
		if index == len(cellRows)-1 {
			return cellRows[:index], row
		}
	}

	return cellRows, nil
}

// cells renders every module as one row of strings, grouping thousands for
// the human formats.
func cells(reports []loc.Report, forPeople bool) [][]string {
	modules := rows(reports)
	out := make([][]string, 0, len(modules))

	for _, r := range modules {
		out = append(out, []string{
			r.Repo,
			r.Module,
			number(r.Files, forPeople),
			number(r.Code, forPeople),
			number(r.Comment, forPeople),
			number(r.Blank, forPeople),
			number(r.Lines, forPeople),
		})
	}

	return out
}

// number writes a count the way its reader wants it: exact for data,
// grouped for people.
func number(value int, forPeople bool) string {
	if forPeople {
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
		english.Plural(total.Modules, "module", "modules"),
		humanize.Comma(int64(total.Files)),
		humanize.Comma(int64(total.Code)),
		humanize.Comma(int64(total.Comment)),
		humanize.Comma(int64(total.Blank)),
		humanize.Comma(int64(total.Lines)),
	}
}

// alignNumbers left-pads every numeric cell to the width of its column. The
// terminal table carries strings only and places every cell at its column's
// left edge, so padding is what lines counts up on their right edge, under
// the last letter of their header. A header wider than every count therefore
// counts for the width too: without it, small counts would sit left under a
// wide header while large ones lined up on the right.
func alignNumbers(cellRows [][]string) [][]string {
	out := make([][]string, 0, len(cellRows))

	for _, cellRow := range cellRows {
		out = append(out, slices.Clone(cellRow))
	}

	for column, width := range headerWidths() {
		if column < firstNumberColumn {
			continue
		}

		widest := width

		for _, cellRow := range cellRows {
			if column < len(cellRow) {
				widest = max(widest, len(cellRow[column]))
			}
		}

		for _, padded := range out {
			if column < len(padded) {
				padded[column] = strings.Repeat(" ", widest-len(padded[column])) + padded[column]
			}
		}
	}

	return out
}

// headerWidths is the width of every column before any cell is counted:
// the width of its header.
func headerWidths() [columnCount]int {
	var widths [columnCount]int

	for column, header := range headers() {
		widths[column] = len(header)
	}

	return widths
}
