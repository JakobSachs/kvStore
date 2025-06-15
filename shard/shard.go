// This is the shard service
// It take's in requests (eiter read or write) for keys, and acts
// on these.

package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"

	. "jakobsachs.blog/kvStore/shared"
)

// TODO: make this less shitty
var store map[string]string
var storeMtx sync.Mutex

func readHandler(r Request) (string, error) {
	slog.Debug("Entering readHandler", "request_id", r.Id, "key", r.Key)
	storeMtx.Lock()
	defer storeMtx.Unlock()

	value, ok := store[r.Key]
	if !ok {
		slog.Debug("Key not found in store", "request_id", r.Id, "key", r.Key)
		return "", nil
	}
	slog.Debug("Key found in store", "request_id", r.Id, "key", r.Key, "value_length", len(value))
	return value, nil
}

func writeHandler(r Request) (string, error) {
	slog.Debug("Entering writeHandler", "request_id", r.Id, "key", r.Key, "value_length", len(r.Value))
	storeMtx.Lock()
	defer storeMtx.Unlock()

	store[r.Key] = r.Value
	slog.Debug("Key successfully written to store", "request_id", r.Id, "key", r.Key, "value_length", len(r.Value))

	return r.Value, nil
}

// handles entire request parsing etc
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Received non-POST request", "method", r.Method, "remote_addr", r.RemoteAddr)
		http.Error(w, "ONLY POST ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err, "remote_addr", r.RemoteAddr)
		http.Error(w, "failed to read request-body", http.StatusInternalServerError)
		return
	}
	slog.Debug("Received raw request body", "body", string(body), "remote_addr", r.RemoteAddr)

	req, err := Deserialize(body)
	if err != nil {
		slog.Error("Failed to parse request", "error", err, "raw_body", string(body), "remote_addr", r.RemoteAddr)
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		return
	}
	slog.Info("Parsed request", "request_id", req.Id, "request_type", req.Type, "key", req.Key, "remote_addr", r.RemoteAddr)

	if req.Type == NoOp {
		slog.Info("Handling NoOp request", "request_id", req.Id, "remote_addr", r.RemoteAddr)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		io.WriteString(w, "ping")
		return
	}

	// route READ/WRITE
	var resp string
	if req.Type == Read {
		slog.Info("Handling Read request", "request_id", req.Id, "key", req.Key, "remote_addr", r.RemoteAddr)
		resp, err = readHandler(req)
	} else if req.Type == Write {
		slog.Info("Handling Write request", "request_id", req.Id, "key", req.Key, "value_length", len(req.Value), "remote_addr", r.RemoteAddr)
		resp, err = writeHandler(req)
	} else {
    // haxor ?
		slog.Error("Unhandled request type", "request_id", req.Id, "request_type", req.Type, "remote_addr", r.RemoteAddr)
		http.Error(w, "invalid request type", http.StatusBadRequest)
		return 
	}

	if err != nil {
		slog.Error("Failed to serve request", "error", err, "request_id", req.Id, "request_type", req.Type, "remote_addr", r.RemoteAddr)
		http.Error(w, "failed to service request", http.StatusInternalServerError)
		return
	}

	slog.Info("Successfully served request", "request_id", req.Id, "response_length", len(resp), "remote_addr", r.RemoteAddr)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, resp)

}

func main() {
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "1" {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	var logHandler slog.Handler
	if os.Getenv("CLI") == "1" {
		logHandler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, opts)
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	if os.Getenv("PROFILE") == "1" {
		slog.Info("Profiling enabled. Access /debug/pprof/")
		// Register pprof handlers to the DefaultServeMux.
		// http.HandleFunc already uses DefaultServeMux.
		http.HandleFunc("/debug/pprof/", pprof.Index)
		http.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		http.HandleFunc("/debug/pprof/profile", pprof.Profile)
		http.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		http.HandleFunc("/debug/pprof/trace", pprof.Trace)
		// You can add other pprof handlers as needed (e.g., for heap, goroutine, etc.)
		// For example, to serve all standard pprof profiles:
		// http.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		// http.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		// http.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		// http.Handle("/debug/pprof/block", pprof.Handler("block"))
	}

	store = make(map[string]string)
	http.HandleFunc("/", handler)

	slog.Info("Starting server", "address", ":8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("Could not start server", "error", err)
	}
}
