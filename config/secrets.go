package config

import (
	"os"
	"strings"
)

type SecretsStatus struct {
	Missing []string
	Status  string
	Mode    string
}

var requiredSecrets = []string{
	"N07_APP_TOKEN",
	"WEB3_STORAGE_TOKEN",
	"SUPABASE_URL",
	"SUPABASE_SERVICE_ROLE_KEY",
}

func InspectSecrets() SecretsStatus {
	missing := make([]string, 0, len(requiredSecrets))
	for _, name := range requiredSecrets {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			missing = append(missing, name)
		}
	}
	status := "ready"
	mode := "full"
	if len(missing) > 0 {
		status = "degraded"
		mode = "restricted"
	}
	return SecretsStatus{Missing: missing, Status: status, Mode: mode}
}

func RequiredSecrets() []string {
	out := make([]string, len(requiredSecrets))
	copy(out, requiredSecrets)
	return out
}
