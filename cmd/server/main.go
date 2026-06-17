// Command server is the "go-api" of the talk: a tiny HTTP/2 cleartext (h2c)
// service that generates a UUID at startup (the pod identity) and returns it on
// every response. Three replicas of this (v1/v2/v3) let the audience see which
// pod answered each request — the whole point of Arcos 4 and 7.
//
// It speaks h2c (HTTP/2 without TLS) on purpose: cleartext is what lets the
// Envoy sidecar read the HTTP/2 frames and load-balance per request in the mesh
// scenario. There is no TLS anywhere in this project.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// podID is generated once at startup and is the canonical identity of this
// instance. Every response carries it so the client can tell pods apart.
var podID = uuid.NewString()

type healthResponse struct {
	Pod     string `json:"pod"`
	Message string `json:"message"`
}

func main() {
	port := getenv("PORT", "8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", healthHandler)
	// Be forgiving: any other path also reports the pod identity.
	mux.HandleFunc("/", healthHandler)

	// h2c.NewHandler upgrades the plain handler to serve HTTP/2 cleartext, so a
	// client using HTTP/2 prior knowledge gets a real multiplexed h2 connection
	// with no TLS handshake.
	handler := h2c.NewHandler(mux, &http2.Server{})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("go-api starting | pod=%s | h2c on :%s", podID, port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{
		Pod:     podID,
		Message: fmt.Sprintf("Hello from %s", podID),
	})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
