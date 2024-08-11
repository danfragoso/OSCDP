package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Disc struct {
	Artist string   `json:"artist"`
	Title  string   `json:"title"`
	Tracks []*Track `json:"tracks"`

	Size int64 `json:"size"`
}

func monitorDiscSize(s chan int64) {
	for {
		size, _ := getDiscSize()
		if size > 0 {
			s <- size
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func getDiscSize() (int64, error) {
	cmd := exec.Command("blockdev", "--getsize64", CDDevice)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	sizeStr := strings.TrimSpace(string(output))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func createDisc(TOC string, size int64) (*Disc, error) {
	parts := strings.Split(TOC, " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid TOC format")
	}

	firstTrack, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid first track number: %v", err)
	}

	lastTrack, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid last track number: %v", err)
	}

	endOfDisc, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid end of disc value: %v", err)
	}

	startFrames := parts[3:]
	numTracks := lastTrack - firstTrack + 1

	if len(startFrames) != numTracks {
		return nil, fmt.Errorf("TOC length mismatch, expected %d frames but got %d", numTracks, len(startFrames))
	}

	tracks := make([]*Track, numTracks)
	for i := 0; i < numTracks; i++ {
		beginFrame, err := strconv.Atoi(startFrames[i])
		if err != nil {
			return nil, fmt.Errorf("invalid begin frame for track %d: %v", i+1, err)
		}

		var endFrame int
		if i < numTracks-1 {
			nextBeginFrame, err := strconv.Atoi(startFrames[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid next begin frame for track %d: %v", i+2, err)
			}
			endFrame = nextBeginFrame
		} else {
			endFrame = endOfDisc
		}

		length := (endFrame - beginFrame) * 1000 / 75

		tracks[i] = &Track{
			Title:  fmt.Sprintf("Track %d", firstTrack+i),
			Number: strconv.Itoa(firstTrack + i),
			Offset: beginFrame * 1000 / 75,
			Length: length,
		}
	}

	disc := &Disc{
		Artist: "Unknown Artist",
		Title:  "Unknown Album",
		Tracks: tracks,
		Size:   size,
	}

	return disc, nil
}

func createAndIdentifyDisk(size int64) (*Disc, error) {
	discID, TOC, err := getDiscIDAndTOC()
	if err != nil {
		return nil, err
	}

	disc, err := createDisc(TOC, size)
	if err != nil {
		return nil, err
	}

	discInfo, err := getDiscInfo(discID)
	if err != nil {
		fmt.Printf("failed to get disc info: %v\n", err)
		return disc, nil
	}

	if discInfo != nil && len(discInfo.Releases) > 0 {
		disc.Title = discInfo.Releases[0].Title
		if len(discInfo.Releases[0].ArtistCredit) > 0 {
			disc.Artist = discInfo.Releases[0].ArtistCredit[0].Name
		}

		if len(discInfo.Releases[0].Media) > 0 {
			for i, track := range discInfo.Releases[0].Media[0].Tracks {
				if i < len(disc.Tracks) {
					disc.Tracks[i].Title = track.Title
				}
			}
		}
	}

	return disc, nil
}
