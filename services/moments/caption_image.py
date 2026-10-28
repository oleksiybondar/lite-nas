"""Optional local BLIP caption worker for Moments image scans.

Deployment installs Pillow, PyTorch, transformers 4.x, and a pinned BLIP base
model under Moments storage. No model download is attempted while the service
runs.
"""

import base64
import binascii
import io
import json
import os
import sys
import warnings

MAX_IMAGE_PIXELS = 50_000_000
MAX_REQUEST_BYTES = 24 * 1024 * 1024
SUPPORTED_FORMATS = ("JPEG", "PNG", "WEBP", "GIF", "BMP", "TIFF")


def decode_image(encoded: str):
    """Decode one bounded still image for local caption inference."""
    from PIL import Image

    Image.MAX_IMAGE_PIXELS = MAX_IMAGE_PIXELS
    warnings.simplefilter("error", Image.DecompressionBombWarning)
    if not isinstance(encoded, str) or len(encoded) > MAX_REQUEST_BYTES:
        raise ValueError("invalid image input")
    data = base64.b64decode(encoded, validate=True)
    if not data or len(data) > 16 * 1024 * 1024:
        raise ValueError("image exceeds byte limit")
    with Image.open(io.BytesIO(data), formats=SUPPORTED_FORMATS) as source:
        if source.width * source.height > MAX_IMAGE_PIXELS:
            raise ValueError("image exceeds pixel limit")
        source.thumbnail((1024, 1024))
        return source.convert("RGB")


def main() -> None:
    """Load local model weights once and caption newline-delimited image requests."""
    try:
        from PIL import Image
        from transformers import BlipForConditionalGeneration, BlipProcessor

        model_dir = os.environ.get(
            "LITE_NAS_BLIP_MODEL_DIR", "/usr/local/lib/lite-nas/moments/blip-model"
        )
        processor = BlipProcessor.from_pretrained(model_dir, local_files_only=True)
        model = BlipForConditionalGeneration.from_pretrained(
            model_dir, local_files_only=True, use_safetensors=True
        )
        model.eval()
    except (ImportError, OSError):
        print('{"ready":false}', flush=True)
        return

    print('{"ready":true}', flush=True)
    while line := sys.stdin.buffer.readline(MAX_REQUEST_BYTES + 1):
        if len(line) > MAX_REQUEST_BYTES or not line.endswith(b"\n"):
            print('{"caption":""}', flush=True)
            return
        try:
            request = json.loads(line)
            image = decode_image(request["image"])
            inputs = processor(image, return_tensors="pt")
            output = model.generate(**inputs, max_new_tokens=40)
            caption = processor.decode(output[0], skip_special_tokens=True).strip()
            print(json.dumps({"caption": caption}), flush=True)
        except (
            KeyError,
            ValueError,
            OSError,
            TypeError,
            binascii.Error,
            Image.DecompressionBombWarning,
            Image.DecompressionBombError,
        ):
            print('{"caption":""}', flush=True)


if __name__ == "__main__":
    main()
