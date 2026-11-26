// Package render writes scan reports as tables: aligned for a terminal,
// delimited for spreadsheets and scripts, JSON for tools.
//
// Every format shows the same rows and columns. The human formats (table,
// markdown) group thousands and end with a TOTAL row; the data formats (tsv,
// csv, json, jsonl) stay pure numbers with nothing after the rows, so a
// consumer never has to filter a summary out.
package render

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/larsartmann/go-output"
)

// Format is an output format as typed for --format.
type Format string

const (
	// FormatAuto picks the output for the destination: Table when stdout is
	// a terminal, TSV when it is piped or redirected.
	FormatAuto Format = "auto"
	// FormatTable is a bordered, right-aligned terminal table.
	FormatTable Format = "table"
	// FormatTSV is tab-separated values, one line per module.
	FormatTSV Format = "tsv"
	// FormatCSV is comma-separated values with quoting where needed.
	FormatCSV Format = "csv"
	// FormatJSON is one indented JSON array of module objects.
	FormatJSON Format = "json"
	// FormatJSONL is one JSON module object per line.
	FormatJSONL Format = "jsonl"
	// FormatMarkdown is a GitHub-flavored Markdown table.
	FormatMarkdown Format = "markdown"
)

// ErrUnknownFormat rejects a value that names no format.
var ErrUnknownFormat = errors.New("unknown format")

// AllFormats lists every format in the order --format documents them.
func AllFormats() []Format {
	return []Format{FormatAuto, FormatTable, FormatTSV, FormatCSV, FormatJSON, FormatJSONL, FormatMarkdown}
}

// Parse converts s to a Format, or returns an error wrapping ErrUnknownFormat
// that names every allowed value.
func Parse(s string) (Format, error) {
	if slices.Contains(AllFormats(), Format(s)) {
		return Format(s), nil
	}

	return "", fmt.Errorf("%w: %q (allowed: %s)", ErrUnknownFormat, s, Names())
}

// Names lists every format the way it is typed, comma-separated.
func Names() string {
	names := make([]string, 0, len(AllFormats()))
	for _, format := range AllFormats() {
		names = append(names, string(format))
	}

	return strings.Join(names, ", ")
}

// String returns the format as typed on the command line.
func (f Format) String() string {
	return string(f)
}

// Output is a concrete output format. FormatAuto resolves to one before
// anything is written, so "write the auto format" cannot be expressed.
type Output string

const (
	OutputTable    Output = "table"
	OutputTSV      Output = "tsv"
	OutputCSV      Output = "csv"
	OutputJSON     Output = "json"
	OutputJSONL    Output = "jsonl"
	OutputMarkdown Output = "markdown"
)

// Resolve names the Output that f stands for. Auto is Table when the
// destination is a terminal and TSV otherwise; every other format is itself.
func (f Format) Resolve(terminal bool) Output {
	if f == FormatAuto {
		if terminal {
			return OutputTable
		}

		return OutputTSV
	}

	return Output(f)
}

// isHuman reports whether an output is read by people rather than parsed.
func (o Output) isHuman() bool {
	return o == OutputTable || o == OutputMarkdown
}

// wireFormat is the go-output format that renders the tabular outputs.
func (o Output) wireFormat() output.Format {
	return output.Format(o)
}
