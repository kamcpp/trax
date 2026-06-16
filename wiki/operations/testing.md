# Testing And E2E Operations

TRAX has unit tests, package-level tests, and compose-backed E2E tests.

## Unit And Package Tests

Core packages with tests include:

- `pkg/trax`
- `pkg/mq/common`
- `pkg/common`
- `pkg/cache`

TRAX requires a modern Go toolchain. Older Go versions fail before meaningful TRAX verification
because dependencies use packages such as `cmp`, `slices`, `maps`, `log/slog`, and `crypto/ecdh`.

## TRAX E2E Suite

The dedicated TRAX E2E suite lives under `tests/e2e/trax`.

It uses:

- PostgreSQL
- RabbitMQ
- Redis
- `traxctrl`
- multiple `traxcoord` services
- `traxcli executor` workers
- a Go test runner

The standalone TRAX harness initializes only:

- `deploy/k8s/init/init_trax_pgsql.sql`
- `tests/e2e/trax/init_test_cluster.sql`

It does not depend on non-TRAX schemas or business services.

Template setup and saga submission are driven through `traxcli` commands executed inside the `traxcli-submitter` container.

Covered scenarios include:

- smoke template creation and submission;
- seven-step successful saga;
- compensation flow;
- deep sub-saga execution;
- saga hierarchy queries;
- topology/routing behavior;
- idempotency behavior.

## Test Isolation

The E2E harness includes support for:

- per-run environment management;
- RabbitMQ readiness checks;
- database helpers;
- service checks;
- result capture;
- Docker info, logs, and database dump scripts.

Testing-only endpoints exist for database switching:

- `traxctrl`: `POST /api/v1/experimental/testing/setdbname`
- `traxcoord`: `POST /api/v1/experimental/testing/setdbname`

These are test/admin affordances and must remain gated outside normal production operation.

## Coverage Direction

TRAX E2E should cover the distributed workflow runtime directly: routing, retries, idempotency,
compensation, sub-saga hierarchy, result capture, and operator inspection. Business-specific test
matrices belong outside this repository.
