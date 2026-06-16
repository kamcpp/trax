# Current Gaps And Mismatches

This page lists current code, wiki, and deployment mismatches.

## Build And Generated Docs

The daemon API packages import generated Swagger packages under `gen-docs/...`. The current `Makefile` `swagger` target only prints a message. A real generation or committed-doc restoration path is needed.

## Placeholder Daemon Run Files

`pkg/daemons/traxcoord/run.go` and `pkg/daemons/traxctrl/run.go` are placeholder consumers and are
not the real startup path. They should be removed or clearly deprecated in code after confirming no
caller uses them.

## Database Naming

Deployment and E2E assets should consistently use TRAX-owned database names.

## Example Seed SQL Ownership

`deploy/k8s/init/{csd,exchange,prtagent,tldinfra}/min/trax.sql` should remain generic enough for
TRAX examples. Business-specific saga templates belong in dependent systems.

## Toolchain

The codebase requires modern Go. The repo should enforce and document the minimum supported
toolchain explicitly.

## Common Package Breadth

`pkg/common` should stay limited to helpers required by TRAX. Any unrelated helpers should be
removed or moved out of this repository.

## Fail-fast Audit

Most required daemon config panics when missing. `common.GetTraxClusterId` appears to contain
fallback behavior and should be audited against the fail-fast rule if used in active TRAX paths.
