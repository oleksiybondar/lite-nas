"""Embed bounded search text with a pinned local E5 or EmbeddingGemma model.

Deployment smoke checks use one-shot mode. Moments starts the resident mode,
which loads the configured weights once and exchanges newline-delimited JSON
through stdin and stdout.
"""

import base64
import json
import os
import pathlib
import sys

MODEL_ROOT = pathlib.Path("/usr/local/lib/lite-nas/moments")
MAX_TEXTS = 10001
MAX_TEXT_BYTES = 2048
EMBEDDING_SIZE = 768


def load_model(model: str):
    """Load the configured tokenizer and ONNX session once for a resident worker."""
    from transformers import AutoTokenizer
    # Import Transformers first: loading ONNX Runtime first registers ONNX's
    # protobuf descriptors twice with this pinned dependency set and aborts.
    import onnxruntime as ort

    model_dir = MODEL_ROOT / ("e5-model/onnx" if model == "e5" else "embeddinggemma-model")
    model_file = model_dir / ("model.onnx" if model == "e5" else "onnx/model_q4.onnx")
    tokenizer = AutoTokenizer.from_pretrained(model_dir, local_files_only=True, trust_remote_code=False)
    options = ort.SessionOptions()
    options.intra_op_num_threads = 2
    options.inter_op_num_threads = 1
    session = ort.InferenceSession(
        str(model_file),
        providers=["CPUExecutionProvider"],
        sess_options=options,
    )
    input_names = {item.name for item in session.get_inputs()}
    return tokenizer, session, input_names


def embed(texts: list[str], model: str = "e5", loaded=None) -> list[str]:
    """Return normalized float32 vectors encoded as base64 in input order."""
    import numpy as np
    tokenizer, session, input_names = loaded or load_model(model)
    vectors: list[str] = []
    for start in range(0, len(texts), 16):
        encoded = tokenizer(
            texts[start : start + 16],
            padding=True,
            truncation=True,
            max_length=256 if model == "e5" else 512,
            return_tensors="np",
        )
        inputs = {name: value for name, value in encoded.items() if name in input_names}
        # The pinned E5 ONNX export requires segment IDs, but its tokenizer omits them.
        if (
            model == "e5"
            and "token_type_ids" in input_names
            and "token_type_ids" not in inputs
        ):
            inputs["token_type_ids"] = np.zeros_like(encoded["input_ids"])
        outputs = session.run(None, inputs)
        if model == "e5":
            mask = encoded["attention_mask"][..., None].astype(np.float32)
            embeddings = (outputs[0] * mask).sum(axis=1) / np.maximum(mask.sum(axis=1), 1)
            size = 384
        else:
            if len(outputs) < 2:
                raise ValueError("embedding model did not return sentence embeddings")
            embeddings = outputs[1]
            size = EMBEDDING_SIZE
        if embeddings.shape != (len(texts[start : start + 16]), size):
            raise ValueError("embedding model returned an invalid vector shape")
        lengths = np.linalg.norm(embeddings, axis=1, keepdims=True)
        if not np.isfinite(embeddings).all() or np.any(lengths <= 1e-12):
            raise ValueError("embedding model returned invalid vectors")
        normalized = (embeddings / lengths).astype("<f4")
        vectors.extend(base64.b64encode(row.tobytes()).decode("ascii") for row in normalized)
    return vectors


def main() -> None:
    """Serve one-shot deployment checks or a persistent configured model worker."""
    os.environ["HF_HUB_OFFLINE"] = "1"
    os.environ["TRANSFORMERS_OFFLINE"] = "1"
    if len(sys.argv) == 3 and sys.argv[1] == "--serve":
        serve(sys.argv[2])
        return
    request = json.loads(sys.stdin.buffer.readline(21 * 1024 * 1024))
    texts = request.get("texts")
    model = request.get("model", "e5")
    if model not in ("e5", "embeddinggemma"):
        raise ValueError("invalid embedding model")
    if not isinstance(texts, list) or not 1 <= len(texts) <= MAX_TEXTS:
        raise ValueError("invalid text count")
    if any(
        not isinstance(text, str)
        or not text
        or len(text.encode("utf-8")) > MAX_TEXT_BYTES
        for text in texts
    ):
        raise ValueError("invalid text")
    print(json.dumps({"vectors": embed(texts, model)}, separators=(",", ":")), flush=True)


def validate_request(request: dict, model: str) -> list[str]:
    """Validate one bounded request before passing text to the resident model."""
    texts = request.get("texts")
    if request.get("model", model) != model:
        raise ValueError("request model does not match resident model")
    if not isinstance(texts, list) or not 1 <= len(texts) <= MAX_TEXTS:
        raise ValueError("invalid text count")
    if any(
        not isinstance(text, str)
        or not text
        or len(text.encode("utf-8")) > MAX_TEXT_BYTES
        for text in texts
    ):
        raise ValueError("invalid text")
    return texts


def serve(model: str) -> None:
    """Keep one embedding model resident and answer newline-delimited requests."""
    if model not in ("e5", "embeddinggemma"):
        raise ValueError("invalid embedding model")
    loaded = load_model(model)
    print('{"ready":true}', flush=True)
    while line := sys.stdin.buffer.readline(21 * 1024 * 1024):
        try:
            request = json.loads(line)
            vectors = embed(validate_request(request, model), model, loaded)
            response = {"vectors": vectors}
        except Exception as error:  # Keep a bad input or inference failure from desynchronizing the stream.
            print(f"embedding request failed: {type(error).__name__}: {error}", file=sys.stderr)
            response = {"error": "embedding request failed"}
        print(json.dumps(response, separators=(",", ":")), flush=True)


if __name__ == "__main__":
    main()
