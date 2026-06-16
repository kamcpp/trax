# Core Requirements

This page captures the durable TRAX requirements that should guide implementation, tests, and
operations. It is intentionally business-neutral: dependent systems provide their own workflow
templates and executors, while TRAX owns the generic orchestration mechanics.

## PostgreSQL Is Authority; RabbitMQ Is Transport

RabbitMQ delivers work and results, but PostgreSQL is the durable state authority. Coordinators
must re-read persisted saga and step state before mutating workflows, especially after duplicate
delivery, reconnects, or delayed messages.

Related pages:

- [Architecture v1](../architecture/v1.md)
- [RabbitMQ Routing](../concepts/rabbitmq-routing.md)
- [Coordinator Algorithms](../architecture/coordinator-algorithms.md)

## Saga Processing Requires Per-Instance Guards

A single saga instance must not be processed concurrently by multiple coordinator paths. Coordinator
work needs per-saga locking or equivalent guards before state transitions, result handling, and
compensation decisions.

Related pages:

- [Coordinator](../concepts/coordinator.md)
- [Coordinator Algorithms](../architecture/coordinator-algorithms.md)
- [State Machine](../architecture/state-machine.md)

## Notifications Wake; They Do Not Decide

`trax_saga_events` and `trax_template_events` are wakeup signals. They reduce latency but do not
carry authoritative workflow decisions. The store remains the source for runnable steps, template
state, and terminal conditions.

Related pages:

- [Notifications](../concepts/notifications.md)
- [PostgreSQL Store](../concepts/postgresql-store.md)
- [Template Hot Reload](../concepts/template-hot-reload.md)

## Template CRUD And Hot Reload Are Core

Saga-template and saga-step-template CRUD are runtime capabilities, not offline setup chores.
Coordinators must react to template changes through notifications and periodic reload checks, then
initialize or unmark runtime bindings as templates change.

Related pages:

- [Saga Template](../concepts/saga-template.md)
- [Saga Step Template](../concepts/saga-step-template.md)
- [Template Management and Hot Reload](../flows/template-management.md)

## Submitter Readiness Means Usable Routing

A submitter is ready only after it has announced successfully and received the cluster IDs plus
inbox/outbox routing data it needs. A process being alive is not enough.

Related pages:

- [Submitter](../concepts/submitter.md)
- [TRAX MQ and Coordination](../concepts/trax-mq-and-coordination.md)
- [Saga Lifecycle](../flows/saga-lifecycle.md)

## Idempotency Is Mandatory For Side Effects

Every logical operation that can cross a retry boundary needs a deterministic identity. Saga and
step idempotency keys must survive retries and flow into executor work that performs side effects.

Related pages:

- [Idempotency](../concepts/idempotency.md)
- [Idempotent Service](../concepts/idempotent-service.md)
- [Executor And CLI Runtime](../architecture/executor-and-cli.md)
- [PostgreSQL Data Model](../data-model/postgresql.md)

## Sub-Saga Hierarchy Is A First-Class Mechanism

TRAX must support workflows that spawn child workflows. Parent saga ID, parent step ID, root saga
ID, and depth are core fields for inspection, compensation analysis, and operator tooling.

Related pages:

- [Sub-saga](../concepts/sub-saga.md)
- [Sub-sagas and Hierarchy](../flows/sub-sagas.md)
- [Saga Instance](../concepts/saga-instance.md)
- [Executor](../concepts/executor.md)

## E2E Must Verify The Distributed Shape

TRAX E2E coverage should exercise the real distributed runtime: PostgreSQL, RabbitMQ, Redis,
`traxctrl`, multiple `traxcoord` instances, submitters, executors, result capture, and red-path
behavior. Unit tests are not enough for routing, locking, and retry semantics.

Related pages:

- [Testing and E2E](../architecture/testing-and-e2e.md)
- [Testing and E2E Operations](../operations/testing.md)
- [Make Targets](../operations/make-targets.md)

## Business Neutrality

TRAX owns workflow execution mechanics. Dependent systems own business templates, payload schemas,
and executor implementations. TRAX requirements should be expressed in terms of
generic orchestration behavior: state, routing, retries, idempotency, compensation, hierarchy,
inspection, and operations.
