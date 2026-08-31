package models

// DefaultCatalog returns the canonical reference-model set used by the selector.
// Values are software estimates only; no hardware driver or device discovery is performed.
func DefaultCatalog() []PerformanceModel {
	return []PerformanceModel{
		BlackwellB200{},
		VeraRubin{},
		CDNA5MI400{},
		Trillium{},
		Atlas950{ScaleFactor: 1},
	}
}
