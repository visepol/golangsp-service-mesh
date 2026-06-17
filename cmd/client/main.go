// Command client is the load generator of the talk — themed as Brazil scoring
// goals. It reuses a SINGLE persistent HTTP/2 (h2c) connection and loops requests
// against the go-api Service; each response is a goal against the opponent
// (server pod) that answered. It keeps a running scoreboard so the audience can
// read the load distribution as a scoreline.
//
// The single reused connection is the crux of the demo:
//   - no-mesh: kube-proxy (L4) pins that one connection to one pod, so every
//     goal goes against the SAME team — load is NOT balanced (a goleada).
//   - mesh: the Envoy sidecar reads the HTTP/2 frames (L7) and spreads goals
//     across teams per the VirtualService weights (60/20/20).
//
// It targets the in-cluster Service DNS directly (ClusterIP). Do NOT use
// kubectl port-forward for this traffic — port-forward pins a single pod and
// masks the real kube-proxy behavior.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

type healthResponse struct {
	Pod     string `json:"pod"`
	Team    string `json:"team"`
	Message string `json:"message"`
}

// scoreboard tallies goals per opponent, preserving first-seen order so the
// printed line stays stable (no jitter) during the live demo.
type scoreboard struct {
	order []string
	goals map[string]int
}

func newScoreboard() *scoreboard { return &scoreboard{goals: map[string]int{}} }

func (s *scoreboard) add(team string) {
	if _, ok := s.goals[team]; !ok {
		s.order = append(s.order, team)
	}
	s.goals[team]++
}

func (s *scoreboard) String() string {
	parts := make([]string, 0, len(s.order))
	for _, t := range s.order {
		parts = append(parts, fmt.Sprintf("%s %d", t, s.goals[t]))
	}
	return strings.Join(parts, "  ·  ")
}

func main() {
	target := getenv("SERVER_URL", "http://go-api:8080/api/v1/health")
	interval := getDuration("REQUEST_INTERVAL", 500*time.Millisecond)
	myTeam := getenv("CLIENT_TEAM", "Brasil")

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

	board := newScoreboard()
	log.Printf("⚽ %s entra em campo | alvo=%s | intervalo=%s", myTeam, target, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		doRequest(client, target, myTeam, board)
	}
}

func doRequest(client *http.Client, target, myTeam string, board *scoreboard) {
	resp, err := client.Get(target)
	if err != nil {
		log.Printf("[ERR] %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERR] lendo resposta: %v", err)
		return
	}

	var hr healthResponse
	if err := json.Unmarshal(body, &hr); err != nil {
		log.Printf("[%d] %s", resp.StatusCode, string(body))
		return
	}

	opponent := hr.Team
	if opponent == "" {
		opponent = shorten(hr.Pod)
	}
	board.add(opponent)

	// "⚽ GOL! Brasil x Alemanha  →  placar: Alemanha 7  ·  Argentina 2  ·  França 2"
	log.Printf("⚽ GOL! %s x %s  →  placar: %s", myTeam, opponent, board)
}

// shorten trims the UUID to the short form (e.g. a1b2c) — fallback identity when
// a pod has no TEAM set, so the demo still works without the theme.
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
