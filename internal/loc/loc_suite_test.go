package loc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestLoc runs the behavior specs of logical-module attribution.
func TestLoc(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logical Module Attribution Suite")
}
