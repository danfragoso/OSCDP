package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

const (
	CONTROLLER_PORT = "/dev/ttyACM0"
)

type Controller struct {
	port io.ReadWriteCloser
}

func InitController() (*Controller, error) {
	options := serial.OpenOptions{
		PortName:        CONTROLLER_PORT,
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		return nil, err
	}

	port.Write([]byte("player_status|Player OK\r"))
	time.Sleep(10 * time.Millisecond)

	return &Controller{port}, nil
}

type KeyCommand struct {
	Event string `json:"event"`
	Key   string `json:"key"`
}

func (c *Controller) ListenKeys(keyPresses chan string) {
	scanner := bufio.NewScanner(c.port)
	for scanner.Scan() {
		line := scanner.Text()
		var command KeyCommand
		err := json.Unmarshal([]byte(line), &command)
		if err != nil {
			fmt.Printf("Error parsing Command JSON: %v", err)
			continue
		}

		if command.Key != "none" {
			keyPresses <- command.Key
		}
	}
}

func (c *Controller) WriteCommand(command string) error {
	_, err := c.port.Write([]byte(command + "\r"))
	time.Sleep(10 * time.Millisecond)
	return err
}
