package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	cases := []struct {
		args     []string
		code     int
		inStdout string
		inStderr string
	}{
		{nil, 2, "", "Usage:"},
		{[]string{"version"}, 0, "agentpool dev", ""},
		{[]string{"help"}, 0, "Usage:", ""},
		{[]string{"bogus"}, 2, "", `unknown command "bogus"`},
		{[]string{"serve", "-addr", "0.0.0.0:43819"}, 1, "", "refusing to start"},
	}
	for _, c := range cases {
		var stdout, stderr bytes.Buffer
		if code := run(c.args, &stdout, &stderr); code != c.code {
			t.Errorf("run(%q) = %d, want %d (stderr: %s)", c.args, code, c.code, stderr.String())
		}
		if !strings.Contains(stdout.String(), c.inStdout) {
			t.Errorf("run(%q) stdout = %q, want it to contain %q", c.args, stdout.String(), c.inStdout)
		}
		if !strings.Contains(stderr.String(), c.inStderr) {
			t.Errorf("run(%q) stderr = %q, want it to contain %q", c.args, stderr.String(), c.inStderr)
		}
	}
}
