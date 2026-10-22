
# CI Runner Platform — Job Lifecycle

## Purpose

This document describes the job states, runner states, and scheduling behavior implemented in the initial in-memory CI Runner Platform prototype.

## Job States

The job model defines these states:

- `QUEUED` — The job is waiting to be assigned.
- `ASSIGNED` — The scheduler has selected a runner and recorded the assignment.
- `RUNNING` — The job has started executing.
- `SUCCEEDED` — The job completed successfully.
- `FAILED` — The job failed.
- `CANCELLED` — The job was cancelled.
- `TIMED_OUT` — The job exceeded its allowed execution time.

The current implementation actively performs only the transition from `QUEUED` to `ASSIGNED`. Execution and terminal-state transitions are not implemented yet.

## Runner States

The runner model defines:

- `provisioning` — The runner is being created.
- `ready` — The runner can accept work.
- `busy` — The runner has reached its configured concurrent-job limit.
- `draining` — The runner should not accept new work.
- `offline` — The runner is unavailable.

The scheduler considers only `ready` and `busy` runners, and a busy runner is eligible if it still has a free concurrency slot.

## Scheduling Eligibility

A job can be scheduled only when its state is `QUEUED`.

A runner must:

1. Be in `ready` or `busy` state.
2. Have an available concurrency slot.
3. Match the requested operating system and architecture when specified.
4. Support Docker or GPU when required.
5. Meet the job's minimum CPU, memory, and disk requirements.

The scheduler selects the first eligible runner in the input slice. It does not rank runners or optimize for cost, utilization, or locality.

## Assignment Behavior

The `Assign` operation:

1. Selects an eligible runner.
2. Changes the job state to `ASSIGNED`.
3. Records the selected runner ID.
4. Sets the assignment timestamp.
5. Adds the job ID to the runner's current-job list.
6. Marks the runner `busy` when it reaches its configured concurrency limit.

If no runner is eligible, assignment returns failure without changing the job or runner values.

The caller must retain the returned job and runner values because this prototype uses in-memory values rather than persistent storage.

## Current Limitations

- No persistent job or runner state.
- No concurrent reservation or synchronization guarantee.
- No distributed coordination across scheduler processes.
- CPU, memory, and disk checks compare job requirements against the runner's configured capacity; they do not measure currently available resources.
- No heartbeat-expiry or stale-runner detection.
- No runner provisioning, job execution, retry, cancellation, or timeout handling.
- No fairness, priority, queue-age, cost, or locality-based scheduling.
- No API or durable queue integration.

These are known prototype limitations, not implemented capabilities.

## Verification

The scheduler tests cover basic matching, concurrency eligibility, resource and capability rejection, successful assignment, no-eligible-runner behavior, and rejection of a job that is not queued.

Run the test suite from the `ci-runner-platform` directory:

```bash
go test -count=1 -v ./...
go vet ./...
```
