package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	grpcapi "github.com/divibisoul/Orquestrador-/api/grpc"
	"github.com/divibisoul/Orquestrador-/api/rest"
	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/mesh"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workers := 4
	if v, err := strconv.Atoi(os.Getenv("NEXUS_WORKERS")); err == nil && v > 0 {
		workers = v
	}
	orch := orchestrator.NewEngine(workers)
	pf := prefrontal.New()
	reg := mesh.NewRegistry()

	restSrv := rest.NewServer(orch, pf, reg)
	httpSrv := &http.Server{Addr: ":8080", Handler: restSrv.Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("ORQUESTRADOR-NEXUS HTTP listening on %s", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	grpcAddr := os.Getenv("NEXUS_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9090"
	}
	grpcEnabled := true
	if v, err := strconv.ParseBool(os.Getenv("NEXUS_GRPC_ENABLED")); err == nil {
		grpcEnabled = v
	}
	var grpcSrv *grpc.Server
	var grpcLis net.Listener
	if grpcEnabled {
		var err error
		grpcLis, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Printf("gRPC listener startup failed")
			log.Fatal(err)
		}
		grpcSrv, err = grpcapi.Serve(grpcLis, orch)
		if err != nil {
			log.Printf("gRPC startup failed")
			log.Fatal(err)
		}
		log.Printf("ORQUESTRADOR-NEXUS gRPC listener started")
	}

	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdown)
	if grpcSrv != nil {
		grpcSrv.GracefulStop()
	}
	if grpcLis != nil {
		_ = grpcLis.Close()
	}
}
