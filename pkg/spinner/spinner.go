package spinner

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"
)

const (
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
)

// Type represents the type of spinner
type Type string

const (
	BrailDot   Type = "brailDot"
	ClassicDot Type = "classicDot"
	DualBall   Type = "dualBall"
	Circle     Type = "circle"
)

var (
	// osCue is a channel to listen for os signals
	osCue            = make(chan os.Signal, 1)
	brailDotFrames   = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	classicDotFrames = []string{".  ", ".. ", "..."}
	classicFrames    = []string{"|", "/", "-", "\\"}
	dualBallFrames   = []string{" ·", "● ", "●●", " ●", "· "}
	circleFrames     = []string{"◐", "◓", "◑", "◒"}

	// out receives the spinner frames and cursor control sequences. It is
	// stderr so that stdout carries only the command result.
	out io.Writer = os.Stderr
	// interactive reports whether out is a terminal; nothing is drawn otherwise.
	interactive = isTerminal(os.Stderr)
)

// isTerminal reports whether f is a character device such as a terminal.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// Spinner represents the cli spinner
type Spinner struct {
	frames   []string
	stop     chan bool
	done     chan bool
	exitFunc func(int)
	out      io.Writer
	enabled  bool
}

// Start starts the spinner
func (s *Spinner) Start() {
	if !s.enabled {
		close(s.done)
		return
	}
	fmt.Fprint(s.out, hideCursor)
	s.run()
}

// Clear stops the spinner
func (s *Spinner) Clear() {
	s.stop <- true
	<-s.done
}

// Get returns the spinner
func Get(spinnerType Type) *Spinner {
	var frames []string
	switch spinnerType {
	case BrailDot:
		frames = brailDotFrames
	case ClassicDot:
		frames = classicDotFrames
	case DualBall:
		frames = dualBallFrames
	case Circle:
		frames = circleFrames
	default:
		frames = classicFrames
	}
	return &Spinner{
		frames:   frames,
		stop:     make(chan bool, 1),
		done:     make(chan bool),
		exitFunc: os.Exit,
		out:      out,
		enabled:  interactive,
	}
}

// run renders the spinner
func (s *Spinner) run() {
	signal.Notify(osCue, handledSignals...)
	i := 0
	for {
		select {
		case <-osCue:
			fmt.Fprint(s.out, showCursor)
			s.exitFunc(0)
		case <-s.stop:
			fmt.Fprint(s.out, "\r \r"+showCursor)
			s.done <- true
			return
		default:
			fmt.Fprintf(s.out, "\r%s", s.frames[i%len(s.frames)])
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// ShowCursor shows the cursor when the spinner output is a terminal.
func ShowCursor() {
	if interactive {
		fmt.Fprint(out, showCursor)
	}
}

// ClearLine erases the spinner frame when the spinner output is a terminal.
func ClearLine() {
	if interactive {
		fmt.Fprint(out, "\r \r")
	}
}
