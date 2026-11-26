package render

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"

	"github.com/larsartmann/modloc/internal/loc"
)

// row is one module as the data formats write it. The field order is the
// column order of the table formats, and the keys are their headers.
type row struct {
	Repo    string `json:"repo"`
	Module  string `json:"module"`
	Files   int    `json:"files"`
	Code    int    `json:"code"`
	Comment int    `json:"comment"`
	Blank   int    `json:"blank"`
	Lines   int    `json:"lines"`
}

// writeJSON writes the rows as one indented JSON array: nothing else is in
// the document, so `jq` and friends read it straight.
func writeJSON(w io.Writer, rows []row) error {
	encoder := jsontext.NewEncoder(w, jsontext.WithIndent("  "))

	return json.MarshalEncode(encoder, rows)
}

// writeJSONL writes one JSON object per line and no array around them, so a
// consumer can read, filter and append the output line by line.
func writeJSONL(w io.Writer, rows []row) error {
	encoder := jsontext.NewEncoder(w)

	for _, r := range rows {
		err := json.MarshalEncode(encoder, r)
		if err != nil {
			return fmt.Errorf("encode module %s: %w", r.Module, err)
		}
	}

	return nil
}

// rows flattens the reports, keeping repository order and each report's row
// order.
func rows(reports []loc.Report) []row {
	out := make([]row, 0)

	for _, report := range reports {
		for _, module := range report.Modules {
			out = append(out, row{
				Repo:    report.Repo,
				Module:  module.Path,
				Files:   module.Files,
				Code:    module.Code,
				Comment: module.Comment,
				Blank:   module.Blank,
				Lines:   module.Lines,
			})
		}
	}

	return out
}
