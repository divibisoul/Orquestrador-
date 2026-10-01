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

	"github.com/divibisoul/Orquestrador-/aeternum"
	"github.com/divibisoul/Orquestrador-/api"
	"github.com/divibisoul/Orquestrador-/api/health"
	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/cognitive"
	"github.com/divibisoul/Orquestrador-/cooperation"
	"github.com/divibisoul/Orquestrador-/jev"
	"github.com/divibisoul/Orquestrador-/learning"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/octacore"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/rgo"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type request struct {
	Operation string            `json:"operation"`
	Payload   []float64         `json:"payload"`
	Metadata  map[string]string `json:"metadata"`
}

type n07CognitiveMeshAdapter struct {
	peers *mesh.PeerClient
}

func (a n07CognitiveMeshAdapter) ConfiguredPeers() []cognitive.PeerDescriptor {
	if a.peers == nil {
		return nil
	}
	raw := a.peers.ConfiguredPeers()
	out := make([]cognitive.PeerDescriptor, 0, len(raw))
	for _, peer := range raw {
		out = append(out, cognitive.PeerDescriptor{Nucleus: peer.Nucleus})
	}
	return out
}

func (a n07CognitiveMeshAdapter) Discover(ctx context.Context, nucleus string) (map[string]any, error) {
	if a.peers == nil {
		return nil, errors.New("mesh peer client unavailable")
	}
	return a.peers.Discover(ctx, nucleus)
}

func (a n07CognitiveMeshAdapter) CallBestDynamic(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
	if a.peers == nil {
		return nil, "", errors.New("mesh peer client unavailable")
	}
	return a.peers.CallBestDynamic(ctx, capability, payload, correlation)
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

	learningStore := backend.NewLearningStoreFromEnv()
	var learningSink learning.Store
	if learningStore.Configured() {
		learningSink = learningStore
	}
	learningMachine, err := learning.New(n, c, learningSink)
	if err != nil {
		log.Fatal(err)
	}
	if learningSink != nil {
		restoreCtx, restoreCancel := context.WithTimeout(context.Background(), 15*time.Second)
		if err := learningMachine.Restore(restoreCtx, 5000); err != nil {
			log.Printf("learning restore blocked: %v", err)
		} else {
			log.Printf("learning experiences restored: %v", learningMachine.Snapshot().Restored)
		}
		restoreCancel()
	}
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		log.Fatal(err)
	}
	semanticMemoryStore := backend.NewSemanticMemoryStoreFromEnv()
	if semanticMemoryStore.Configured() {
		if err := e.SetMemoryStore(semanticMemoryStore); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("semantic memory store disabled: Supabase credentials are not configured")
	}
	if err := orchestrator.RegisterLearningOperations(e, learningMachine, n); err != nil {
		log.Fatal(err)
	}
	if err := orchestrator.RegisterSuperGPUOperations(e); err != nil {
		log.Fatal(err)
	}
	if err := orchestrator.RegisterAdvancedOperations(e); err != nil {
		log.Fatal(err)
	}
	if err := orchestrator.RegisterOrbitalReasoningOperations(e); err != nil {
		log.Fatal(err)
	}
	if client, err := jev.NewFromEnv(); err == nil {
		if err := orchestrator.RegisterJevOperations(e, client); err != nil {
			log.Fatal(err)
		}
		log.Printf("Jev decision capability enabled: %s", client.Model)
	} else {
		log.Printf("Jev decision capability disabled: %v", err)
	}
	cfg := backend.DefaultConfig()
	saraProxy := backend.NewSARAProxy(cfg)
	if saraProxy.Configured() {
		if err := backend.RegisterSARAOperations(e, saraProxy); err != nil {
			log.Fatal(err)
		}
	}
	if err := rgo.RegisterOperation(e, saraProxy); err != nil {
		log.Fatal(err)
	}

	peerClient, err := mesh.NewPeerClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	peerClient.SetRouteScorer(learningMachine)
	peerClient.SetRouteOutcomeObserver(learningMachine)

	coordinator, err := cooperation.New(peerClient, learningMachine)
	if err != nil {
		log.Fatal(err)
	}
	if err := orchestrator.RegisterCooperationOperations(e, coordinator); err != nil {
		log.Fatal(err)
	}

	clareiraReporter, err := mesh.NewClareiraReporter(peerClient)
	if err != nil {
		log.Fatal(err)
	}
	g.SetExecutionReporter(supergpu.ReporterFunc(func(ctx context.Context, event supergpu.ExecutionEvent) error {
		correlationID := event.CorrelationID
		if strings.TrimSpace(correlationID) == "" {
			correlationID = protocol.NewTraceID()
		}
		reportCtx := supergpu.WithCorrelationID(ctx, correlationID)
		clareiraErr := clareiraReporter.Report(reportCtx, event)
		if !saraProxy.Configured() {
			return clareiraErr
		}
		_, vagusErr := saraProxy.PublishVagus(ctx, map[string]any{
			"vagus_version":  "1.0",
			"message_id":     protocol.NewTraceID(),
			"correlation_id": correlationID,
			"source":         "N07.SuperGPU",
			"target":         "VagusNerveBus",
			"priority":       100,
			"ttl":            5000,
			"type":           "supergpu." + event.Phase,
			"payload": map[string]any{
				"operation":   event.Operation,
				"device_id":   event.DeviceID,
				"backend":     event.Backend,
				"input_size":  event.InputSize,
				"output_size": event.OutputSize,
				"error":       event.Error,
			},
		}, correlationID)
		if clareiraErr != nil && vagusErr != nil {
			return fmt.Errorf("clareira=%v; vagus=%v", clareiraErr, vagusErr)
		}
		if clareiraErr != nil {
			return clareiraErr
		}
		return vagusErr
	}))

	hortaCore, err := aeternum.NewHortaCore(e, saraProxy)
	if err != nil {
		log.Fatal(err)
	}
	hortaCore.SetPeerClient(peerClient)
	if err := orchestrator.RegisterCognitiveOperations(e, n07CognitiveMeshAdapter{peers: peerClient}, saraProxy, backend.NewSupabaseStore(cfg)); err != nil {
		log.Fatal(err)
	}
	if err := orchestrator.RegisterPrimordialCompositionOperation(e); err != nil {
		log.Fatal(err)
	}
	if err := rgo.RegisterTrinityOperation(e, saraProxy, peerClient); err != nil {
		log.Fatal(err)
	}
	octacoreProcessor, err := octacore.NewProcessor(octacore.DefaultConfig(), g, peerClient, saraProxy)
	if err != nil {
		log.Fatal(err)
	}
	octacoreProcessor.SetVagusPublisher(func(ctx context.Context, event octacore.VagusEnvelope) error {
		if !saraProxy.Configured() {
			return errors.New("VAGUS_CONTROL_SARA_UNCONFIGURED")
		}
		payload := map[string]any{
			"vagus_version":  event.VagusVersion,
			"message_id":     event.MessageID,
			"correlation_id": event.CorrelationID,
			"source":         event.Source,
			"target":         event.Target,
			"priority":       event.Priority,
			"ttl":            event.TTL,
			"type":           event.Type,
			"payload":        event.Payload,
		}
		_, err := saraProxy.PublishVagus(ctx, payload, event.CorrelationID)
		return err
	})
	if err := octacore.RegisterOperations(e, octacoreProcessor); err != nil {
		log.Fatal(err)
	}
	octacoreFusion, err := octacore.NewFusion(e, peerClient, g)
	if err != nil {
		log.Fatal(err)
	}
	if err := octacoreFusion.Register(); err != nil {
		log.Fatal(err)
	}

	unified := backend.NewUnified(e, cfg)
	openAICompat := api.NewOpenAICompatHandler(peerClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", openAICompat.ServeModels)
	mux.HandleFunc("/v1/chat/completions", openAICompat.ServeChat)
	mux.Handle("/v1/", unified.Handler())
	mux.Handle("/api/health/dashboard", health.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, e.Health()) })
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if err := requireAppBearer(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, e.Stats())
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { writeMetrics(w, e.Stats()) })
	mux.HandleFunc("/identity", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, orchestrator.N07Identity()) })
	mux.HandleFunc("/topology", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, orchestrator.SOULTopology()) })
	mux.HandleFunc("/v1/aeternum/health", func(w http.ResponseWriter, r *http.Request) {
		if err := requireAppBearer(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, hortaCore.Health())
	})
	mux.HandleFunc("/v1/aeternum/processors", func(w http.ResponseWriter, r *http.Request) {
		if err := requireAppBearer(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(hortaCore.Processors()), "processors": hortaCore.Processors()})
	})
	mux.HandleFunc("/v1/aeternum/capabilities", func(w http.ResponseWriter, r *http.Request) {
		if err := requireAppBearer(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(hortaCore.Capabilities()), "modules": hortaCore.Capabilities()})
	})
	mux.HandleFunc("/v1/aeternum/module", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
			return
		}
		if err := requireAppBearer(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		var req struct {
			ModuleID string            `json:"module_id"`
			Payload  []float64         `json:"payload"`
			Metadata map[string]string `json:"metadata"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		result, err := hortaCore.Execute(r.Context(), req.ModuleID, req.Payload, req.Metadata)
		if err != nil {
			if strings.HasPrefix(err.Error(), "AETERNUM_BLOCKED_INFRASTRUCTURE:") || strings.HasPrefix(err.Error(), "AETERNUM_ADAPTER_REQUIRED:") || strings.HasPrefix(err.Error(), "AETERNUM_PEER_REQUIRED:") {
				writeJSON(w, http.StatusConflict, map[string]any{"status": "BLOCKED", "error": err.Error()})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]any{"status": "ERROR", "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
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
			if _, err := a.Register(regCtx, registry, e.Operations()); err != nil {
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
	w.Header().Set("Content-Type", "text/plain; version=0.4.0")
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
