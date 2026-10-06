package orchestrator

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

type ExternalLearningRunner struct {
	Provider string
	Command  string
}

func NewExternalLearningRunners() map[string]ExternalLearningRunner {
	return map[string]ExternalLearningRunner{
		"FedML": {
			Provider: "FedML",
			Command:  strings.TrimSpace(os.Getenv("SOUL_FEDML_COMMAND")),
		},
		"Hivemind": {
			Provider: "Hivemind",
			Command:  strings.TrimSpace(os.Getenv("SOUL_HIVEMIND_COMMAND")),
		},
	}
}

func (r ExternalLearningRunner) Run(ctx context.Context, payload []byte) ([]byte, error) {
	if strings.TrimSpace(r.Command) == "" {
		return nil, errors.New(r.Provider + "_COMMAND_NOT_CONFIGURED")
	}
	cmd := exec.CommandContext(ctx, "sh", "-lc", r.Command)
	cmd.Stdin = bytes.NewReader(payload)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}
