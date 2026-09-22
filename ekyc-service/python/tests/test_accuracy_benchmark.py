"""Accuracy benchmark for the real EasyOCR/face_recognition models
extract_id.py wraps - as opposed to test_extract_id.py, which only checks
input validation and never imports the heavy libraries.

IMPORTANT about "accuracy": there is no way to mathematically guarantee a
fixed accuracy number for a wrapper script over third-party ML models
against arbitrary real-world input - that number depends on the model,
the specific images fed to it, and drifts as those libraries are
upgraded. What this module gives instead is a *reproducible regression
benchmark* against a fixed, checked-in fixture set (see fixtures/NOTICE.md
for provenance): it measures actual accuracy on every run and fails loudly
if it drops below the asserted floor, so a library upgrade or a code
change that silently degrades quality gets caught - not a promise about
photos these fixtures never saw.

Requires easyocr/face_recognition to be installed (they're heavy - see
requirements.txt); every test here is skipped, not failed, when they
aren't available, so test_extract_id.py's validation suite still runs
in a lightweight environment.
"""
import os

import pytest

import extract_id as m

easyocr = pytest.importorskip("easyocr")
face_recognition = pytest.importorskip("face_recognition")

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")


def _char_accuracy(expected: str, got: str) -> float:
    """1 - (Levenshtein distance / len(expected)), clamped to [0,1] - the
    standard character-level accuracy metric for OCR benchmarks."""
    if not expected:
        return 1.0 if not got else 0.0

    prev_row = list(range(len(got) + 1))
    for i, ec in enumerate(expected, start=1):
        cur_row = [i] + [0] * len(got)
        for j, gc in enumerate(got, start=1):
            cost = 0 if ec == gc else 1
            cur_row[j] = min(
                prev_row[j] + 1,       # deletion
                cur_row[j - 1] + 1,    # insertion
                prev_row[j - 1] + cost,  # substitution
            )
        prev_row = cur_row
    distance = prev_row[-1]
    return max(0.0, 1.0 - distance / len(expected))


def _render_text_png_b64(text: str, width=400, height=80, font_size=36) -> str:
    import base64
    import io

    from PIL import Image, ImageDraw, ImageFont

    img = Image.new("RGB", (width, height), (255, 255, 255))
    draw = ImageDraw.Draw(img)
    try:
        font = ImageFont.truetype("/System/Library/Fonts/Menlo.ttc", font_size)
    except OSError:
        font = ImageFont.load_default()
    draw.text((10, 10), text, fill=(0, 0, 0), font=font)

    buf = io.BytesIO()
    img.save(buf, format="PNG")
    return base64.b64encode(buf.getvalue()).decode()


# ---------------------------------------------------------------------
# OCR accuracy
# ---------------------------------------------------------------------


def test_ocr_accuracy_natural_text():
    """EasyOCR's home turf: clean, rendered natural-language text. This
    is the baseline sanity check - if this doesn't hit ~90%+, something
    is wrong with the pipeline (image decoding, reader init) rather than
    with MRZ text specifically."""
    expected = "HELLO WORLD 12345"
    resp = m.mode_ocr({"image_b64": _render_text_png_b64(expected)})
    recognized = " ".join(line["text"] for line in resp["lines"]).strip().upper()

    accuracy = _char_accuracy(expected, recognized)
    print(f"\n[benchmark] natural text: expected={expected!r} got={recognized!r} accuracy={accuracy:.2%}")
    assert accuracy >= 0.90, (
        f"natural-text OCR accuracy {accuracy:.2%} fell below the 90% floor "
        f"(expected {expected!r}, got {recognized!r})"
    )


def test_ocr_accuracy_mrz_style_text():
    """MRZ text is a harder case for a general-purpose OCR model than
    natural language: dense uppercase+digit+'<' strings with no word
    boundaries, in a font EasyOCR wasn't trained on specifically (real
    MRZ printing uses OCR-B, chosen for exactly this kind of machine
    readability - see the native cardreader path's Readme note on the
    same tradeoff). This measures and reports the real number rather
    than assuming it matches the natural-text case."""
    expected = "IDVNMC1234567X3<<<<<<<<<<<<<<<"
    resp = m.mode_ocr({"image_b64": _render_text_png_b64(expected, width=600)})
    recognized = "".join(line["text"] for line in resp["lines"]).strip().upper()

    accuracy = _char_accuracy(expected, recognized)
    print(f"\n[benchmark] MRZ-style text: expected={expected!r} got={recognized!r} accuracy={accuracy:.2%}")
    assert accuracy >= 0.70, (
        f"MRZ-style OCR accuracy {accuracy:.2%} fell below the 70% floor "
        f"(expected {expected!r}, got {recognized!r}) - this is the real, "
        f"measured EasyOCR accuracy on this fixture, not a design target; "
        f"see the module docstring on why MRZ text is a harder case, and "
        f"prefer the native cardreader path (or a provider trained on "
        f"OCR-B) when MRZ accuracy matters more than this floor"
    )


# ---------------------------------------------------------------------
# Face embedding/comparison accuracy
# ---------------------------------------------------------------------


def _embed(path):
    with open(path, "rb") as f:
        import base64

        b64 = base64.b64encode(f.read()).decode()
    resp = m.mode_face_embed({"image_b64": b64})
    assert "embedding" in resp, f"expected a face embedding for {path}, got {resp}"
    return resp["embedding"]


def test_face_same_person_matches():
    a = _embed(os.path.join(FIXTURES_DIR, "person_a_1.jpg"))
    b = _embed(os.path.join(FIXTURES_DIR, "person_a_2.jpg"))
    resp = m.mode_face_compare({"embedding_a": a, "embedding_b": b})
    print(f"\n[benchmark] same-person distance={resp['distance']:.4f} match={resp['match']}")
    assert resp["match"] is True, f"expected two photos of the same person to match, got {resp}"


def test_face_different_person_does_not_match():
    a = _embed(os.path.join(FIXTURES_DIR, "person_a_1.jpg"))
    b = _embed(os.path.join(FIXTURES_DIR, "person_b_1.jpg"))
    resp = m.mode_face_compare({"embedding_a": a, "embedding_b": b})
    print(f"\n[benchmark] different-person distance={resp['distance']:.4f} match={resp['match']}")
    assert resp["match"] is False, f"expected two different people not to match, got {resp}"


def test_face_self_comparison_near_zero_distance():
    path = os.path.join(FIXTURES_DIR, "person_a_1.jpg")
    a = _embed(path)
    b = _embed(path)
    resp = m.mode_face_compare({"embedding_a": a, "embedding_b": b})
    assert resp["distance"] < 0.01, f"expected ~0 distance re-embedding the same image, got {resp}"


def test_face_no_face_in_blank_image(blank_image_b64):
    resp = m.mode_face_embed({"image_b64": blank_image_b64})
    assert resp == {"error": "no face detected"}


def test_ocr_blank_image_returns_no_lines(blank_image_b64):
    resp = m.mode_ocr({"image_b64": blank_image_b64})
    assert resp["lines"] == []
