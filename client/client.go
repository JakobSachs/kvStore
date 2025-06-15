package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	. "jakobsachs.blog/kvStore/shared"
)

type Client struct {
}

func main() {
	// Create a NoOp request
	req := Request{
		Type: NoOp,
		Id:   1, // Example Id
		// Key and Value are nil for NoOp
	}

	// Marshal the request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Error marshalling request to JSON: %v", err)
	}

	// Define the server URL (replace with your actual server URL)
	serverURL := "http://localhost:8080/request" // Placeholder URL

	// Create a new HTTP POST request
	httpReq, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request using the default client
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("Error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Print the response status
	fmt.Printf("Sent NoOp Request: %s %s\n", httpReq.Method, httpReq.URL)
	fmt.Printf("Response Status: %s\n", resp.Status)

	// You can also read and print the response body if needed
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// log.Fatalf("Error reading response body: %v", err)
	// }
	// fmt.Printf("Response Body: %s\n", string(responseBody))
}
