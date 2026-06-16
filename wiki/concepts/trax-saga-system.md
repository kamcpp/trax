# TRAX Saga System

TRAX is a distributed workflow and saga orchestration system for multi-step operations that need persistent state, asynchronous execution, idempotency, and optional compensation.

## Core Model

A saga starts from a `SagaTemplate`. The template contains an ordered list of `SagaStepTemplate` IDs. A runtime submission creates one `SagaInstance` and one `SagaStepInstance` per step.

The current step instance state determines what the coordinator can do next:

- schedule a forward execution request;
- wait for a result;
- schedule compensation;
- mark the saga committed, compensated, blocked, or invalid.

## Actors

- `traxcoord`: coordinator daemon that advances state.
- `traxctrl`: control/read API for templates, clusters, instances, trees, annexes, and operator actions.
- `traxcli`: CLI that can manage templates, submit workflows, run executors, and watch progress.
- submitter: client runtime that announces to coordinators and publishes saga submissions.
- executor: worker runtime bound to a step route and an `IdempotentService`.

## State And Transport

PostgreSQL is the source of truth. RabbitMQ is the transport. PostgreSQL notifications are wakeups.

This split is important: if a message is duplicated or a notification is missed, the coordinator should still recover by reading durable state and using idempotency keys.

## Idempotency

TRAX uses deterministic idempotency keys for saga and step rows. The database enforces uniqueness, and store methods named `Save*Idempotently` make repeated create attempts safe.

Executors also receive idempotency keys so step implementations can protect downstream side effects.

## Scope Boundary

TRAX is business-neutral. It owns the reusable workflow engine, not business-specific saga
definitions. Dependent systems should provide their own templates, payload schemas, and executor
implementations.

## See Also

- [Saga Lifecycle](../flows/saga-lifecycle.md)
- [PostgreSQL Data Model](../data-model/postgresql.md)
- [API Surface](../reference/api-surface.md)

## Related Concepts

- [Saga Template](saga-template.md): durable workflow definition.
- [Saga Instance](saga-instance.md): runtime workflow execution.
- [Coordinator](coordinator.md): runtime actor that advances sagas.
- [Submitter](submitter.md): runtime actor that creates sagas.
- [Executor](executor.md): runtime actor that runs steps.
- [Idempotency](idempotency.md): retry and duplicate-protection model.
