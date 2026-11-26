package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/larsartmann/modloc/internal/loc"
)

func TestProbe2(t *testing.T) {
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
	t.Logf("raw cobra usage:\n%s", c.RootCommand().UsageString())
	f := c.RootCommand().Flags().Lookup("language")
	t.Logf("language usage=%q def=%q type=%v", f.Usage, f.DefValue, f.Value.Type())
	_ = loc.RootModule
}
