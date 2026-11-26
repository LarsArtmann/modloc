package scc_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/modloc/internal/loc"
	"github.com/larsartmann/modloc/internal/scc"
)

// golden is real scc 4.1.0 output for the contract fixture: four Go files
// under two modules, three Rust files, and the three Cargo manifests.
var golden = filepath.Join("testdata", "by-file.json")

// fakeSCC writes an executable script that stands in for scc and returns its
// path. body runs under /bin/sh, inside the directory being counted.
func fakeSCC(body string) string {
	bin := filepath.Join(GinkgoT().TempDir(), "scc")
	Expect(os.WriteFile(bin, []byte("#!/bin/sh\n"+body), 0o750)).To(Succeed())

	return bin
}

// counter pins a Counter to bin and captures what scc writes to stderr.
func counter(bin string) (scc.Counter, *bytes.Buffer) {
	var diagnostics bytes.Buffer

	return scc.New(scc.WithBinary(bin), scc.WithStderr(&diagnostics)), &diagnostics
}

// goldenPath is the absolute path of the golden output, for a script that
// runs in another directory.
func goldenPath() string {
	path, err := filepath.Abs(golden)
	Expect(err).NotTo(HaveOccurred())

	return path
}

// goFiles are the Go files of the golden output, in scc's order.
func goFiles() []string {
	return []string{"main.go", "internal/x/x.go", "internal/y/y.go", "internal/y/deep/d.go"}
}

// sccFile is one code-only file.
func sccFile(path string, lines int) loc.File {
	return loc.File{Path: path, Lines: lines, Code: lines}
}

// filePaths lists the paths of files, in order.
func filePaths(files []loc.File) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}

	return paths
}

// write creates a file and its directories.
func write(path, content string) {
	Expect(os.MkdirAll(filepath.Dir(path), 0o750)).To(Succeed())
	Expect(os.WriteFile(path, []byte(content), 0o640)).To(Succeed())
}

// realSCC skips when no scc is on PATH, naming the reason.
func realSCC() {
	if _, err := exec.LookPath("scc"); err != nil {
		Skip("VACUOUS: scc is not on PATH, so the real engine cannot be exercised")
	}
}
