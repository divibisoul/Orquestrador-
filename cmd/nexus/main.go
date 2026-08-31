package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/mesh"
)

type request struct { Goal string `json:"goal"` }

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workers := 4
	if v, err := strconv.Atoi(os.Getenv("NEXUS_WORKERS")); err == nil && v > 0 { workers = v }
	orch := orchestrator.NewEngine(workers)
	pf := prefrontal.New()
	reg := mesh.NewRegistry()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "nodes": len(reg.Snapshot())})
	})
	mux.HandleFunc("/v1/plan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { w.WriteHeader(http.StatusMethodNotAllowed); return }
		var q request
		if json.NewDecoder(r.Body).Decode(&q) != nil || q.Goal == "" { http.Error(w, "goal required", http.StatusBadRequest); return }
		plan := pf.GeneratePlan(q.Goal, []string{"evaluate", "execute"})
		decision := pf.Decide([]prefrontal.Plan{plan}, float64(len(plan.Steps)))
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(decision)
	})
	mux.HandleFunc("/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(reg.Snapshot())
	})

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("ORQUESTRADOR-NEXUS listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatal(err) }
	}()

	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
	_ = orch
}
