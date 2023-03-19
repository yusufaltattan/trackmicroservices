package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	port     = ":3001"
	url      = "https://api.audd.io/recognize"
	apiToken = "114ffe7ad6cdf435bba4937395473f6d"
)

type requestBody struct {
	Audio string `json:"audio"`
}

type response struct {
	Status string `json:"status"`
	Result struct {
		Id string `json:"title"`
	} `json:"result"`
}

func main() {
	http.HandleFunc("/search", searchHandler)
	fmt.Printf("Listening on port %s\n", port)
	http.ListenAndServe(port, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var reqBody requestBody
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	jsonBytes, err := json.Marshal(map[string]interface{}{
		"api_token": apiToken,
		"audio":     reqBody.Audio,
	})
	if err != nil {
		http.Error(w, "Failed to marshal request body", http.StatusInternalServerError)
		return
	}

	sendData := bytes.NewBuffer(jsonBytes)

	resp, err := http.Post(url, "application/json", sendData)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	//LOG
	fmt.Println(string(responseBody))

	var respBody response
	err = json.Unmarshal(responseBody, &respBody)
	if err != nil {
		http.Error(w, "Failed to parse response body", http.StatusInternalServerError)
		return
	}

	// Create a new response object containing only the title field
	titleResponse := struct {
		Title string `json:"Id"`
	}{
		Title: respBody.Result.Id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(titleResponse)
}
