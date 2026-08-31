package core

import "errors"

// MaxSimulationParallelUnits is the hard runtime ceiling for simulated worker goroutines.
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
	return Config{
		Enabled:           false,
		Mode:              "auto",
		PrecisionFallback: FP32,
		MaxParallelUnits:  8192,
		Simulation: SimulationConfig{
			EfficiencyFactor: 0.7,
			UseSparsity:      false,
		},
	}
}

func (c Config) Validate() error {
	if c.EfficiencyFactorInvalid() {
		return errors.New("efficiency_factor must be > 0 and <= 1")
	}
	if c.MaxParallelUnits < 1 {
		return errors.New("max_parallel_units must be >= 1")
	}
	switch c.Mode {
	case "auto", "blackwell", "vera_rubin", "cdna5", "trillium", "atlas":
	default:
		return errors.New("mode must be auto, blackwell, vera_rubin, cdna5, trillium, or atlas")
	}
	if !c.PrecisionFallback.Valid() {
		return errors.New("precision_fallback must be a supported precision")
	}
	return nil
}

func (c Config) EffectiveParallelUnits() int {
	units := c.MaxParallelUnits
	if units < 1 {
		units = 1
	}
	if units > MaxSimulationParallelUnits {
		units = MaxSimulationParallelUnits
	}
	return units
}

func (c Config) EfficiencyFactorInvalid() bool {
	return c.Simulation.EfficiencyFactor <= 0 || c.Simulation.EfficiencyFactor > 1
}
