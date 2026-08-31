# Neural Control Fabric

The Neural Control Fabric is the adaptive boundary between executive cognition and compute execution.

Pipeline: State Encoder -> Task/Hardware features -> capability/cost/latency/quality predictors -> route scorer -> Compute Fabric -> measured outcome -> feedback/replay.

Initial predictors may be deterministic baselines. A learned model is promoted only when validation metrics beat the baseline and confidence is calibrated.

The design deliberately avoids claiming AGI: it supplies a trainable control architecture and model-provider interfaces.
