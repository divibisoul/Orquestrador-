package cognitive

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func cognitiveConfigFromEnv() Config {
	cfg := DefaultConfig()
	cfg.Enabled = parseBool(os.Getenv("N07_COGNITIVE_LOOP_ENABLED"))
	if value, err := strconv.Atoi(os.Getenv("N07_COGNITIVE_WORKING_MEMORY_ITEMS")); err == nil && value > 0 {
		cfg.WorkingMemoryItems = value
	}
	if value, err := strconv.Atoi(os.Getenv("N07_COGNITIVE_MAX_PARALLEL_STEPS")); err == nil && value > 0 {
		cfg.MaxParallelSteps = value
	}
	if cfg.WorkingMemoryTTL <= 0 {
		cfg.WorkingMemoryTTL = 15 * time.Minute
	}
	if cfg.GoalTTL <= 0 {
		cfg.GoalTTL = 10 * time.Minute
	}
	return cfg
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
