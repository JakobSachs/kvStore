package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	. "jakobsachs.blog/kvStore/shared"
)

const SERVER_URL = "http://localhost:8080/request"

func main() {
	req := Request{
		Type: NoOp,
		Id:   1,
	}

	jsonData, err := Serialize(req)
	if err != nil {
		log.Fatalf("Error marshalling request to JSON: %v", err)
	}

	httpReq, err := http.NewRequest("POST", SERVER_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("Error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Sent NoOp Request: %s %s\n", httpReq.Method, httpReq.URL)
	fmt.Printf("Response Status: %s\n", resp.Status)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Printf("Response Body: %s\n", string(responseBody))
}
