# TRAX Resilience TODO

This page tracks resilience work owned by TRAX.

## Implemented Or Substantially Present

- Coordinator readiness checks running state, DB circuit state, and RabbitMQ connection health.
- Submitter announcement has fast exponential backoff before falling back to the normal announcement interval.
- Store interface has update/delete methods for saga templates and saga-step templates.
- PostgreSQL store emits `trax_template_events` on template insert/update/delete.
- Store listener supports multiple channels.
- Coordinator has notification fanout and subscribes separately for saga events and template events.
- Template reload reacts to `trax_template_events` and keeps periodic polling as a fallback.
- `traxctrl` exposes template CRUD and step-template CRUD endpoints.
- E2E suite includes idempotency, topology, compensation, deep sub-saga, hierarchy, and seven-step saga scenarios.

## Still Open

- Run the full standalone unit suite with a modern Go toolchain and record results.
- Run standalone compose-backed TRAX E2E and record results.
- Audit the current resilience behavior against [Core Requirements](../reference/core-requirements.md).
- Add generated Swagger docs to the standard build path so image builds do not fail on missing `gen-docs` packages.
- Tighten testing endpoint gating and document the exact enabling env vars.
- Add targeted RabbitMQ reconnect, duplicate-delivery, and stale-consumer tests where coverage is thin.
- Normalize deployment/test naming to TRAX-owned database and service names.
- Keep example templates generic and move any business-specific examples out of TRAX.

## Related Wiki Pages

- [Architecture v1](../architecture/v1.md)
- [Template Management and Hot Reload](../flows/template-management.md)
- [Testing and E2E Operations](../operations/testing.md)
- [Core Requirements](../reference/core-requirements.md)
