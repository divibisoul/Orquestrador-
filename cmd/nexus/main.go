package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type request struct {
	Operation string            `json:"operation"`
	Payload   []float64         `json:"payload"`
	Metadata  map[string]string `json:"metadata"`
}

func main() {
	syncMeshSecretAlias()
	network, err := neural.New(8, .05)
	if err != nil {
		log.Fatal(err)
	}
	cortex, err := prefrontal.New(.10, 32)
	if err != nil {
		log.Fatal(err)
	}
	compute := supergpu.New(nil)
	compute.Discover()
	engine, err := orchestrator.New(network, cortex, compute)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, engine.Health())
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, engine.Stats())
	})
	mux.HandleFunc("/topology", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, orchestrator.SOULTopology())
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		writeMetrics(w, engine.Stats())
	})
	mux.Handle("/api/soul-mesh", mesh.NewEnhancedFederatedHTTPGateway(engine))
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
		result, err := engine.Execute(r.Context(), req.Operation, req.Payload, req.Metadata)
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
			operations := append([]string{}, engine.Operations()...)
			operations = append(operations, "mesh.ping", "mesh.discovery", "mesh.capabilities", "mesh.describe", "mesh.fusion.describe", "mesh.fusion.execute", "mesh.supergpu.describe", "mesh.supergpu.execute", "mesh.supergpu.parallel")
			if _, err := a.Register(regCtx, registry, unique(operations)); err != nil {
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
	_ = engine.Shutdown(shutdown)
	_ = srv.Shutdown(shutdown)
}

func syncMeshSecretAlias() {
	primary := strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET"))
	legacy := strings.TrimSpace(os.Getenv("SOUL_MESH_SECRET"))
	if primary == "" && legacy != "" {
		if err := os.Setenv("SOUL_MESH_HMAC_SECRET", legacy); err != nil {
			log.Printf("mesh secret alias setup failed: %v", err)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeMetrics(w http.ResponseWriter, stats map[string]any) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	metrics, ok := stats["metrics"].(map[string]any)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("n07_metrics_error 1\n"))
		return
	}
	for _, key := range []string{"requests", "success", "errors", "cancelled", "in_flight", "latency_p95_ms"} {
		if value, exists := metrics[key]; exists {
			_, _ = fmt.Fprintf(w, "n07_%s %v\n", key, value)
		}
	}
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
