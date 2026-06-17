// Command client is the load generator of the talk. It reuses a SINGLE
// persistent HTTP/2 (h2c) connection and loops requests against the go-api
// Service, printing which pod (UUID) answered each one.
//
// The single reused connection is the crux of the demo:
//   - no-mesh: kube-proxy (L4) pins that one connection to one pod, so every
//     request shows the SAME UUID — load is NOT balanced.
//   - mesh: the Envoy sidecar reads the HTTP/2 frames (L7) and balances each
//     request across pods per the VirtualService weights (60/20/20).
//
// It targets the in-cluster Service DNS directly (ClusterIP). Do NOT use
// kubectl port-forward for this traffic — port-forward pins a single pod and
// masks the real kube-proxy behavior.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/http2"
)

type healthResponse struct {
	Pod     string `json:"pod"`
	Message string `json:"message"`
}

func main() {
	target := getenv("SERVER_URL", "http://go-api:8080/api/v1/health")
	interval := getDuration("REQUEST_INTERVAL", 500*time.Millisecond)

	// A single http2.Transport in h2c mode. AllowHTTP + a DialTLSContext that
	// returns a plain TCP connection gives us HTTP/2 cleartext with "prior
	// knowledge" — no TLS, no upgrade dance. Reusing one *http.Client means one
	// connection, multiplexing every request. This is what concentrates load on
	// a single pod when there is no mesh.
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	log.Printf("go-client starting | target=%s | interval=%s", target, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		doRequest(client, target)
	}
}

func doRequest(client *http.Client, target string) {
	resp, err := client.Get(target)
	if err != nil {
		log.Printf("[ERR] %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERR] reading body: %v", err)
		return
	}

	var hr healthResponse
	if err := json.Unmarshal(body, &hr); err != nil {
		log.Printf("[%d] %s", resp.StatusCode, string(body))
		return
	}

	// Match the slides' on-screen format: "[200 OK] Responding from Pod: <uuid>".
	log.Printf("[%d %s] Responding from Pod: %s", resp.StatusCode, http.StatusText(resp.StatusCode), shorten(hr.Pod))
}

// shorten trims the UUID to the short form shown on the slides (e.g. a1b2c),
// which stays unique across the three pods while being easy to read live.
func shorten(id string) string {
	if len(id) >= 5 {
		return id[:5]
	}
	return id
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if ms, err := strconv.Atoi(v); err == nil {
		return time.Duration(ms) * time.Millisecond
	}
	return fallback
}
