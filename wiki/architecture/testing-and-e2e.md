# Testing and E2E

TRAX ships unit tests plus a compose-backed E2E suite.

## Unit Tests

Important package test surfaces:

- `pkg/trax`
- `pkg/mq/common`
- `pkg/common`
- `pkg/cache`

The current code requires a modern Go toolchain. A stale shell Go version will fail before meaningful TRAX verification.

## E2E Tests

The standalone E2E suite lives in `tests/e2e/trax` and uses `tests/e2e/common` helpers.

The standalone harness now treats TRAX as self-contained:

- per-test databases are created dynamically;
- only the base `trax` schema plus the `test_cluster` seed are initialized;
- saga templates are created through `traxcli` inside the `traxcli-submitter` container instead of
  by loading non-TRAX SQL.

It covers:

- smoke template setup;
- successful multi-step saga execution;
- compensation;
- topology/routing;
- idempotency;
- deep sub-saga execution;
- hierarchy queries.

The compose environment includes PostgreSQL, RabbitMQ, Redis, `traxctrl`, multiple `traxcoord` instances, executor workers, and a test runner.

See [Testing and E2E Operations](../operations/testing.md) for commands and operational notes.
