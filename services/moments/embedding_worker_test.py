"""Boundary tests for the optional local embedding process."""

import io
import json
import pathlib
import sys
import unittest
from unittest.mock import patch

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
        request = json.dumps({"texts": ["query: собака"]}).encode() + b"\n"
        output = io.StringIO()
        with patch.object(sys, "stdin", io.TextIOWrapper(io.BytesIO(request))):
            with patch.object(sys, "stdout", output):
                with patch.object(embed_text, "embed", return_value=["vector"]):
                    embed_text.main()
        self.assertEqual(json.loads(output.getvalue()), {"vectors": ["vector"]})


if __name__ == "__main__":
    unittest.main()
