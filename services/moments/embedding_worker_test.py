"""Boundary tests for the optional local embedding process."""

import io
import json
import pathlib
import sys
import unittest
from types import SimpleNamespace
from unittest.mock import Mock, patch

sys.path.insert(0, str(pathlib.Path(__file__).parent))
import embed_text  # noqa: E402


class EmbeddingWorkerTests(unittest.TestCase):
    """Check requests before model dependencies are imported or loaded."""

    def test_rejects_oversized_text(self) -> None:
        """An oversized sentence must not be passed to the model."""
        request = json.dumps({"texts": ["x" * 2049]}).encode() + b"\n"
        with patch.object(sys, "stdin", io.TextIOWrapper(io.BytesIO(request))):
            with self.assertRaisesRegex(ValueError, "invalid text"):
                embed_text.main()

    def test_accepts_non_latin_text(self) -> None:
        """A Ukrainian sentence passes the input boundary unchanged."""
        query = "task: search result | query: маленька дівчинка і квіти"
        request = json.dumps({"texts": [query]}).encode() + b"\n"
        output = io.StringIO()
        with patch.object(sys, "stdin", io.TextIOWrapper(io.BytesIO(request))):
            with patch.object(sys, "stdout", output):
                with patch.object(embed_text, "embed", return_value=["vector"]):
                    embed_text.main()
        self.assertEqual(json.loads(output.getvalue()), {"vectors": ["vector"]})

    def test_rejects_unknown_model(self) -> None:
        """An unknown model is rejected before dependencies are loaded."""
        request = b'{"model":"unknown","texts":["query: sky"]}\n'
        with patch.object(sys, "stdin", io.TextIOWrapper(io.BytesIO(request))):
            with self.assertRaisesRegex(ValueError, "invalid embedding model"):
                embed_text.main()

    def test_e5_supplies_required_segment_ids(self) -> None:
        """E5 receives zero segment IDs when its tokenizer omits that required input."""
        import numpy as np

        encoded = {
            "input_ids": np.array([[1, 2]], dtype=np.int64),
            "attention_mask": np.array([[1, 1]], dtype=np.int64),
        }
        tokenizer = Mock(return_value=encoded)
        session = Mock()
        session.get_inputs.return_value = [
            SimpleNamespace(name=name)
            for name in ("input_ids", "attention_mask", "token_type_ids")
        ]
        session.run.return_value = [np.ones((1, 2, 384), dtype=np.float32)]
        modules = {
            "transformers": SimpleNamespace(
                AutoTokenizer=SimpleNamespace(from_pretrained=Mock(return_value=tokenizer))
            ),
            "onnxruntime": SimpleNamespace(
                SessionOptions=Mock,
                InferenceSession=Mock(return_value=session),
            ),
        }
        with patch.dict(sys.modules, modules):
            embed_text.embed(["passage: sky"], "e5")
        self.assertTrue(
            np.array_equal(
                session.run.call_args.args[1]["token_type_ids"],
                np.zeros_like(encoded["input_ids"]),
            )
        )


if __name__ == "__main__":
    unittest.main()
