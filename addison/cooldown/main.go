package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type AudioRequest struct {
	Audio string `json:"Audio"`
}

type SearchResponse struct {
	Id string `json:"Id"`
}

type trackResponse struct {
	Id    string `json:"Id"`
	Audio string `json:"Audio"`
}

func searchTrack(audio string) (string, error) {
	searchBody, err := json.Marshal(AudioRequest{audio})
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:3001/search", "application/json", bytes.NewBuffer(searchBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("searchTrack returned status %d", resp.StatusCode)
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return "", err
	}

	return searchResponse.Id, nil
}

func getTrack(id string) (*trackResponse, error) {

	id = strings.ReplaceAll(id, " ", "+")
	id = strings.ReplaceAll(id, "'", "")

	log.Printf("id after +: %s", string(id))

	trackURL := fmt.Sprintf("http://localhost:3000/tracks/%s", id)

	resp, err := http.Get(trackURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getTrack returned status %d", resp.StatusCode)
	}

	var trackResponseData trackResponse

	if err := json.NewDecoder(resp.Body).Decode(&trackResponseData); err != nil {
		return nil, err
	}

	return &trackResponseData, nil
}

func coolTownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req AudioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	id, err := searchTrack(req.Audio)
	if err != nil {
		http.Error(w, "Failed to search for track", http.StatusInternalServerError)
		return
	}
	log.Printf("id from audio: %s", string(id))

	trackData, err := getTrack(id)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}
	log.Printf("id from tracks: %s", trackData.Id)

	responseData := map[string]string{
		"Audio": trackData.Audio,
		"Id":    trackData.Id,
	}

	responseBody, err := json.Marshal(responseData)
	if err != nil {
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func main() {
	http.HandleFunc("/cooltown", coolTownHandler)

	log.Fatal(http.ListenAndServe(":3002", nil))
}
