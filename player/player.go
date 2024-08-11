package main

import (
	"fmt"
	"os/exec"
)

type Player struct {
	Disc *Disc
	MPV  *MPV

	Position int // in ms
	Status   string
}

func ejectDisc() error {
	return exec.Command("eject", "/dev/cdrom").Run()
}

func (p *Player) StartDisc() error {
	err := p.MPV.StartDisc()
	if err != nil {
		return err
	}

	p.Status = "Playing"

	return nil
}

func (p *Player) PlayPause() {
	if p.Status == "Playing" {
		if p.MPV.Pause() == nil {
			p.Status = "Paused"
		}
	} else {
		if p.MPV.Play() == nil {
			p.Status = "Playing"
		}
	}
}

func (p *Player) PreviousTrack() {
	p.MPV.PreviousTrack()
}

func (p *Player) NextTrack() {
	p.MPV.NextTrack()
}

func (p *Player) HandleKey(key string) {
	switch key {
	case "Play/Pause":
		p.PlayPause()
	case "Prev":
		p.PreviousTrack()
	case "Next":
		p.NextTrack()
	case "Eject":
		p.EjectDisc()
	}
}

func (p *Player) Reset() {
	p.MPV.Stop()
	p.Disc = nil
	p.Position = 0
	p.Status = "Stopped"
}

func (p *Player) EjectDisc() error {
	p.Reset()
	return ejectDisc()
}

func (p *Player) GetCurrentTrack() *Track {
	if p.Disc == nil {
		return nil
	}

	p.UpdatePosition()

	for _, track := range p.Disc.Tracks {
		if p.Position >= track.Offset-2500 && p.Position <= track.Offset+track.Length-2500 {
			return track
		}
	}

	return nil
}

func (p *Player) UpdatePosition() {
	if p.Disc == nil {
		return
	}

	pos, err := p.MPV.GetTimePosition()
	if err != nil {
		return
	}

	p.Position = pos
}

func (p *Player) UpdateStatus() {
	playing, err := p.MPV.IsPlaying()
	if err != nil {
		return
	}

	if playing {
		p.Status = "Playing"
	} else {
		p.Status = "Paused"
	}
}

func (p *Player) GetPrettyPosition() string {
	if p.Disc == nil {
		return "00:00/00:00"
	}

	track := p.GetCurrentTrack()
	if track == nil {
		return "00:00/00:00"
	}

	trackPos := p.Position - track.Offset
	minutes := trackPos / 60000
	seconds := (trackPos / 1000) % 60
	if seconds < 0 {
		seconds = 0
	}

	trackMinutes := track.Length / 60000
	trackSeconds := (track.Length / 1000) % 60

	return padLeft(minutes, 2) + ":" + padLeft(seconds, 2) + "/" + padLeft(trackMinutes, 2) + ":" + padLeft(trackSeconds, 2)
}

func padLeft(n int, width int) string {
	return fmt.Sprintf("%0*d", width, n)
}

func (p *Player) UpdateController(c *Controller) {
	if p.Disc == nil {
		c.WriteCommand(`player_status|No Disc`)
		c.WriteCommand(`time|`)
		c.WriteCommand(`album|`)
		c.WriteCommand(`artist|`)
		c.WriteCommand(`track|`)
		return
	} else {
		c.WriteCommand(`player_status|` + p.Status)
		c.WriteCommand(`album|` + p.Disc.Title)
		c.WriteCommand(`artist|` + p.Disc.Artist)

		c.WriteCommand(`time|` + p.GetPrettyPosition())

		track := p.GetCurrentTrack()
		if track != nil {
			c.WriteCommand(`track|` + track.Number + ". " + track.Title)
		}
	}
}

func InitPlayer(mpv *MPV) *Player {
	return &Player{
		Disc: nil,
		MPV:  mpv,
	}
}
