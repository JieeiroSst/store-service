#!/usr/bin/env python3
"""Stdin/stdout JSON CLI ekyc-service's Go pyextract adapter shells out to
(see internal/adapter/secondary/pyextract/client.go) for the higher-
accuracy, heavier alternative to the in-house cardreader/facebio
algorithms: EasyOCR (deep-learning OCR) for card text, and
face_recognition's dlib ResNet-based 128-d embedding for face biometrics.

Protocol: read exactly one JSON object from stdin, write exactly one JSON
object to stdout, exit 0. On failure, write {"error": "..."} and exit 1.
This process is spawned fresh per request by the Go adapter (see that
package's doc comment for why, and the tradeoff that implies).

Modes:
  ocr           {"image_b64": "..."}
                -> {"lines": [{"text": "...", "confidence": 0.0-1.0}, ...]}
                Raw recognized text lines with EasyOCR's own per-detection
                confidence - Go still does ICAO 9303 TD1 field parsing/
                checksum validation (internal/domain/mrz), this only
                recognizes characters. Detections below MIN_OCR_CONFIDENCE
                are dropped rather than returned as noise for the caller
                to filter itself.
  face_embed    {"image_b64": "..."}
                -> {"embedding": [128 floats], "face_location": [...],
                    "face_count": int, "quality_score": float}
                or {"error": "no face detected"} /
                   {"error": "multiple faces detected (N) - resubmit a
                    single-subject frame"}
  face_compare  {"embedding_a": [128 floats], "embedding_b": [128 floats]}
                -> {"distance": float, "match": bool}

Every mode validates its own required fields (type, length, range) up
front and raises ValidationError with a specific, actionable message
instead of letting a malformed request surface as an opaque KeyError/
TypeError/IndexError from deep inside a library call - see
tests/test_extract_id.py for the exhaustive case list this guards
against, and README.md for the accuracy benchmark that exercises the
actual OCR/face models against known-good fixtures.
"""
import base64
import binascii
import io
import json
import sys

# EasyOCR detections scoring below this are dropped as noise rather than
# handed to the Go side - a confident wrong-alphabet detection is exactly
# the kind of thing that would otherwise get selected as a fake "MRZ line"
# by internal/adapter/secondary/pyextract/cardreader.go's heuristic.
MIN_OCR_CONFIDENCE = 0.15

# The 128-d embedding face_recognition's ResNet model always produces.
FACE_EMBEDDING_LENGTH = 128


class ValidationError(ValueError):
    """A malformed request - reported the same way as any other failure
    (see main()), but as its own type so tests can assert specifically on
    "the input was rejected" versus "the input was fine but processing
    failed" (e.g. no face found)."""


def _read_request():
    raw = sys.stdin.read()
    if not raw.strip():
        raise ValidationError("empty request: expected a JSON object on stdin")
    try:
        req = json.loads(raw)
    except json.JSONDecodeError as exc:
        raise ValidationError(f"request is not valid JSON: {exc}") from exc
    if not isinstance(req, dict):
        raise ValidationError(f"request must be a JSON object, got {type(req).__name__}")
    return req


def _require_str(req, key):
    value = req.get(key)
    if not isinstance(value, str) or not value:
        raise ValidationError(f"'{key}' is required and must be a non-empty string")
    return value


def _require_embedding(req, key):
    value = req.get(key)
    if not isinstance(value, list) or len(value) != FACE_EMBEDDING_LENGTH:
        raise ValidationError(
            f"'{key}' must be a {FACE_EMBEDDING_LENGTH}-element array of numbers, "
            f"got {type(value).__name__} of length "
            f"{len(value) if isinstance(value, list) else 'n/a'}"
        )
    if not all(isinstance(v, (int, float)) for v in value):
        raise ValidationError(f"'{key}' must contain only numbers")
    return value


def _decode_image_b64(b64):
    from PIL import Image, UnidentifiedImageError
    import numpy as np

    try:
        data = base64.b64decode(b64, validate=True)
    except (binascii.Error, ValueError) as exc:
        raise ValidationError(f"'image_b64' is not valid base64: {exc}") from exc
    if not data:
        raise ValidationError("'image_b64' decoded to zero bytes")

    try:
        img = Image.open(io.BytesIO(data))
        img.load()  # Image.open is lazy; this forces the actual decode.
        img = img.convert("RGB")
    except (UnidentifiedImageError, OSError) as exc:
        raise ValidationError(f"'image_b64' is not a decodable image: {exc}") from exc

    array = np.array(img)
    if array.size == 0:
        raise ValidationError("decoded image has zero pixels")
    return array


def mode_ocr(req):
    image = _decode_image_b64(_require_str(req, "image_b64"))

    import easyocr

    # gpu=False: this runs inside a container with no GPU access - see the
    # Dockerfile's comment on why the CPU-only torch wheel is used.
    reader = easyocr.Reader(["en"], gpu=False, verbose=False)
    results = reader.readtext(image, detail=1, paragraph=False)

    # Sort top-to-bottom by each detection's bounding box vertical center,
    # since MRZ lines must be reassembled in reading order.
    def y_center(detection):
        bbox = detection[0]
        return sum(point[1] for point in bbox) / len(bbox)

    results.sort(key=y_center)
    lines = [
        {"text": text, "confidence": float(conf)}
        for (_bbox, text, conf) in results
        if conf >= MIN_OCR_CONFIDENCE
    ]
    return {"lines": lines}


def mode_face_embed(req):
    image = _decode_image_b64(_require_str(req, "image_b64"))

    import face_recognition

    locations = face_recognition.face_locations(image)
    if not locations:
        return {"error": "no face detected"}
    if len(locations) > 1:
        # An eKYC face-scan/card-portrait frame should contain exactly one
        # subject - ambiguity here is a capture problem worth surfacing to
        # the caller rather than silently guessing "the biggest face".
        return {
            "error": f"multiple faces detected ({len(locations)}) - resubmit a single-subject frame",
            "face_count": len(locations),
        }

    location = locations[0]
    encodings = face_recognition.face_encodings(image, known_face_locations=[location])
    if not encodings:
        return {"error": "face detected but encoding failed"}

    top, right, bottom, left = location
    face_area = max(bottom - top, 0) * max(right - left, 0)
    image_area = image.shape[0] * image.shape[1]
    # Fraction of the frame the face occupies is a cheap, dependency-free
    # proxy for capture quality: a face that's too small/far away yields
    # an unreliable embedding even when detection nominally "succeeds".
    quality_score = face_area / image_area if image_area > 0 else 0.0

    return {
        "embedding": encodings[0].tolist(),
        "face_location": list(location),
        "face_count": 1,
        "quality_score": quality_score,
    }


def mode_face_compare(req):
    a_raw = _require_embedding(req, "embedding_a")
    b_raw = _require_embedding(req, "embedding_b")

    import face_recognition
    import numpy as np

    a = np.array(a_raw, dtype=float)
    b = np.array(b_raw, dtype=float)
    distance = float(face_recognition.face_distance([a], b)[0])
    # face_recognition's own documented default match threshold.
    match = distance <= 0.6
    return {"distance": distance, "match": match}


MODES = {
    "ocr": mode_ocr,
    "face_embed": mode_face_embed,
    "face_compare": mode_face_compare,
}


def handle_request(req):
    """Dispatches a parsed request to its mode handler. Split out from
    main() so tests can exercise the dispatch/validation logic directly
    without going through stdin/stdout/sys.exit."""
    mode = req.get("mode")
    if not isinstance(mode, str) or mode not in MODES:
        raise ValidationError(f"'mode' must be one of {sorted(MODES)}, got {mode!r}")
    return MODES[mode](req)


def main():
    try:
        req = _read_request()
        result = handle_request(req)
    except Exception as exc:  # noqa: BLE001 - this is a CLI boundary, not library code
        json.dump({"error": str(exc)}, sys.stdout)
        sys.exit(1)

    json.dump(result, sys.stdout)


if __name__ == "__main__":
    main()
