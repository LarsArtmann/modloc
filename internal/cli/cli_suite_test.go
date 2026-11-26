package cli_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestCLI runs the behavior specs of the modloc command line.
func TestCLI(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "modloc Command Line Suite")
	cleanupBinary()
}
