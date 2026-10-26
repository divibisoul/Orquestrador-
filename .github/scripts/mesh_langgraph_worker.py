#!/usr/bin/env python3
"""Persistent, test-only Mesh sidecar for real local Transformers + LangGraph inference."""
import hashlib
import importlib.metadata
import json
import sys
import time
from pathlib import Path
from typing import TypedDict

import torch
import transformers
from huggingface_hub import snapshot_download
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, START, StateGraph
from transformers import pipeline

MODEL_ID = "distilbert/distilbert-base-uncased"
MODEL_REVISION = "12040accade4e8a0f71eabdb258fecc2e7e948be"
DEFAULT_TEXT = (
    "SOUL neocortex preserves correlation while routing a neural signal "
    "through authorized processing and an agent decision."
)

torch.set_num_threads(1)
snapshot = snapshot_download(
    repo_id=MODEL_ID,
    revision=MODEL_REVISION,
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
if Path(snapshot).name != MODEL_REVISION:
    raise RuntimeError(f"MODEL_REVISION_MISMATCH:{Path(snapshot).name}")
weights_path = Path(snapshot) / "model.safetensors"
if not weights_path.is_file():
    raise RuntimeError("SAFE_TENSORS_WEIGHTS_MISSING")

def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()

WEIGHTS_SHA256 = sha256_file(weights_path)
extractor = pipeline(
    task="feature-extraction",
    model=snapshot,
    tokenizer=snapshot,
    device="cpu",
    framework="pt",
)
extractor.model.eval()

def run_model(text: str) -> dict:
    started = time.perf_counter_ns()
    with torch.inference_mode():
        value = extractor(text, return_tensors=True, truncation=True, max_length=32)
    latency_ms = (time.perf_counter_ns() - started) / 1_000_000
    tensor = value if isinstance(value, torch.Tensor) else torch.as_tensor(value)
    tensor = tensor.detach().cpu().contiguous()
    if tensor.ndim != 3 or tensor.shape[0] != 1:
        raise RuntimeError(f"UNEXPECTED_FEATURE_SHAPE:{tuple(tensor.shape)}")
    if not torch.isfinite(tensor).all().item():
        raise RuntimeError("NON_FINITE_MODEL_OUTPUT")
    return {
        "model_revision_sha": MODEL_REVISION,
        "weights_sha256": WEIGHTS_SHA256,
        "output_sha256": hashlib.sha256(tensor.numpy().tobytes()).hexdigest(),
        "output_shape": list(tensor.shape),
        "output_preview": [round(float(v), 8) for v in tensor[0, 0, :8]],
        "inference_latency_ms": round(latency_ms, 3),
        "task": "feature-extraction",
        "model_type": extractor.model.config.model_type,
        "tokenizer": type(extractor.tokenizer).__name__,
        "torch_version": torch.__version__,
        "transformers_version": transformers.__version__,
        "backend": "PyTorch CPU",
    }

class GraphState(TypedDict, total=False):
    trace_id: str
    correlation_id: str
    input_text: str
    authorized_by: str
    authorized: bool
    graph_nodes: list[str]
    model_revision_sha: str
    weights_sha256: str
    output_sha256: str
    output_shape: list[int]
    output_preview: list[float]
    inference_latency_ms: float
    task: str
    model_type: str
    tokenizer: str
    torch_version: str
    transformers_version: str
    backend: str

def authorize_node(state: GraphState) -> dict:
    if not state.get("trace_id") or not state.get("correlation_id"):
        raise RuntimeError("TRACE_AND_CORRELATION_REQUIRED")
    if state.get("authorized_by") != "N07.Mesh":
        raise RuntimeError("N07_MESH_AUTHORIZATION_REQUIRED")
    return {"authorized": True, "graph_nodes": ["N07.Mesh.authorization"]}

def local_inference_node(state: GraphState) -> dict:
    if not state.get("authorized"):
        raise RuntimeError("LANGGRAPH_AUTHORIZATION_NOT_SATISFIED")
    result = run_model(state.get("input_text") or DEFAULT_TEXT)
    result["graph_nodes"] = list(state.get("graph_nodes") or []) + ["Transformers.PyTorch.CPU"]
    return result

builder = StateGraph(GraphState)
builder.add_node("authorize", authorize_node)
builder.add_node("local_inference", local_inference_node)
builder.add_edge(START, "authorize")
builder.add_edge("authorize", "local_inference")
builder.add_edge("local_inference", END)
checkpointer = MemorySaver()
graph = builder.compile(checkpointer=checkpointer)

print(json.dumps({
    "ready": True,
    "model_id": MODEL_ID,
    "model_revision_sha": MODEL_REVISION,
    "weights_sha256": WEIGHTS_SHA256,
    "backend": "PyTorch CPU",
    "task": "feature-extraction",
    "tokenizer": type(extractor.tokenizer).__name__,
    "langgraph_version": importlib.metadata.version("langgraph"),
    "checkpointer": type(checkpointer).__name__,
}, separators=(",", ":")), flush=True)

for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    try:
        request = json.loads(line)
        if request.get("mode") == "shutdown":
            break
        mode = request.get("mode")
        trace_id = request.get("trace_id", "")
        correlation_id = request.get("correlation_id", "")
        text = request.get("input_text") or DEFAULT_TEXT
        if mode == "direct":
            result = run_model(text)
            response = {
                **result,
                "trace_id": trace_id,
                "correlation_id": correlation_id,
                "checkpoint_written": False,
                "checkpointer": "none",
                "graph_nodes": [],
            }
        elif mode == "graph":
            config = {"configurable": {"thread_id": correlation_id}}
            initial: GraphState = {
                "trace_id": trace_id,
                "correlation_id": correlation_id,
                "input_text": text,
                "authorized_by": request.get("authorized_by", ""),
            }
            final = graph.invoke(initial, config=config)
            snapshot_state = graph.get_state(config)
            checkpoint_written = bool(
                snapshot_state.values.get("output_sha256")
                and snapshot_state.values.get("trace_id") == trace_id
                and snapshot_state.values.get("correlation_id") == correlation_id
            )
            if not checkpoint_written:
                raise RuntimeError("LANGGRAPH_CHECKPOINT_NOT_PERSISTED")
            response = {
                **final,
                "trace_id": trace_id,
                "correlation_id": correlation_id,
                "checkpoint_written": checkpoint_written,
                "checkpointer": type(checkpointer).__name__,
            }
        else:
            raise RuntimeError(f"UNSUPPORTED_MODE:{mode}")
        print(json.dumps(response, separators=(",", ":")), flush=True)
    except Exception as exc:
        print(json.dumps({"error": f"{type(exc).__name__}:{exc}"}, separators=(",", ":")), flush=True)
