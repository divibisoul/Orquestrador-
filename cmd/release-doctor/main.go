package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type peerReport struct {
	Nucleus    string `json:"nucleus"`
	Configured bool   `json:"configured"`
	URL        string `json:"url,omitempty"`
	Reachable  bool   `json:"reachable"`
	Status     int    `json:"status,omitempty"`
	LatencyMS  int64  `json:"latencyMs,omitempty"`
	Error      string `json:"error,omitempty"`
}

type report struct {
	ReadyForE2E      bool         `json:"readyForE2E"`
	ConfiguredPeers   int          `json:"configuredPeers"`
	ReachablePeers    int          `json:"reachablePeers"`
	MissingNuclei     []string     `json:"missingNuclei,omitempty"`
	HMACConfigured    bool         `json:"hmacConfigured"`
	ContractVersion   string       `json:"contractVersion"`
	GeneratedAt       string       `json:"generatedAt"`
	Peers             []peerReport `json:"peers"`
}

func main() {
	probe := flag.Bool("probe", false, "probe configured peer /mesh/discovery endpoints")
	timeout := flag.Duration("timeout", 3*time.Second, "per-peer probe timeout")
	jsonOut := flag.Bool("json", true, "emit JSON report")
	flag.Parse()

	nuclei := []string{"N01", "N02", "N03", "N04", "N05", "N06"}
	reports := make([]peerReport, 0, len(nuclei))
	missing := make([]string, 0)
	for _, nucleus := range nuclei {
		url := strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_MESH_"+nucleus+"_URL")), "/")
		item := peerReport{Nucleus: nucleus, Configured: url != "", URL: url}
		if url == "" {
			missing = append(missing, nucleus)
			reports = append(reports, item)
			continue
		}
		if *probe {
			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			started := time.Now()
			resp, err := (&http.Client{Timeout: *timeout}).Get(url + "/mesh/discovery")
			item.LatencyMS = time.Since(started).Milliseconds()
			cancel()
			if err != nil {
				item.Error = err.Error()
			} else {
				item.Reachable = resp.StatusCode >= 200 && resp.StatusCode < 300
				item.Status = resp.StatusCode
				resp.Body.Close()
			}
			_ = ctx
		}
		reports = append(reports, item)
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].Nucleus < reports[j].Nucleus })
	configured, reachable := 0, 0
	for _, item := range reports {
		if item.Configured { configured++ }
		if item.Reachable { reachable++ }
	}
	out := report{
		ReadyForE2E:    configured == len(nuclei) && (!*probe || reachable == len(nuclei)),
		ConfiguredPeers: configured,
		ReachablePeers: reachable,
		MissingNuclei: missing,
		HMACConfigured: strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET")) != "",
		ContractVersion: "1.1.0",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Peers: reports,
	}
	if *jsonOut {
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil { panic(err) }
		fmt.Println(string(data))
		return
	}
	fmt.Printf("configured=%d/6 reachable=%d/6 hmac=%t readyForE2E=%t\n", configured, reachable, out.HMACConfigured, out.ReadyForE2E)
}
