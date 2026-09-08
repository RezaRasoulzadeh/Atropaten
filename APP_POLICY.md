# Application policy

## Explicit destructive actions

Users may remove records and their dependent operational records after an explicit confirmation. The UI must not disable a destructive action solely because dependencies exist.

When a dependency affects an immutable inventory or financial ledger, the application removes the operational records and records a compensating movement or reversal first. Immutable ledger entries are never rewritten or physically deleted, so totals and audit history remain numerically consistent.

Archive is available when a non-destructive catalog action is preferred. Every remove, delete, archive, or restore action must use the shared destructive-action styling and a clear confirmation where data can be lost.
