import hashlib
import json
import math
import os
import statistics
import time
from pathlib import Path

import torch
import transformers
from huggingface_hub import snapshot_download
from transformers import pipeline

MODEL_ID = "distilbert/distilbert-base-uncased"
TEXT = (
    "SOUL neocortex preserves correlation while routing a neural signal "
    "through authorized processing and an agent decision."
)
torch.set_num_threads(1)

# Resolve the model once and record the immutable Hub snapshot revision that was
# actually downloaded. Download only the PyTorch safetensors and tokenizer files.
snapshot = snapshot_download(
    repo_id=MODEL_ID,
    revision="main",
    allow_patterns=[
        "config.json",
        "model.safetensors",
        "tokenizer.json",
        "tokenizer_config.json",
        "special_tokens_map.json",
        "added_tokens.json",
        "vocab.txt",
    ],
)
model_sha = Path(snapshot).name
if len(model_sha) != 40 or any(c not in "0123456789abcdef" for c in model_sha.lower()):
    raise RuntimeError(f"MODEL_REVISION_SHA_NOT_RESOLVED:{model_sha}")
if not (Path(snapshot) / "model.safetensors").is_file():
    raise RuntimeError("SAFE_TENSORS_WEIGHTS_MISSING")

extractor = pipeline(
    task="feature-extraction",
    model=snapshot,
    tokenizer=snapshot,
    device="cpu",
    framework="pt",
)
extractor.model.eval()

def run_once():
    started = time.perf_counter_ns()
    with torch.inference_mode():
        result = extractor(TEXT, return_tensors=True, truncation=True, max_length=32)
    elapsed_ms = (time.perf_counter_ns() - started) / 1_000_000
    tensor = result if isinstance(result, torch.Tensor) else torch.as_tensor(result)
    tensor = tensor.detach().cpu().contiguous()
    if tensor.ndim != 3 or tensor.shape[0] != 1:
        raise RuntimeError(f"UNEXPECTED_FEATURE_SHAPE:{tuple(tensor.shape)}")
    if not torch.isfinite(tensor).all().item():
        raise RuntimeError("NON_FINITE_MODEL_OUTPUT")
    digest = hashlib.sha256(tensor.numpy().tobytes()).hexdigest()
    return elapsed_ms, tensor, digest

# Warm-up is excluded from the ten reported measurements.
for _ in range(2):
    run_once()

latencies = []
digests = []
first = None
for _ in range(10):
    elapsed, tensor, digest = run_once()
    latencies.append(round(elapsed, 3))
    digests.append(digest)
    if first is None:
        first = tensor

result = {
    "stage": "GROUP_F_LOCAL_TRANSFORMERS_PYTORCH",
    "status": "PASS" if len(set(digests)) == 1 else "FAIL",
    "model_id": MODEL_ID,
    "model_revision_sha": model_sha,
    "weights": "model.safetensors",
    "transformers_version": transformers.__version__,
    "torch_version": torch.__version__,
    "backend": "PyTorch CPU",
    "device": "cpu",
    "task": "feature-extraction",
    "tokenizer": type(extractor.tokenizer).__name__,
    "model_type": extractor.model.config.model_type,
    "output_shape": list(first.shape),
    "output_literal_first_token_first_8": [round(float(x), 8) for x in first[0, 0, :8]],
    "output_sha256": digests[0],
    "all_ten_outputs_identical": len(set(digests)) == 1,
    "latency_ms_each": latencies,
    "latency_p50_ms": round(statistics.median(latencies), 3),
    "warmup_calls_excluded": 2,
    "measured_calls": len(latencies),
    "tensorflow_installed": __import__("importlib").util.find_spec("tensorflow") is not None,
    "jax_installed": __import__("importlib").util.find_spec("jax") is not None,
}
print(json.dumps(result, sort_keys=True))
if result["status"] != "PASS":
    raise SystemExit(1)
if result["tensorflow_installed"] or result["jax_installed"]:
    raise RuntimeError("GROUP_F_ENVIRONMENT_NOT_ISOLATED_FROM_TF_JAX")
