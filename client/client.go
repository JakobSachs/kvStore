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

func sendRequestToServer(req Request) {
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

	// Consider making the printing more generic or returning values if this is a library function
	fmt.Printf("Sent Request (Type: %v, Id: %v): %s %s\n", req.Type, req.Id, httpReq.Method, httpReq.URL)
	fmt.Printf("Response Status: %s\n", resp.Status)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Printf("Response Body: %s\n", string(responseBody))
}

func main() {
	req := Request{
		Type: NoOp,
		Id:   1,
	}
	sendRequestToServer(req)

  req.Type = Write
  req.Id += 1
  req.Key = "Beep boop"
  req.Value = "boop beep ?"
	sendRequestToServer(req)

  req.Type = Read
  req.Id += 1
  req.Key = "Beep boop"
  req.Value = ""
	sendRequestToServer(req)


}
