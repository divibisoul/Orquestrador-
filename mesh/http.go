		return
	}
	if envelope.Target != "N07" && envelope.Target != "BROADCAST" {
		g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "target is not N07"})
		return
	}
	capability := canonicalCapability(wire)
	if capability == "" {
		g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "capability is required"})
		return
	}
	if canonicalKind(wire) == "request" && (wire.Type == "PING" || capability == "mesh.ping") {
		g.respond(w, http.StatusOK, envelope, "TASK_RESULT", map[string]any{"ok": true, "nucleus": "N07", "contractVersion": protocol.SoulMeshContractVersion})
		return
	}
	if canonicalKind(wire) == "request" && capability == "mesh.describe" {
		g.respond(w, http.StatusOK, envelope, "TASK_RESULT", map[string]any{"nucleus": "N07", "operations": g.Engine.Operations(), "transports": []string{"LOOPBACK_HTTP", "HTTP"}})
		return
	}
	var values []float64
	metadata := envelope.NestedMetadata()
	if metadata == nil {
		metadata = map[string]string{}
	}
	if strings.HasPrefix(capability, "sara.") {
		// SARA aceita payload estruturado; o protocolo N07 interno ainda usa []float64.
		// Mantemos ambos os contratos sem criar um segundo transporte.
		raw, err := json.Marshal(envelope.NestedPayload())
		if err != nil {
			g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "invalid SARA payload"})
			return
		}
		metadata["sara_payload_json"] = string(raw)
		if payloadMap := envelope.NestedPayload(); payloadMap != nil {
			if v, ok := payloadMap["input"].(string); ok {
				metadata["sara_input"] = v
			}
			if v, ok := payloadMap["cycle_id"].(string); ok {
				metadata["sara_cycle_id"] = v
			}
			if contextPayload, ok := payloadMap["context"].(map[string]any); ok && contextPayload != nil {
				rawContext, marshalErr := json.Marshal(contextPayload)
				if marshalErr != nil {
					g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "invalid SARA context"})
					return
				}
				if len(rawContext) > 256*1024 {
					g.respond(w, http.StatusRequestEntityTooLarge, envelope, "ERROR", map[string]any{"error": "SARA context exceeds 256 KiB"})
					return
				}
				metadata["sara_context_json"] = string(rawContext)
			}
		}
	} else {
		var err error
		values, err = payloadValues(envelope.NestedPayload())
		if err != nil {
			g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": err.Error()})