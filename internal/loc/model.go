package loc

import "context"

// Module path constants. RootModule is the repository root's own module; a
// file there is attributed to it only when the root holds a marker.
// Unattributed collects files of the language that no marker encloses.
const (
	RootModule   = "."
	Unattributed = "OTHER"
)

// File is one counted source file, relative to the repository root.
type File struct {
	// Path is slash-separated and relative to the repository root.
	Path string
	// Lines is the total number of lines in the file.
	Lines int
	// Code is the number of lines that carry code.
	Code int
	// Comment is the number of comment lines.
	Comment int
	// Blank is the number of empty lines.
	Blank int
}

// Module is one report row: a logical module and the totals of its files.
type Module struct {
	// Path is the module directory relative to the repository root,
	// RootModule for the root module, Unattributed when no marker encloses
	// the files.
	Path string
	// Files is the number of counted files in the module.
	Files int
	// Lines, Code, Comment and Blank are the sums over those files.
	Lines   int
	Code    int
	Comment int
	Blank   int
}

// Totals sums a set of modules. It is deliberately not a Module: a total has
// no path, so it can never be mistaken for a row.
type Totals struct {
	// Modules is the number of modules summed.
	Modules int
	// Files is the number of counted files.
	Files int
	// Lines, Code, Comment and Blank are the sums over those files.
	Lines   int
	Code    int
	Comment int
	Blank   int
}

// Add returns the sum of two totals.
func (t Totals) Add(other Totals) Totals {
	return Totals{
		Modules: t.Modules + other.Modules,
		Files:   t.Files + other.Files,
		Lines:   t.Lines + other.Lines,
		Code:    t.Code + other.Code,
		Comment: t.Comment + other.Comment,
		Blank:   t.Blank + other.Blank,
	}
}

// Report is the scan result for one repository.
type Report struct {
	// Repo is the path exactly as the caller gave it.
	Repo string
	// Modules holds one row per logical module, by code lines descending;
	// equal counts are ordered by path.
	Modules []Module
	// Skipped lists directories that could not be read while looking for
	// markers, relative to Repo; their subtrees may hide modules.
	Skipped []string
}

// Total sums the modules of the report.
func (r Report) Total() Totals {
	var total Totals

	for _, module := range r.Modules {
		total = total.Add(Totals{
			Modules: 1,
			Files:   module.Files,
			Lines:   module.Lines,
			Code:    module.Code,
			Comment: module.Comment,
			Blank:   module.Blank,
		})
	}

	return total
}

// FileCounter counts the lines of every source file under a directory. The
// paths it returns are relative to that directory.
type FileCounter interface {
	CountFiles(ctx context.Context, dir string) ([]File, error)
}
