package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/larsartmann/modloc/internal/loc"
)

type fakeCounter struct{}

func (fakeCounter) CountFiles(context.Context, string) ([]loc.File, error) { return nil, nil }

func TestProbe(t *testing.T) {
	for _, args := range [][]string{{"--bogus"}, {"-l"}, {"-x"}, {"-l", "cpp"}, {"--format", "xml"}, {"--color", "blue"}, {"/nope"}, {"/etc/hostname"}, {"--help"}, {"-h"}, {"--version"}, {"-v"}, {"version"}, {"doctor"}, {"describe"}, {"completion", "bash"}, {}} {
		var out, errb bytes.Buffer
		d := deps{stdout: &out, stderr: &errb, counter: fakeCounter{}, probe: func(context.Context) (string, string, error) {
			return "/bin/scc", "scc version 4.1.0", nil
		}, version: "0.1.0"}
		c, err := Build(d)
		if err != nil {
			t.Fatal(err)
		}
		c.RootCommand().SetOut(&out)
		c.RootCommand().SetErr(&errb)
		err = execute(c, context.Background(), args)
		o, e := out.String(), errb.String()
		t.Logf("args=%q err=%v isFlagValueError=%v exit=%d\nSTDOUT:\n%s\nSTDERR:\n%s\n----", args, err, isFlagValueError(err), exitOf(err), o, e)
	}
}

func exitOf(err error) int {
	var coder interface{ ExitCode() int }
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}
	if err == nil {
		return 0
	}
	return 1
}

var _ = os.Stdout
