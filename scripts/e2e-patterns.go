package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func detectFailurePatternsImproved(current []probe) []string {
	data, err := os.ReadFile(env("SOUL_E2E_HISTORY_FILE", ".soul/e2e-history.json"))
	if err != nil {
		return nil
	}

	var history []historyEntry
	if err := json.Unmarshal(data, &history); err != nil {
		return nil
	}

	cutoff := time.Now().Add(-2 * time.Hour)
	patterns := make([]string, 0)
	seen := make(map[string]bool)

	for _, item := range current {
		if item.Status == "healthy" {
			continue
		}

		key := item.Source + "→" + item.Target
		if seen[key] {
			continue
		}
		seen[key] = true

		count := 0
		for i := len(history) - 1; i >= 0; i-- {
			entry := history[i]
			if entry.At.Before(cutoff) {
				break
			}
			if entry.Source == item.Source && entry.Target == item.Target && entry.Status != "healthy" {
				count++
			}
		}

		if count >= 3 {
			patterns = append(patterns, fmt.Sprintf(
				"%s→%s falhou %d vezes nas últimas 2 horas; verificar rede, autenticação e logs do peer.",
				item.Source,
				item.Target,
				count,
			))
		}
	}

	if len(patterns) == 0 {
		return nil
	}

	for i := range patterns {
		patterns[i] = strings.TrimSpace(patterns[i])
	}
	return unique(patterns)
}
