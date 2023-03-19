package main

// These are the imports
import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Struct for the request
type Track struct {
	Id    string `json:"Id"`
	Audio string `json:"Audio"`
}

func main() {
	http.HandleFunc("/tracks/", handleTrack)
	http.HandleFunc("/tracks", handleListTracks)

	log.Println("Tracks microservice started on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func trackExists(id string) bool {
	_, err := os.Stat(filepath.Join("tracks", id+".wav"))
	return err == nil
}

func getTrack(id string) (string, error) {
	filePath := filepath.Join("tracks", id+".wav")
	audioBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	audioBase64 := base64.StdEncoding.EncodeToString(audioBytes)
	return audioBase64, nil
}

func getTrackIds() ([]string, error) {
	var ids []string
	files, err := ioutil.ReadDir("tracks")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if ext := filepath.Ext(name); ext == ".wav" {
				ids = append(ids, name[:len(name)-len(ext)])
			}
		}
	}
	return ids, nil
}

func handleListTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	trackIds, err := getTrackIds()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Failed to get track IDs: %v", err)
		return
	}

	response := struct {
		TrackIds []string `json:"trackIds"`
	}{
		TrackIds: trackIds,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Failed to encode track IDs response: %v", err)
		return
	}
}

func handleTrack(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		if id == "" {
			// GET /tracks
			ids, err := getTrackIds()
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Failed to get track ids: %v", err)
				return
			}

			response := make(map[string][]string)
			response["ids"] = ids

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Failed to encode track ids response: %v", err)
				return
			}
		} else {
			// GET /tracks/id
			track, err := getTrack(id)
			if err != nil {
				if os.IsNotExist(err) {
					http.Error(w, "Track not found", http.StatusNotFound)
				} else {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					log.Printf("Failed to get track: %v", err)
				}
				return
			}

			response := Track{Id: id, Audio: track}

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Failed to encode track response: %v", err)
				return
			}
		}
	case http.MethodPut:
		// PUT /tracks/id
		var track Track
		err := json.NewDecoder(r.Body).Decode(&track)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id = strings.ReplaceAll(id, "'", "")
		if track.Id != id {
			http.Error(w, "Id in URL does not match Id in request body", http.StatusBadRequest)
			return
		}

		if trackExists(track.Id) {
			http.Error(w, "Track with ID "+track.Id+" already exists", http.StatusBadRequest)
			return
		}

		audioBytes, err := base64.StdEncoding.DecodeString(track.Audio)
		if err != nil {
			http.Error(w, "invalid base64 encoding", http.StatusBadRequest)
			return
		}

		fileName := track.Id + ".wav"
		filePath := filepath.Join("tracks", fileName)
		err = ioutil.WriteFile(filePath, audioBytes, 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case http.MethodDelete:
		// DELETE /tracks/id
		id = strings.ReplaceAll(id, "'", "")
		if !trackExists(id) {
			http.Error(w, "Track not found", http.StatusNotFound)
			return
		}

		filePath := filepath.Join("tracks", id+".wav")
		err := os.Remove(filePath)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Failed to delete track: %v", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
