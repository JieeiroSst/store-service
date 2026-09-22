"""Exhaustive input-validation coverage for extract_id.py's request
dispatch (handle_request) and image decoding (_decode_image_b64) - every
case a malformed or edge-case request from the Go adapter could hit,
each asserted to fail with a specific ValidationError message rather than
an opaque crash. These run without easyocr/face_recognition installed
(validation happens before those modules are imported) - see
test_accuracy_benchmark.py for tests that exercise the real models.
"""
import base64
import json

import pytest

import extract_id as m


# ---------------------------------------------------------------------
# Request framing: not JSON, not an object, missing/invalid mode.
# ---------------------------------------------------------------------


def test_read_request_empty_stdin_rejected(monkeypatch):
    monkeypatch.setattr(m.sys, "stdin", _stdin(""))
    with pytest.raises(m.ValidationError, match="empty request"):
        m._read_request()


def test_read_request_invalid_json_rejected(monkeypatch):
    monkeypatch.setattr(m.sys, "stdin", _stdin("{not json"))
    with pytest.raises(m.ValidationError, match="not valid JSON"):
        m._read_request()


def test_read_request_non_object_rejected(monkeypatch):
    monkeypatch.setattr(m.sys, "stdin", _stdin("[1, 2, 3]"))
    with pytest.raises(m.ValidationError, match="must be a JSON object"):
        m._read_request()


def test_handle_request_missing_mode_rejected():
    with pytest.raises(m.ValidationError, match="'mode'"):
        m.handle_request({})


def test_handle_request_unknown_mode_rejected():
    with pytest.raises(m.ValidationError, match="'mode'"):
        m.handle_request({"mode": "not_a_real_mode"})


def test_handle_request_non_string_mode_rejected():
    with pytest.raises(m.ValidationError, match="'mode'"):
        m.handle_request({"mode": 42})


# ---------------------------------------------------------------------
# image_b64 field validation (shared by ocr and face_embed).
# ---------------------------------------------------------------------


@pytest.mark.parametrize("mode", ["ocr", "face_embed"])
def test_missing_image_b64_rejected(mode):
    with pytest.raises(m.ValidationError, match="image_b64"):
        m.handle_request({"mode": mode})


@pytest.mark.parametrize("mode", ["ocr", "face_embed"])
def test_non_string_image_b64_rejected(mode):
    with pytest.raises(m.ValidationError, match="image_b64"):
        m.handle_request({"mode": mode, "image_b64": 12345})


@pytest.mark.parametrize("mode", ["ocr", "face_embed"])
def test_empty_string_image_b64_rejected(mode):
    with pytest.raises(m.ValidationError, match="image_b64"):
        m.handle_request({"mode": mode, "image_b64": ""})


def test_invalid_base64_rejected():
    with pytest.raises(m.ValidationError, match="not valid base64"):
        m._decode_image_b64("not-valid-base64!!!")


def test_valid_base64_non_image_rejected():
    garbage = base64.b64encode(b"this is not an image, just plain bytes").decode()
    with pytest.raises(m.ValidationError, match="not a decodable image"):
        m._decode_image_b64(garbage)


def test_valid_base64_empty_payload_rejected():
    empty = base64.b64encode(b"").decode()
    with pytest.raises(m.ValidationError, match="zero bytes"):
        m._decode_image_b64(empty)


def test_valid_image_decodes_to_nonempty_array(blank_image_b64):
    array = m._decode_image_b64(blank_image_b64)
    assert array.size > 0
    assert array.ndim == 3  # H x W x RGB


# ---------------------------------------------------------------------
# face_compare field validation.
# ---------------------------------------------------------------------


def test_face_compare_missing_embeddings_rejected():
    with pytest.raises(m.ValidationError, match="embedding_a"):
        m.handle_request({"mode": "face_compare"})


def test_face_compare_wrong_length_embedding_rejected():
    with pytest.raises(m.ValidationError, match="embedding_a"):
        m.handle_request(
            {"mode": "face_compare", "embedding_a": [0.0, 1.0], "embedding_b": [0.0] * 128}
        )


def test_face_compare_non_numeric_embedding_rejected():
    with pytest.raises(m.ValidationError, match="embedding_a"):
        m.handle_request(
            {
                "mode": "face_compare",
                "embedding_a": ["not", "numbers"] + [0.0] * 126,
                "embedding_b": [0.0] * 128,
            }
        )


def test_face_compare_not_a_list_rejected():
    with pytest.raises(m.ValidationError, match="embedding_b"):
        m.handle_request(
            {"mode": "face_compare", "embedding_a": [0.0] * 128, "embedding_b": "nope"}
        )


# ---------------------------------------------------------------------
# main()'s stdin/stdout/exit-code contract, end to end (still without
# needing easyocr/face_recognition - these all fail at validation).
# ---------------------------------------------------------------------


def test_main_writes_error_json_and_exits_1_on_bad_input(monkeypatch, capsys):
    monkeypatch.setattr(m.sys, "stdin", _stdin("not json"))
    with pytest.raises(SystemExit) as exc_info:
        m.main()
    assert exc_info.value.code == 1
    out = json.loads(capsys.readouterr().out)
    assert "error" in out


def test_main_writes_error_json_for_unknown_mode(monkeypatch, capsys):
    monkeypatch.setattr(m.sys, "stdin", _stdin(json.dumps({"mode": "bogus"})))
    with pytest.raises(SystemExit) as exc_info:
        m.main()
    assert exc_info.value.code == 1
    out = json.loads(capsys.readouterr().out)
    assert "mode" in out["error"]


def _stdin(text):
    import io

    return io.StringIO(text)
