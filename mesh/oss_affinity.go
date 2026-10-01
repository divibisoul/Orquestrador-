package mesh

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ossAffinityManifest struct {
	SchemaVersion string `json:"schemaVersion"`
	Capabilities  map[string][]string `json:"capabilities"`
}

var (
	ossAffinityOnce sync.Once
	ossAffinityData map[string][]string
)

func loadOSSAffinityFromBytes(data []byte) (map[string][]string, error) {
	var wrapper struct {
		OSSAffinityRouting *ossAffinityManifest `json:"ossAffinityRouting"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.OSSAffinityRouting == nil {
		return nil, errors.New("ossAffinityRouting section is missing")
	}
	if wrapper.OSSAffinityRouting.SchemaVersion != "1.0.0" {
		return nil, errors.New("unsupported ossAffinityRouting schema version")
	}
	if wrapper.OSSAffinityRouting.Capabilities == nil {
		return nil, errors.New("ossAffinityRouting capabilities are missing")
	}
	out := make(map[string][]string, len(wrapper.OSSAffinityRouting.Capabilities))
	for capability, targets := range wrapper.OSSAffinityRouting.Capabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" || len(targets) == 0 {
			return nil, errors.New("invalid OSS affinity entry")
		}
		clean := make([]string, 0, len(targets))
		seen := map[string]struct{}{}
		for _, target := range targets {
			target = strings.TrimSpace(target)
			if target == "" {
				return nil, errors.New("empty OSS affinity target")
			}
			if _, ok := seen[target]; ok {
				continue
			}
			seen[target] = struct{}{}
			clean = append(clean, target)
		}
		out[capability] = clean
	}
	return out, nil
}

func loadOSSAffinityFile() map[string][]string {
	candidates := make([]string, 0, 3)
	if configured := strings.TrimSpace(os.Getenv("SOUL_CAPABILITY_AUTHORITY_PATH")); configured != "" {
		candidates = append(candidates, configured)
	}
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		candidates = append(candidates, filepath.Join(cwd, "soul-capability-authority.json"))
	}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "soul-capability-authority.json"))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		affinity, err := loadOSSAffinityFromBytes(data)
		if err == nil {
			return affinity
		}
	}
	return nil
}

func runtimeOSSAffinityTargets(capability string) []string {
	ossAffinityOnce.Do(func() {
		ossAffinityData = loadOSSAffinityFile()
	})
	targets := ossAffinityData[strings.TrimSpace(capability)]
	if len(targets) == 0 {
		return nil
	}
	return append([]string(nil), targets...)
}

func ossAffinityRank(capability, nucleus string) int {
	targets := runtimeOSSAffinityTargets(capability)
	for index, target := range targets {
		if target == nucleus {
			return index
		}
	}
	return len(targets) + 1000
}
