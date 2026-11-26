package loc_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/loc"
)

// goFile is a counted Go file: one code line per argument.
func goFile(path string, code int) loc.File {
	return loc.File{Path: path, Lines: code, Code: code}
}

// rustFile is a counted Rust file: one code line per argument.
func rustFile(path string, code int) loc.File {
	return loc.File{Path: path, Lines: code, Code: code}
}

// touched writes an empty file, creating its directories.
func touched(path string) {
	Expect(os.MkdirAll(filepath.Dir(path), 0o750)).To(Succeed())
	Expect(os.WriteFile(path, nil, 0o640)).To(Succeed())
}

// repository is an empty directory standing in for a repository.
func repository() string {
	dir := GinkgoT().TempDir()
	touched(filepath.Join(dir, ".keep"))

	return dir
}

// paths lists the module paths of a report, in row order.
func paths(modules []loc.Module) []string {
	out := make([]string, 0, len(modules))
	for _, module := range modules {
		out = append(out, module.Path)
	}

	return out
}
