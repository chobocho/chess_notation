package web

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseWebCommand(t *testing.T) {
	cases := []struct {
		in   string
		want WebCmd
	}{
		{"exit", CmdExit},
		{"EXIT", CmdExit},
		{" quit ", CmdExit},
		{"q", CmdExit},
		{"help", CmdHelp},
		{"?", CmdHelp},
		{"H", CmdHelp},
		{"", CmdNoop},
		{"   ", CmdNoop},
		{"foo", CmdUnknown},
		{"helpme", CmdUnknown},
	}
	for _, c := range cases {
		if got := ParseWebCommand(c.in); got != c.want {
			t.Errorf("parse %q = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRunConsoleExit(t *testing.T) {
	in := strings.NewReader("exit\n")
	var out bytes.Buffer
	var called atomic.Bool
	done := make(chan struct{})
	go func() {
		RunConsole(context.Background(), in, &out, func() { called.Store(true) })
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunConsole did not return on exit")
	}
	if !called.Load() {
		t.Fatal("onExit was not invoked")
	}
	body := out.String()
	if !strings.Contains(body, "shutting down") {
		t.Errorf("missing shutdown message: %q", body)
	}
	if !strings.Contains(body, "help, ?, h") {
		t.Errorf("missing help banner: %q", body)
	}
}

func TestRunConsoleHelpAndUnknown(t *testing.T) {
	in := strings.NewReader("help\nfoo\nexit\n")
	var out bytes.Buffer
	done := make(chan struct{})
	go func() {
		RunConsole(context.Background(), in, &out, func() {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("did not exit in time")
	}
	body := out.String()
	// Banner printed at startup + again after "help".
	if strings.Count(body, "shut down the server") < 2 {
		t.Errorf("expected help text twice, got:\n%s", body)
	}
	if !strings.Contains(body, `unknown command "foo"`) {
		t.Errorf("missing unknown message: %q", body)
	}
}

func TestRunConsoleCancelsOnCtxDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Reader that never returns.
	pr, pw := io.Pipe()
	defer pw.Close()
	done := make(chan struct{})
	go func() {
		RunConsole(ctx, pr, io.Discard, func() {})
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunConsole did not return after ctx cancel")
	}
}

func TestRunConsoleEOFReturns(t *testing.T) {
	in := strings.NewReader("") // immediate EOF
	var out bytes.Buffer
	done := make(chan struct{})
	go func() {
		RunConsole(context.Background(), in, &out, func() {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunConsole did not return on EOF")
	}
}
