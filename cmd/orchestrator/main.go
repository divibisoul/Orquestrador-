package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	network, err := neural.New(16, 0.01)
	if err != nil {
		log.Fatal(err)
	}
	cortex, err := prefrontal.New(0, 128)
	if err != nil {
		log.Fatal(err)
	}
	compute := supergpu.New(nil)
	engine, err := orchestrator.New(network, cortex, compute)
	if err != nil {
		log.Fatal(err)
	}
	compute.Discover()

	mux := http.NewServeMux()
	gateway := mesh.NewHTTPGateway(engine)
	mux.Handle("/api/soul-mesh", gateway)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, engine.Health())
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, engine.Stats())
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		stats := engine.Stats()
		metrics := stats["metrics"].(map[string]any)
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		for _, metric := range []struct {
			name string
			key  string
		}{
			{"n07_requests_total", "requests"},
			{"n07_success_total", "success"},
			{"n07_errors_total", "errors"},
			{"n07_cancelled_total", "cancelled"},
			{"n07_in_flight", "in_flight"},
			{"n07_latency_p95_ms", "latency_p95_ms"},
		} {
			value, ok := metrics[metric.key]
			if !ok {
				continue
			}
			_, _ = w.Write([]byte(metric.name + " " + jsonNumber(value) + "\n"))
		}
	})
	server := &http.Server{Addr: envOr("N07_ADDR", ":8080"), Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("http shutdown: %v", err)
		}
		if err := engine.Shutdown(shutdownCtx); err != nil {
			log.Printf("engine shutdown: %v", err)
		}
	}()

	log.Printf("N07 orchestrator listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func jsonNumber(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "0"
	}
	return string(b)
}
