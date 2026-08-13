# Notice

This service is a Go port of the API surface implemented by
[`@linkforty/core`](https://github.com/LinkForty/core) (TypeScript/Fastify),
licensed AGPL-3.0-only. This port is a derivative work and is distributed
under the same license -- see [LICENSE](LICENSE).

No source files were copied verbatim; behavior (routes, field names,
status codes, the fingerprint/attribution algorithm, HTML templates,
database schema) was re-implemented in Go from a reading of the upstream
source, preserving upstream's structure and comments where useful for
future maintainers to cross-reference the two implementations. See
`CONTRACT_TEST_CHECKLIST.md` for the list of intentional and unintentional
(flagged) deviations from upstream behavior.

Upstream project: https://github.com/LinkForty/core
Upstream license: AGPL-3.0-only (see LICENSE, identical text)
