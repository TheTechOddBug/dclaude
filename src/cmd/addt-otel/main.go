// addt-otel is a simple OpenTelemetry collector that logs received telemetry data.
// It listens for OTLP HTTP requests and outputs them to stdout/file for debugging.
package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	port     = flag.Int("port", 4318, "Port to listen on")
	logFile  = flag.String("log", "", "Log file path (default: stdout)")
	verbose  = flag.Bool("verbose", false, "Verbose output (show full payloads)")
	jsonOut  = flag.Bool("json", false, "Output as JSON lines")
)

// Logger handles output formatting
type Logger struct {
	out     io.Writer
	jsonOut bool
	verbose bool
}

// LogEntry represents a log entry for JSON output
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Type      string                 `json:"type"`
	Count     int                    `json:"count,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func (l *Logger) log(telemetryType string, data map[string]interface{}, count int) {
	timestamp := time.Now().Format(time.RFC3339)

	if l.jsonOut {
		entry := LogEntry{
			Timestamp: timestamp,
			Type:      telemetryType,
			Count:     count,
		}
		if l.verbose {
			entry.Data = data
		}
		jsonBytes, _ := json.Marshal(entry)
		fmt.Fprintln(l.out, string(jsonBytes))
	} else {
		if l.verbose && data != nil {
			jsonBytes, _ := json.MarshalIndent(data, "", "  ")
			fmt.Fprintf(l.out, "[%s] %s (%d items):\n%s\n", timestamp, telemetryType, count, string(jsonBytes))
		} else {
			fmt.Fprintf(l.out, "[%s] %s: received %d items\n", timestamp, telemetryType, count)
		}
	}
}

func main() {
	flag.Parse()

	// Setup output
	var out io.Writer = os.Stdout
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		out = f
	}

	logger := &Logger{
		out:     out,
		jsonOut: *jsonOut,
		verbose: *verbose,
	}

	// Setup HTTP handlers
	http.HandleFunc("/v1/traces", makeHandler(logger, "traces"))
	http.HandleFunc("/v1/metrics", makeHandler(logger, "metrics"))
	http.HandleFunc("/v1/logs", makeHandler(logger, "logs"))
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", rootHandler)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Fprintf(os.Stderr, "addt-otel listening on %s\n", addr)
	fmt.Fprintf(os.Stderr, "Endpoints:\n")
	fmt.Fprintf(os.Stderr, "  POST /v1/traces  - Receive trace data\n")
	fmt.Fprintf(os.Stderr, "  POST /v1/metrics - Receive metrics data\n")
	fmt.Fprintf(os.Stderr, "  POST /v1/logs    - Receive log data\n")
	fmt.Fprintf(os.Stderr, "  GET  /health     - Health check\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func makeHandler(logger *Logger, telemetryType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read body, handling gzip compression
		var reader io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			reader = gzReader
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		// Parse based on content type
		contentType := r.Header.Get("Content-Type")
		var data map[string]interface{}
		count := 0

		if strings.Contains(contentType, "application/json") {
			if err := json.Unmarshal(body, &data); err == nil {
				count = countItems(data, telemetryType)
			}
		} else if strings.Contains(contentType, "application/x-protobuf") {
			// For protobuf, we just count bytes and note it's binary
			data = map[string]interface{}{
				"format": "protobuf",
				"bytes":  len(body),
			}
			count = 1 // We can't easily count items in protobuf without full parsing
		} else {
			// Try JSON anyway
			if err := json.Unmarshal(body, &data); err == nil {
				count = countItems(data, telemetryType)
			} else {
				data = map[string]interface{}{
					"format": "unknown",
					"bytes":  len(body),
				}
				count = 1
			}
		}

		logger.log(telemetryType, data, count)

		// Return success response (OTLP expects empty JSON object)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

func countItems(data map[string]interface{}, telemetryType string) int {
	count := 0
	var resourceKey, itemKey string

	switch telemetryType {
	case "traces":
		resourceKey = "resourceSpans"
		itemKey = "scopeSpans"
	case "metrics":
		resourceKey = "resourceMetrics"
		itemKey = "scopeMetrics"
	case "logs":
		resourceKey = "resourceLogs"
		itemKey = "scopeLogs"
	}

	if resources, ok := data[resourceKey].([]interface{}); ok {
		for _, res := range resources {
			if resMap, ok := res.(map[string]interface{}); ok {
				if items, ok := resMap[itemKey].([]interface{}); ok {
					count += len(items)
				}
			}
		}
	}

	if count == 0 {
		count = 1 // At least we received something
	}
	return count
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "addt-otel collector\n\n")
	fmt.Fprintf(w, "Endpoints:\n")
	fmt.Fprintf(w, "  POST /v1/traces  - Receive trace data\n")
	fmt.Fprintf(w, "  POST /v1/metrics - Receive metrics data\n")
	fmt.Fprintf(w, "  POST /v1/logs    - Receive log data\n")
	fmt.Fprintf(w, "  GET  /health     - Health check\n")
}
