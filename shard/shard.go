// This is the shard service
// It take's in requests (eiter read or write) for keys, and acts
// on these.

package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	. "jakobsachs.blog/kvStore/shared"
)

// TODO: make this less shitty
var store map[string]string
var storeMtx sync.Mutex

func readHandler(r Request) (string, error) {
	storeMtx.Lock()
	defer storeMtx.Unlock()

	return store[r.Key], nil
}

func writeHandler(r Request) (string, error) {
	storeMtx.Lock()
	defer storeMtx.Unlock()

	store[r.Key] = r.Value

	return r.Value, nil
}

// handles entire request parsing etc
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "ONLY POST ALLOWED", 404)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("ERROR: failed to read request-body: %v\n", err)
		http.Error(w, "failed to read request-body", 500)
		return
	}
	fmt.Printf("Received raw request body: %s\n", body)

	req, err := Deserialize(body)
	if err != nil {
		fmt.Printf("ERROR: failed to parse request: %v\n", err)
		http.Error(w, "failed to parse request", 404)
		return
	}
	fmt.Printf("Parsed request: %+v\n", req)

	if req.Type == NoOp {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		io.WriteString(w, "ping")
		return
	}

	// route READ
	var resp string
	if req.Type == Read {
		resp, err = readHandler(req)
	} else if req.Type == Write {
		resp, err = writeHandler(req)
	} else {
    // haxor ?
		fmt.Println("ERROR: Unhandled request type")
		http.Error(w, "invalid request type", 404)
	}

	if err != nil {
		fmt.Printf("ERROR: failed to serve request: %v\n", err)
		http.Error(w, "failed to service request", 404)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, resp)

}

func main() {
	store = make(map[string]string)
	http.HandleFunc("/", handler)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Could not start server: %s\n", err)
	}
}
