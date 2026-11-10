package helpers

import (
	"log"

	"github.com/aryannr97/unfold/pkg/spinner"
)

// Unblock is a channel to unblock the spinner
var Unblock = make(chan bool)

// GracefullyExit recovers from the panic and shows the cursor
func GracefullyExit() {
	spinner.ShowCursor()
	err := recover()
	if err != nil {
		spinner.ClearLine()
		log.Println(err)
	}
}
