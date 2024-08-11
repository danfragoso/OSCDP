package main

import (
	"time"
)

func init() {
	initDisplay()
	initKeys()
}

func main() {
	displayHeaderWithInfo("v0.1")
	time.Sleep(2 * time.Second)

	displayHeaderWithGlyph("☺")
	time.Sleep(2 * time.Second)

	displayHeaderWithInfo("Waiting Player...")
	clearAndRenderButtonCues()

	keyPresses := make(chan string)
	go listenKeys(keyPresses)

	displayCommands := make(chan *DisplayCommand)
	go listenDisplayCommands(displayCommands)

	for {
		select {
		case keyPress := <-keyPresses:
			println(`{"event": "keypress", "key": "` + keyPress + `"}`)

		case displayCommand := <-displayCommands:
			if displayCommand != nil {
				renderDisplayCommand(displayCommand.Section, displayCommand.Content)
			}

		default:
			time.Sleep(13 * time.Millisecond)
		}
	}
}
