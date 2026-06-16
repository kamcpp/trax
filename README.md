# TRAX

TRAX is a standalone distributed workflow and saga orchestration system.

It owns:

- durable saga and step state;
- template and cluster management;
- RabbitMQ-based distributed step routing;
- coordinator and control-plane daemons;
- a generic executor and submitter CLI surface;
- TRAX-focused unit and end-to-end tests.

This repository is now the canonical home for TRAX. Other systems should depend on TRAX through its Go packages, APIs, and runtime binaries rather than carrying private copies of the subsystem.

## Main Entry Points

- Wiki: `wiki/index.md`
- Architecture: `wiki/architecture/v1.md`
- Data model: `wiki/data-model/postgresql.md`
- Runtime flow: `wiki/flows/saga-lifecycle.md`
- API surface: `wiki/reference/api-surface.md`
- TRAX E2E tests: `tests/e2e/trax/README.md`

## Current Runtime Shape

- `traxcoord`: coordinator daemon that advances workflows.
- `traxctrl`: read/control daemon for templates, clusters, saga status, annexes, hierarchy, and operator overrides.
- `traxcli`: CLI for templates, submitter, executor, and watch flows.
- PostgreSQL: durable state authority.
- RabbitMQ: execution transport.
- Redis/cache: distributed mutex support.

## Local Wiki

Run the wiki server with:

```bash
make wiki
```

Default port is `3334`.
