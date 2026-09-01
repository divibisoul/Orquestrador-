package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type request struct {
	Operation string            `json:"operation"`
	Payload   []float64         `json:"payload"`
	Metadata  map[string]string `json:"metadata"`
}

func main() {
	n, err := neural.New(8, .05)
	if err != nil {
		log.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		log.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, e.Health()) })
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, e.Stats()) })
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { writeMetrics(w, e.Stats()) })
	mux.Handle("/api/soul-mesh", mesh.NewHTTPGateway(e))
	mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
			return
		}
		var req request
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if req.Operation == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "operation is required"})
			return
		}
		result, err := e.Execute(r.Context(), req.Operation, req.Payload, req.Metadata)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, result)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	addr := os.Getenv("N07_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if registry := strings.TrimSpace(os.Getenv("N07_MESH_REGISTRY_URL")); registry != "" {
		go func() {
			a, err := mesh.New(registry)
			if err != nil {
				log.Printf("mesh adapter init failed: %v", err)
				return
			}
			regCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if _, err := a.Register(regCtx, registry, []string{"neural.forward", "neural.learn", "compute.execute", "cognitive.execute", "mesh.ping", "mesh.describe"}); err != nil {
				log.Printf("mesh registration failed: %v", err)
			}
		}()
	}
	go func() {
		log.Printf("N07 Orquestrador listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = e.Shutdown(shutdown)
	_ = srv.Shutdown(shutdown)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeMetrics(w http.ResponseWriter, s map[string]any) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	m, ok := s["metrics"].(map[string]any)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("n07_metrics_error 1\n"))
		return
	}
	for _, k := range []string{"requests", "success", "errors", "cancelled", "in_flight", "latency_p95_ms"} {
		if v, exists := m[k]; exists {
			_, _ = fmt.Fprintf(w, "n07_%s %v\n", k, v)
		}
	}
}
