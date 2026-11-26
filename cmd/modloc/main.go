// Command modloc counts lines of code per logical module (the nearest
// go.mod or Cargo.toml), not per repository.
package main

import (
	"context"
	"os"

	"github.com/larsartmann/modloc/internal/cli"
)

// version and commit are stamped at link time (-X main.version=...).
var (
	version = "dev"
	commit  = "" //nolint:gochecknoglobals // -X main.commit ldflags stamp; a parameter cannot reach the linker
)

func main() {
	os.Exit(cli.Main(context.Background(), os.Args[1:], version, commit))
}
