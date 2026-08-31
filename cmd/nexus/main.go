package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	grpcapi "github.com/divibisoul/Orquestrador-/api/grpc"
	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/mesh"
	"google.golang.org/grpc"
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

	httpSrv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("ORQUESTRADOR-NEXUS HTTP listening on %s", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatal(err) }
	}()

	grpcAddr := os.Getenv("NEXUS_GRPC_ADDR")
	if grpcAddr == "" { grpcAddr = ":9090" }
	grpcEnabled := true
	if v, err := strconv.ParseBool(os.Getenv("NEXUS_GRPC_ENABLED")); err == nil { grpcEnabled = v }
	var grpcSrv *grpc.Server
	var grpcLis net.Listener
	if grpcEnabled {
		var err error
		grpcLis, err = net.Listen("tcp", grpcAddr)
		if err != nil { log.Fatalf("gRPC listen %s: %v", grpcAddr, err) }
		grpcSrv, err = grpcapi.Serve(grpcLis, orch)
		if err != nil { log.Fatalf("gRPC startup: %v", err) }
		log.Printf("ORQUESTRADOR-NEXUS gRPC listening on %s", grpcAddr)
	}

	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdown)
	if grpcSrv != nil { grpcSrv.GracefulStop() }
	if grpcLis != nil { _ = grpcLis.Close() }
}
