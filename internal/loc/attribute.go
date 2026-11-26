package loc

import (
	"cmp"
	"path/filepath"
	"slices"
	"strings"
)

// Attribute totals the files of one language per logical module. Every file
// whose extension matches the language joins the nearest marker directory
// above it. Files with no marker above them join RootModule when the root is
// a marker, and Unattributed otherwise. The rows are ordered by code lines
// descending, then by path, so the biggest module is always first and equal
// modules keep a stable order.
func Attribute(files []File, markers []string, lang Language) []Module {
	set := make(map[string]struct{}, len(markers))
	for _, marker := range markers {
		set[Normalize(marker)] = struct{}{}
	}

	_, rootIsModule := set[RootModule]

	totals := make(map[string]*Module)

	for _, file := range files {
		path := Normalize(file.Path)
		if filepath.Ext(path) != lang.Extension() {
			continue
		}

		name := enclosing(path, set, rootIsModule)

		module := totals[name]
		if module == nil {
			module = &Module{Path: name}
			totals[name] = module
		}

		addFile(module, file)
	}

	return sortedModules(totals)
}

// enclosing returns the nearest marker directory around path, walking up. A
// path with no marker above it falls back to RootModule when the root is a
// marker, and to Unattributed when it is not.
func enclosing(path string, markers map[string]struct{}, rootIsModule bool) string {
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		if _, ok := markers[dir]; ok {
			return dir
		}

		if dir == RootModule {
			break
		}
	}

	if rootIsModule {
		return RootModule
	}

	return Unattributed
}

// addFile folds one file's lines into its module.
func addFile(module *Module, file File) {
	module.Files++
	module.Lines += file.Lines
	module.Code += file.Code
	module.Comment += file.Comment
	module.Blank += file.Blank
}

// sortedModules returns the modules as report rows: code lines descending,
// then path ascending.
func sortedModules(totals map[string]*Module) []Module {
	modules := make([]Module, 0, len(totals))
	for _, module := range totals {
		modules = append(modules, *module)
	}

	slices.SortFunc(modules, func(a, b Module) int {
		if byCode := cmp.Compare(b.Code, a.Code); byCode != 0 {
			return byCode
		}

		return strings.Compare(a.Path, b.Path)
	})

	return modules
}

// Normalize writes a path the way report rows are written: slash-separated
// and without a leading "./".
func Normalize(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(path), "./")
}
