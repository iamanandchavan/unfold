package spinner

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// captureStdout returns everything written to os.Stdout while fn runs.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	os.Stdout = orig
	w.Close()
	b, _ := io.ReadAll(r)
	return string(b)
}

func TestSpinnerWritesOnlyToItsOutput(t *testing.T) {
	var buf bytes.Buffer
	stdout := captureStdout(t, func() {
		s := Get(BrailDot)
		s.out, s.enabled = &buf, true
		go s.Start()
		time.Sleep(250 * time.Millisecond)
		s.Clear()
	})
	if stdout != "" {
		t.Errorf("spinner wrote %q to stdout", stdout)
	}
	if got := buf.String(); !strings.Contains(got, brailDotFrames[0]) || !strings.HasSuffix(got, "\033[?25h") {
		t.Errorf("spinner output = %q, want frames and a final show-cursor sequence", got)
	}
}

func TestSpinnerIsSilentWhenNotInteractive(t *testing.T) {
	var buf bytes.Buffer
	stdout := captureStdout(t, func() {
		s := Get(BrailDot)
		s.out, s.enabled = &buf, false
		go s.Start()
		s.Clear() // must not block
		ShowCursor()
	})
	if stdout != "" || buf.Len() != 0 {
		t.Errorf("non-interactive spinner wrote stdout=%q out=%q", stdout, buf.String())
	}
}

func TestIsTerminal(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()
	if isTerminal(f) || isTerminal(w) {
		t.Error("regular files and pipes must not be treated as terminals")
	}
}

func TestCursorHelpersWriteToSpinnerOutputWhenInteractive(t *testing.T) {
	var buf bytes.Buffer
	origOut, origInteractive := out, interactive
	out, interactive = &buf, true
	defer func() { out, interactive = origOut, origInteractive }()

	stdout := captureStdout(t, func() {
		ClearLine()
		ShowCursor()
	})
	if stdout != "" || buf.String() != "\r \r\033[?25h" {
		t.Errorf("stdout=%q out=%q", stdout, buf.String())
	}
}
