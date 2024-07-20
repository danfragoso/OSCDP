package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func getTrackOffsets() ([]int, error) {
	cmd := exec.Command("cdparanoia", "-Q")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing cdparanoia: %v", err)
	}

	var trackOffsets []int
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	trackPattern := regexp.MustCompile(`^\s*\d+\.\s+\d+\s+\[\d+:\d+.\d+\]\s+(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		match := trackPattern.FindStringSubmatch(line)
		if match != nil {
			offset, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing track offset: %v", err)
			}
			trackOffsets = append(trackOffsets, offset)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading cdparanoia output: %v", err)
	}

	return trackOffsets, nil
}

func calculateChecksum(trackOffsets []int) int {
	checksum := 0
	for _, offset := range trackOffsets {
		for offset > 0 {
			checksum += offset % 10
			offset /= 10
		}
	}
	return checksum
}

func calculateDiscId(trackOffsets []int) string {
	numTracks := len(trackOffsets)
	checksum := calculateChecksum(trackOffsets)
	leadOut := trackOffsets[len(trackOffsets)-1] + 1 // Simplified assumption, adjust as needed

	discId := fmt.Sprintf("%02x-%06x %06x", numTracks, checksum, leadOut)
	return strings.ToUpper(discId)
}

func main() {
	trackOffsets, err := getTrackOffsets()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(trackOffsets) == 0 {
		fmt.Println("Error: No tracks found.")
		return
	}

	discId := calculateDiscId(trackOffsets)
	fmt.Println("DiscID:", discId)
}
