package cli

import (
	"context"
	"fmt"
)

// runDoctor names the scc that modloc will run, or fails with the reason it
// cannot count.
func runDoctor(ctx context.Context, env Environment) error {
	path, version, err := env.Scc(ctx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(env.Stdout, "scc  ok  %s (%s)\n", path, version)

	return err
}
