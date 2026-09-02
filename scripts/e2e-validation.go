package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	contractVersion = "1.1.0"
	protocolVersion = "soul-mesh/1"
	maxResponse    = 1 << 20
)

type probe struct {
	Pair             string `json:"pair"`
	Source           string `json:"source"`
	Target           string `json:"target"`
	Route            string `json:"route"`
	CorrelationID    string `json:"correlationId"`
	Status           string `json:"status"`
	HTTP             int    `json:"http,omitempty"`
	LatencyMs        int64  `json:"latencyMs"`
	ReturnedID       string `json:"returnedCorrelationId,omitempty"`
	Error            string `json:"error,omitempty"`
	CorrectionRecipe string `json:"correctionRecipe"`
}

type report struct {
	GeneratedAt     string  `json:"generatedAt"`
	Protocol        string  `json:"protocol"`
	ContractVersion string  `json:"contractVersion"`
	PairsExpected   int     `json:"pairsExpected"`
	DirectedProbes  int     `json:"directedProbes"`
	Passed          int     `json:"passed"`
	Failed          int     `json:"failed"`
	Limitations     string  `json:"limitations"`
	Results         []probe `json:"results"`
}

func main() {
	client := &http.Client{Timeout: 8 * time.Second}
	nuclei := []string{"N01", "N02", "N03", "N04", "N05", "N06"}
	token := strings.TrimSpace(os.Getenv("SOUL_MESH_TOKEN"))
	endpointSuffix := env("SOUL_MESH_ENDPOINT", "/api/soul-mesh")

	var results []probe
	for i := 0; i < len(nuclei); i++ {
		for j := i + 1; j < len(nuclei); j++ {
			pair := nuclei[i] + "↔" + nuclei[j]
			for _, direction := range [][2]string{{nuclei[i], nuclei[j]}, {nuclei[j], nuclei[i]}} {
				results = append(results, runProbe(context.Background(), client, pair, direction[0], direction[1], endpointSuffix, token))
			}
		}
	}

	failed := 0
	passed := 0
	for _, result := range results {
		if result.Status == "healthy" {
			passed++
		} else {
			failed++
		}
		encoded, _ := json.Marshal(result)
		fmt.Println(string(encoded))
	}

	reportData := report{
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Protocol:        protocolVersion,
		ContractVersion: contractVersion,
		PairsExpected:   15,
		DirectedProbes:  len(results),
		Passed:          passed,
		Failed:          failed,
		Limitations:     "O executor roda no N07. Os 30 testes exercitam o contrato direcionado via endpoints configurados; eles não substituem uma prova física de processo-origem quando a infraestrutura não permite que N07 observe o tráfego interno entre dois núcleos. O teste federado N01→N07 existente continua sendo complementar.",
		Results:         results,
	}

	if err := writeJSON("e2e-validation-report.json", reportData); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func runProbe(ctx context.Context, client *http.Client, pair, source, target, suffix, token string) probe {
	correlationID := traceID()
	result := probe{Pair: pair, Source: source, Target: target, CorrelationID: correlationID, Status: "failed"}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_MESH_"+target+"_URL")), "/")
	if baseURL == "" {
		result.Status = "not-configured"
		result.CorrectionRecipe = fmt.Sprintf("Configure SOUL_MESH_%s_URL. Sem endpoint não há prova E2E; não classificar como saudável.", target)
		return result
	}
	endpoint := baseURL + suffix
	result.Route = endpoint

	message := map[string]any{
		"protocol":        protocolVersion,
		"contractVersion": contractVersion,
		"id":              correlationID,
		"correlationId":   correlationID,
		"source":          source,
		"target":          target,
		"kind":            "request",
		"capability":      "mesh.ping",
		"payload":         map[string]any{"e2e": true, "requestedBy": "N07", "pair": pair},
		"timestamp":       time.Now().UnixMilli(),
		"meta":            map[string]any{"runtime": "Orquestrador-", "transport": "HTTP", "encoding": "json", "version": contractVersion, "traceId": correlationID},
	}
	body, err := json.Marshal(message)
	if err != nil {
		result.Error = err.Error()
		result.CorrectionRecipe = "Corrija a serialização do envelope e mantenha protocol/contractVersion/correlationId obrigatórios."
		return result
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		result.Error = err.Error()
		result.CorrectionRecipe = "Corrija a URL/rota publicada do núcleo alvo."
		return result
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-soul-correlation-id", correlationID)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	started := time.Now()
	response, err := client.Do(req)
	result.LatencyMs = time.Since(started).Milliseconds()
	if err != nil {
		result.Status = "unreachable"
		result.Error = err.Error()
		result.CorrectionRecipe = fmt.Sprintf("%s→%s inacessível. Verifique serviço, ingress/DNS e faça redeploy controlado; depois valide novamente.", source, target)
		return result
	}
	defer response.Body.Close()

	result.HTTP = response.StatusCode
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, maxResponse))
	var reply map[string]any
	if err := json.Unmarshal(responseBody, &reply); err != nil {
		result.Status = "invalid-response"
		result.Error = "response is not valid JSON"
		result.CorrectionRecipe = "Corrija o adaptador Soul Mesh para devolver envelope JSON canônico com correlationId."
		return result
	}

	returnedCorrelation, _ := reply["correlationId"].(string)
	result.ReturnedID = returnedCorrelation
	validEnvelope := reply["protocol"] == protocolVersion && reply["contractVersion"] == contractVersion && returnedCorrelation == correlationID
	if response.StatusCode == http.StatusOK && validEnvelope {
		result.Status = "healthy"
		result.CorrectionRecipe = "Nenhuma: contrato, autenticação, rota e correlationId confirmados."
		return result
	}

	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		result.Status = "unauthorized"
		result.CorrectionRecipe = "401/403: verifique SOUL_MESH_TOKEN e sincronize a mesma credencial nos dois lados; não desative autenticação."
	case http.StatusNotFound:
		result.Status = "route-not-found"
		result.CorrectionRecipe = "404: confirme SOUL_MESH_ENDPOINT e a rota exposta pelo runtime; adapte a borda sem substituir o runtime existente."
	case http.StatusConflict:
		result.Status = "replay-or-conflict"
		result.CorrectionRecipe = "409: regenere messageId/correlationId e valide nonce/replay; não reutilize mensagens."
	case http.StatusBadRequest:
		result.Status = "contract-error"
		result.CorrectionRecipe = "400: normalize o envelope para soul-mesh/1 e contract 1.1.0 e assegure correlationId igual na resposta."
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		result.Status = "timeout"
		result.CorrectionRecipe = "Timeout: preserve timeout curto, acione circuit breaker e rota alternativa antes de repetir; não aumente timeout cegamente."
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		result.Status = "remote-error"
		result.CorrectionRecipe = "5xx: inspecione logs/telemetria do alvo, restaure dependências e só então reabra o tráfego."
	default:
		result.Status = "invalid-response"
		if !validEnvelope {
			result.CorrectionRecipe = "Resposta sem envelope canônico: alinhe protocol, contractVersion e correlationId sem migrar prematuramente para 1.2.0."
		} else {
			result.CorrectionRecipe = "Status HTTP inesperado: revisar contrato da rota e política de erro do peer."
		}
	}
	return result
}

func traceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("e2e-%d", time.Now().UnixNano())
	}
	return "e2e-" + hex.EncodeToString(b[:])
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}
