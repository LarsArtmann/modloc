package cli

import (
	"context"
	"fmt"
)

// runDoctor names the scc that modloc will run, or fails with the reason it
// cannot count.
func runDoctor(ctx context.Context, d deps) error {
	path, version, err := d.probe(ctx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(d.stdout, "scc  ok  %s (%s)\n", path, version)

	return err
}
