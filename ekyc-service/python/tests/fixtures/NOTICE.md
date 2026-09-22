Test fixture provenance
=======================

- `sample_face.jpg` - from esimov/pigo (MIT licensed), `testdata/sample.jpg`.
- `person_a_1.jpg`, `person_a_2.jpg` - two distinct official White House
  portrait photos of the same public figure (U.S. federal government work,
  public domain), from ageitgey/face_recognition's own `examples/`
  directory (`obama.jpg`, `obama2.jpg`) - the library's own canonical
  same-person test pair.
- `person_b_1.jpg` - an official White House photo of a different public
  figure (same provenance), from the same directory (`biden.jpg`) - the
  different-person half of that same canonical pair.

Used only to benchmark face_recognition's same/different-person accuracy
in test_accuracy_benchmark.py - not used anywhere in application code.
