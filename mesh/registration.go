package mesh

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type PeerRegistration struct {
	Nucleus       string    `json:"nucleus"`
	Endpoint      string    `json:"endpoint"`
	CorrelationID string    `json:"correlationId"`
	Capabilities  []string  `json:"capabilities"`
	RegisteredAt  time.Time `json:"registeredAt"`
}

type RegistrationRegistry struct {
	mu      sync.RWMutex
	entries map[string]PeerRegistration
}

func NewRegistrationRegistry() *RegistrationRegistry {
	return &RegistrationRegistry{entries: make(map[string]PeerRegistration)}
}

func (r *RegistrationRegistry) Register(message protocol.Message) (PeerRegistration, error) {
	if r == nil {
		return PeerRegistration{}, errors.New("registration registry is unavailable")
	}
	nucleus := strings.TrimSpace(message.Source)
	if nucleus == "" || nucleus == "N07" {
		return PeerRegistration{}, errors.New("invalid registering nucleus")
	}
	endpoint := strings.TrimSpace(message.Metadata["endpoint"])
	if endpoint == "" {
		return PeerRegistration{}, errors.New("registration endpoint is required")
	}
	registration := PeerRegistration{
		Nucleus: nucleus,
		Endpoint: endpoint,
		CorrelationID: message.CorrelationID,
		Capabilities: strings.Fields(strings.TrimSpace(message.Metadata["capabilities"])),
		RegisteredAt: time.Now().UTC(),
	}
	r.mu.Lock()
	r.entries[nucleus] = registration
	r.mu.Unlock()
	return registration, nil
}

func (r *RegistrationRegistry) Snapshot() []PeerRegistration {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]PeerRegistration, 0, len(r.entries))
	for _, entry := range r.entries {
		entry.Capabilities = append([]string(nil), entry.Capabilities...)
		out = append(out, entry)
	}
	return out
}

func DefaultRegistrationRegistry() *RegistrationRegistry {
	return defaultRegistrationRegistry
}

var defaultRegistrationRegistry = NewRegistrationRegistry()

func RegisterRegistrationOperation(e *orchestrator.Engine, registry *RegistrationRegistry) error {
	if e == nil || registry == nil {
		return errors.New("registration operation requires engine and registry")
	}
	handler := func(_ context.Context, message protocol.Message) (protocol.Result, error) {
		registration, err := registry.Register(message)
		if err != nil {
			return protocol.Result{
				TraceID: message.TraceID,
				CorrelationID: message.CorrelationID,
				Source: "N07.mesh",
				Target: message.Source,
				Status: "rejected",
			}, err
		}
		return protocol.Result{
			TraceID: message.TraceID,
			CorrelationID: message.CorrelationID,
			Source: "N07.mesh",
			Target: message.Source,
			Status: "ok",
			Metadata: map[string]string{
				"registered_nucleus":  registration.Nucleus,
				"registered_endpoint": registration.Endpoint,
				"registered_at":       registration.RegisteredAt.Format(time.RFC3339Nano),
			},
		}, nil
	}
	if err := e.Register("mesh.register", handler); err != nil {
		return err
	}
	return e.Register("mesh.register@1.0.0", handler)
}
