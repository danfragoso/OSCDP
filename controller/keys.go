package main

import (
	"machine"
	"time"
)

const (
	debounceTime = 250 * time.Millisecond
	tickInterval = 13 * time.Millisecond
)

var (
	lastKeyPress time.Time
)

var KeyMap = map[string]machine.Pin{
	"Play/Pause": machine.GP15, // A
	"Next":       machine.GP17, // B
	"Prev":       machine.GP19, // X
	"Eject":      machine.GP21, // Y
}

func initKeys() {
	for _, pin := range KeyMap {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
}

func listenKeys(keyPresses chan string) {
	for {
		for keyCode, key := range KeyMap {
			if checkKey(key) {
				keyPresses <- keyCode
			}
		}

		time.Sleep(tickInterval)
	}
}

func checkKey(key machine.Pin) bool {
	if !key.Get() && time.Since(lastKeyPress) > debounceTime {
		lastKeyPress = time.Now()
		return true
	}

	return false
}
