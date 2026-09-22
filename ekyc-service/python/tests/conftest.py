import base64
import io
import os
import sys

# extract_id.py is a standalone script (not an installed package), so make
# its directory importable the same way the Go adapter invokes it: by
# path, not by package name.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest  # noqa: E402


def _png_bytes(width=64, height=64, color=(200, 200, 200)):
    from PIL import Image

    img = Image.new("RGB", (width, height), color)
    buf = io.BytesIO()
    img.save(buf, format="PNG")
    return buf.getvalue()


@pytest.fixture
def blank_image_b64():
    """A trivially valid, decodable image with no face and no text -
    exercises the "processing succeeded but found nothing" path, as
    opposed to a malformed-input rejection."""
    return base64.b64encode(_png_bytes()).decode()


@pytest.fixture
def sample_face_path():
    """Path to a real sample photo with exactly one face, used by the
    accuracy-benchmark tests. Skips (not fails) when absent, since it's a
    binary fixture too large to script-generate and isn't checked into
    the repo - see tests/README in the benchmark test module for how to
    provide one locally/in CI."""
    path = os.environ.get(
        "EKYC_TEST_FACE_IMAGE",
        os.path.join(os.path.dirname(__file__), "fixtures", "sample_face.jpg"),
    )
    if not os.path.exists(path):
        pytest.skip(f"no sample face image at {path} (set EKYC_TEST_FACE_IMAGE to override)")
    return path
