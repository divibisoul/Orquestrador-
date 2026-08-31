package core

import "errors"

const MaxSimulationParallelUnits = 1000

type Config struct {
	Enabled           bool
	Mode              string
	PrecisionFallback Precision
	MaxParallelUnits  int
	Simulation        SimulationConfig
}

type SimulationConfig struct {
	EfficiencyFactor float64
	UseSparsity      bool
}

func DefaultConfig() Config {
	return Config{Enabled: false, Mode: "auto", PrecisionFallback: FP32, MaxParallelUnits: MaxSimulationParallelUnits, Simulation: SimulationConfig{EfficiencyFactor: 0.7, UseSparsity: false}}
}

func (c Config) Validate() error {
	if c.EfficiencyFactorInvalid() {
		return errors.New("efficiency_factor must be > 0 and <= 1")
	}
	if c.MaxParallelUnits < 1 {
		return errors.New("max_parallel_units must be >= 1")
	}
	if c.MaxParallelUnits > MaxSimulationParallelUnits {
		return errors.New("max_parallel_units exceeds simulation safety limit of 1000")
	}
	switch c.Mode {
	case "auto", "blackwell", "vera_rubin", "cdna5", "trillium", "atlas":
	default:
		return errors.New("mode must be auto, blackwell, vera_rubin, cdna5, trillium, or atlas")
	}
	return nil
}

func (c Config) EfficiencyFactorInvalid() bool {
	return c.Simulation.EfficiencyFactor <= 0 || c.Simulation.EfficiencyFactor > 1
}
