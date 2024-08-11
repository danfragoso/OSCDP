package main

import (
	"image/color"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/freesans"
)

var ( // LCD
	LCD_DC_PIN  = machine.GP8
	LCD_CS_PIN  = machine.GP9
	LCD_CLK_PIN = machine.GP10
	LCD_DIN_PIN = machine.GP11
	LCD_RST_PIN = machine.GP12
	LCD_BL_PIN  = machine.GP13

	LCD_WIDTH  = 240
	LCD_HEIGHT = 240

	display st7789.Device
)

func initDisplay() {
	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 0,
		SCK:       LCD_CLK_PIN,
		SDO:       LCD_DIN_PIN,
		SDI:       LCD_DC_PIN,
		Mode:      0,
	})

	display = st7789.New(machine.SPI1,
		LCD_RST_PIN,
		LCD_DC_PIN,
		LCD_CS_PIN,
		LCD_BL_PIN)

	display.Configure(st7789.Config{
		Rotation:   st7789.NO_ROTATION,
		RowOffset:  80,
		FrameRate:  st7789.FRAMERATE_111,
		VSyncLines: st7789.MAX_VSYNC_SCANLINES,
	})

	display.FillScreen(color.RGBA{0, 0, 0, 255})
}

type DisplayCommand struct {
	Section string `json:"section"`
	Content string `json:"content"`
}

func listenDisplayCommands(displayCommands chan *DisplayCommand) {
	var msgBuffer []byte

	for {
		chr, err := machine.Serial.ReadByte()
		if err == nil {
			msgBuffer = append(msgBuffer[:], chr)
			if chr == '\r' {
				info := strings.Split(string(msgBuffer), "|")
				if len(info) != 2 {
					msgBuffer = nil
					continue
				}

				displayCommands <- &DisplayCommand{
					Section: info[0],
					Content: info[1],
				}

				msgBuffer = nil
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
}

type DisplayState struct {
	Track        string
	Artist       string
	Album        string
	Time         string
	PlayerStatus string
	IPAddr       string
}

var displayState = &DisplayState{}

func renderDisplayCommand(section string, content string) {
	switch section {
	case "track":
		if displayState.Track != content {
			displayState.Track = content
			clearAndRenderTrack(content)
		}

	case "artist":
		if displayState.Artist != content {
			displayState.Artist = content
			clearAndRenderArtist(content)
		}

	case "album":
		if displayState.Album != content {
			displayState.Album = content
			clearAndRenderAlbum(content)
		}

	case "time":
		if displayState.Time != content {
			displayState.Time = content
			clearAndRenderTime(content)
		}

	case "player_status":
		if displayState.PlayerStatus != content {
			displayState.PlayerStatus = content
			displayHeaderWithInfo(content)
		}
	}
}

func clearAndRenderTrack(track string) {
	display.FillRectangle(0, 40, 240, 36, color.RGBA{0, 0, 0, 255})
	tinyfont.WriteLine(&display, &freesans.Regular12pt7b, 12, 64, track, color.RGBA{255, 255, 255, 255})
}

func clearAndRenderArtist(artist string) {
	display.FillRectangle(0, 76, 240, 36, color.RGBA{0, 0, 0, 255})
	tinyfont.WriteLine(&display, &freesans.Regular12pt7b, 12, 100, artist, color.RGBA{255, 255, 255, 255})
}

func clearAndRenderAlbum(album string) {
	display.FillRectangle(0, 112, 240, 36, color.RGBA{0, 0, 0, 255})
	tinyfont.WriteLine(&display, &freesans.Regular12pt7b, 12, 136, album, color.RGBA{255, 255, 255, 255})
}

func clearAndRenderTime(time string) {
	display.FillRectangle(0, 154, 240, 36, color.RGBA{0, 0, 0, 255})
	_, outboxWidth := tinyfont.LineWidth(&freemono.Bold12pt7b, time)
	tinyfont.WriteLine(&display, &freemono.Bold12pt7b, (240-int16(outboxWidth))/2, 178, time, color.RGBA{255, 255, 255, 255})
}

func displayHeaderWithInfo(info string) {
	display.FillRectangle(0, 0, 240, 30, color.RGBA{255, 255, 255, 255})
	tinyfont.WriteLine(&display, &freesans.Bold12pt7b, 14, 22, "OSCDP", color.RGBA{0, 0, 0, 255})

	_, outboxWidth := tinyfont.LineWidth(&freesans.Regular9pt7b, info)
	tinyfont.WriteLine(&display, &freesans.Regular9pt7b, (236-int16(outboxWidth))/1, 20, info, color.RGBA{0, 0, 0, 255})
}

func displayHeaderWithGlyph(glyph string) {
	display.FillRectangle(0, 0, 240, 30, color.RGBA{255, 255, 255, 255})
	tinyfont.WriteLine(&display, &freesans.Bold12pt7b, 14, 22, "OSCDP", color.RGBA{0, 0, 0, 255})

	_, outboxWidth := tinyfont.LineWidth(&MediaFont18, glyph)
	tinyfont.WriteLine(&display, &MediaFont18, (236-int16(outboxWidth))/1, 20, glyph, color.RGBA{0, 0, 0, 255})
}

func clearAndRenderButtonCues() {
	cue := "⏏ ⏮ ⏭ ⏯"
	_, outboxWidth := tinyfont.LineWidth(&MediaFont22, cue)
	display.FillRectangle(0, 200, 240, 40, color.RGBA{0, 0, 0, 255})
	tinyfont.WriteLineColors(&display, &MediaFont22, (240-int16(outboxWidth))/2, 224, cue, []color.RGBA{
		color.RGBA{0, 255, 255, 255}, // ⏏ Bright Cyan
		color.RGBA{0, 0, 0, 255},     // Whitespace
		color.RGBA{255, 255, 0, 255}, // ⏮ Bright Yellow
		color.RGBA{0, 0, 0, 255},     // Whitespace
		color.RGBA{255, 0, 255, 255}, // ⏭ Bright Magenta
		color.RGBA{0, 0, 0, 255},     // Whitespace
		color.RGBA{0, 255, 0, 255},   // ⏯ Bright Green
	})
}

//album|Unknown Album
//artist|Unknown Artist
//track|3. Unknown Track
//time|[00:00/04:00]
//player_status|Playing
