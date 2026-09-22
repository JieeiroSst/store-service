# ekyc-service Python sidecar

`extract_id.py` is the stdin/stdout JSON CLI the Go `pyextract` adapter
(`internal/adapter/secondary/pyextract`) shells out to when
`CARD_READER_PROVIDER=python` / `FACE_ANALYZER_PROVIDER=python` - see that
package's doc comment and `extract_id.py`'s own docstring for the request/
response protocol.

## Setup

```
pip install dlib-bin==20.0.1 face-recognition-models==0.3.0
pip install --no-deps face_recognition==1.3.0
pip install click==8.1.7 numpy==1.26.4 pillow==10.4.0 easyocr==1.7.2
pip install -r requirements-dev.txt
```

Installing `requirements.txt` directly with a plain `pip install -r` will
try to compile the real `dlib` from source - see that file's header
comment for why, and follow the exact order above (or just `docker build`
the service, which does the same thing).

## Testing

```
pytest
```

Two layers, kept separate so the fast one can run anywhere:

- **`test_extract_id.py`** - input validation and request dispatch. Every
  malformed/edge-case request the Go adapter (or a bug in it) could send:
  bad JSON, wrong types, missing fields, invalid base64, non-image bytes,
  wrong-length embeddings, unknown mode, and so on. Runs without
  easyocr/face_recognition installed - validation happens before those
  modules are imported - so this is the tier to run in a fast lint/CI
  step.
- **`test_accuracy_benchmark.py`** - exercises the real models against
  fixed fixtures (see `tests/fixtures/NOTICE.md` for their provenance)
  and asserts a measured accuracy floor: natural-text OCR (≥90%), face
  same/different-person classification, and a documented, genuinely-lower
  floor for MRZ-style text specifically (EasyOCR is a general-purpose
  model, not trained on the OCR-B font real MRZ printing uses - see that
  test's docstring). Skips automatically if easyocr/face_recognition
  aren't installed.

**On "90% accuracy":** that number is a property of the underlying models
against real-world input, not something a wrapper script can mathematically
guarantee in the abstract. What's here instead is honest: every claim is
tied to a specific, reproducible, checked-in fixture and measured on every
test run, so a regression (library upgrade, code change) fails loudly
instead of silently. The MRZ-specific benchmark's floor is set to what
EasyOCR actually measures on that fixture, not to 90% - if MRZ accuracy
specifically needs to be higher, the native `cardreader` path (tuned
specifically for the MRZ font/layout) or a provider trained on OCR-B is
the better fit; see the root `Readme`'s comparison of the two providers.
