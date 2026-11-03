"""Embed bounded search text with a pinned local multilingual E5 ONNX model.

The service starts this process on demand. It accepts one JSON request on stdin,
prints one JSON response, and exits; no media bytes, paths, tokens, or database
handles enter the model process.
"""

import base64
import json
import os
import pathlib
import sys

MODEL_DIR = pathlib.Path("/usr/local/lib/lite-nas/moments/e5-model/onnx")
MAX_TEXTS = 10001
MAX_TEXT_BYTES = 2048


def embed(texts: list[str]) -> list[str]:
    """Return normalized float32 vectors encoded as base64 in input order."""
    import numpy as np
    import onnxruntime as ort
    from transformers import AutoTokenizer

    tokenizer = AutoTokenizer.from_pretrained(
        MODEL_DIR, local_files_only=True, trust_remote_code=False
    )
    options = ort.SessionOptions()
    options.intra_op_num_threads = 2
    options.inter_op_num_threads = 1
    session = ort.InferenceSession(
        str(MODEL_DIR / "model.onnx"),
        providers=["CPUExecutionProvider"],
        sess_options=options,
    )
    input_names = {item.name for item in session.get_inputs()}
    vectors: list[str] = []
    for start in range(0, len(texts), 16):
        encoded = tokenizer(
            texts[start : start + 16],
            padding=True,
            truncation=True,
            max_length=256,
            return_tensors="np",
        )
        inputs = {name: value.astype(np.int64) for name, value in encoded.items() if name in input_names}
        hidden = session.run(None, inputs)[0]
        mask = encoded["attention_mask"][..., None].astype(np.float32)
        pooled = (hidden * mask).sum(axis=1) / np.maximum(mask.sum(axis=1), 1)
        lengths = np.linalg.norm(pooled, axis=1, keepdims=True)
        normalized = (pooled / np.maximum(lengths, 1e-12)).astype("<f4")
        vectors.extend(base64.b64encode(row.tobytes()).decode("ascii") for row in normalized)
    return vectors


def main() -> None:
    """Validate the single local request before loading model dependencies."""
    os.environ["HF_HUB_OFFLINE"] = "1"
    os.environ["TRANSFORMERS_OFFLINE"] = "1"
    request = json.loads(sys.stdin.buffer.readline(21 * 1024 * 1024))
    texts = request.get("texts")
    if not isinstance(texts, list) or not 1 <= len(texts) <= MAX_TEXTS:
        raise ValueError("invalid text count")
    if any(
        not isinstance(text, str)
        or not text
        or len(text.encode("utf-8")) > MAX_TEXT_BYTES
        for text in texts
    ):
        raise ValueError("invalid text")
    print(json.dumps({"vectors": embed(texts)}, separators=(",", ":")), flush=True)


if __name__ == "__main__":
    main()
