package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uploadedlobster.com/discid"
)

const musicBrainzURL = "https://musicbrainz.org/ws/2/"

type DiscIDResponse struct {
	ID       string `json:"id"`
	Releases []struct {
		Title        string `json:"title"`
		ArtistCredit []struct {
			Name string `json:"name"`
		} `json:"artist-credit"`
		Media []struct {
			Tracks []struct {
				Title  string `json:"title"`
				Number string `json:"number"`
			} `json:"tracks"`
		} `json:"media"`
	} `json:"releases"`
}

func getDiscInfo(discID string) (*DiscIDResponse, error) {
	requestURL := fmt.Sprintf("%sdiscid/%s?inc=recordings+artists&fmt=json", musicBrainzURL, url.PathEscape(discID))

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "OSCDP/v0.1 ( danilo.fragoso@gmail.com )")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded, try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var discResponse DiscIDResponse
	err = json.Unmarshal(bodyBytes, &discResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &discResponse, nil
}

var (
	CDDevice = "/dev/sr0"
)

func getDiscIDAndTOC() (string, string, error) {
	disc, err := discid.ReadFeatures(CDDevice, discid.FeatureRead|discid.FeatureMCN)
	if err != nil {
		return "", "", err
	}
	defer disc.Close()
	return disc.ID(), disc.TOCString(), nil
}
