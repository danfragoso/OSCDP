package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"time"
)

const mpvSocketPath = "/tmp/oscdp-mpv-ipc"

type MPV struct {
	conn net.Conn
}

type MPVResponse struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func InitMPV() (*MPV, error) {
	cmd := exec.Command("mpv", "--idle", "--input-ipc-server="+mpvSocketPath)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	time.Sleep(5 * time.Second)

	conn, err := net.Dial("unix", mpvSocketPath)
	if err != nil {
		return nil, err
	}

	return &MPV{conn: conn}, nil
}

func (mpv *MPV) Stop() error {
	return mpv.SendSuccessCommand(`{"command": ["stop"]}`)
}

func (mpv *MPV) NextTrack() error {
	return mpv.SendSuccessCommand(`{"command": ["add", "chapter", 1]}`)
}

func (mpv *MPV) PreviousTrack() error {
	return mpv.SendSuccessCommand(`{"command": ["add", "chapter", -1]}`)
}

func (mpv *MPV) StartDisc() error {
	return mpv.SendSuccessCommand(`{"command": ["loadfile", "cdda://"]}`)
}

func (mpv *MPV) Play() error {
	return mpv.SendSuccessCommand(`{"command": ["set_property", "pause", false]}`)
}

func (mpv *MPV) Pause() error {
	return mpv.SendSuccessCommand(`{"command": ["set_property", "pause", true]}`)
}

func (mpv *MPV) GetTimePosition() (int, error) {
	response, err := mpv.SendCommand(`{"command": ["get_property", "time-pos"]}`)
	if err != nil {
		return 0, err
	}

	seconds, ok := response.Data.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected data type for time-pos: %T", response.Data)
	}

	return int(seconds * 1000), nil
}

func (mpv *MPV) IsPlaying() (bool, error) {
	response, err := mpv.SendCommand(`{"command": ["get_property", "pause"]}`)
	if err != nil {
		return false, err
	}

	paused, ok := response.Data.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected data type for pause: %T", response.Data)
	}

	return !paused, nil
}

func (mpv *MPV) SendSuccessCommand(command string) error {
	response, err := mpv.SendCommand(command)
	if err != nil {
		return err
	}

	if response.Error == "success" {
		return nil
	}

	return err
}

func (mpv *MPV) SendCommand(command string) (*MPVResponse, error) {
	_, err := mpv.conn.Write([]byte(command + "\n"))
	if err != nil {
		return nil, fmt.Errorf("error sending command to MPV: %v", err)
	}

	reader := bufio.NewReader(mpv.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading response from MPV: %v", err)
	}

	result := new(MPVResponse)
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %v", err)
	}

	return result, nil
}
