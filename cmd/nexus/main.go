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

	"github.com/divibisoul/Orquestrador-/api/health"
	"github.com/divibisoul/Orquestrador-/backend"
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

	unified := backend.New(e, backend.DefaultConfig())
	mux := http.NewServeMux()
	mux.Handle("/v1/", unified.Handler())
	mux.Handle("/api/health/dashboard", health.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, e.Health())
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if err := requireAppBearer(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, e.Stats())
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeMetrics(w, e.Stats())
	})
	mux.HandleFunc("/identity", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, orchestrator.N07Identity())
	})
	mux.HandleFunc("/topology", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, orchestrator.SOULTopology())
	})
	mux.Handle("/api/soul-mesh", mesh.NewEnhancedFederatedHTTPGateway(e))
	mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
			return
		}
		if err := requireAppBearer(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
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
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
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
			if _, err := a.Register(regCtx, registry, []string{"neural.forward@1.0.0", "neural.learn@1.0.0", "compute.execute@1.0.0", "cognitive.execute@1.0.0", "supergpu.describe@1.0.0", "supergpu.execute@1.0.0", "supergpu.parallel@1.0.0", "supergpu.memory@1.0.0", "mesh.ping", "mesh.describe"}); err != nil {
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

func syncMeshSecretAlias() {
	primary := strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET"))
	legacy := strings.TrimSpace(os.Getenv("SOUL_MESH_SECRET"))
	if primary == "" && legacy != "" {
		if err := os.Setenv("SOUL_MESH_HMAC_SECRET", legacy); err != nil {
			log.Printf("mesh secret alias setup failed: %v", err)
		}
	}
}

func requireAppBearer(r *http.Request) error {
	expected := strings.TrimSpace(os.Getenv("N07_APP_TOKEN"))
	if expected == "" {
		return errors.New("N07_APP_TOKEN is not configured")
	}
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authorization) < 7 || !strings.EqualFold(authorization[:7], "bearer ") {
		return errors.New("Bearer authentication required")
	}
	provided := strings.TrimSpace(authorization[7:])
	if provided == "" || provided != expected {
		return errors.New("invalid application token")
	}
	return nil
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
