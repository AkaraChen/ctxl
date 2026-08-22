# ADR: Schema-driven filesystem surfaces

Status: accepted (describes the shipped architecture).

## Context

ctxl sells a generated context scheme. Step 1 must turn one JSON schema into disk files, a cobra CLI, and runtime skills without a second source of truth. The only backend that exists is the local filesystem.

## Decision

Keep a single Go module with a one-way dependency: `cli` → `core`. `core` does not import cobra.

- A schema names its backend with top-level `backend.type`. The only valid value today is `filesystem`. Additional fields on `backend` are reserved for later backends and filesystem options.
- `core/schema` owns types, parse, load, and validation.
- `core/store` owns filesystem paths and the five shape read/write flows.
- `core/skillsgen` owns skill markdown generated from a schema.
- `cli` owns flags, subcommands, and stdout.
- `cmd/ctxl` is the generic binary. It does not embed a schema; callers pass `--schema FILE`.
- The same `cli.New` API accepts an already-parsed schema so a branded binary can embed JSON.

Filesystem rules live in `core/store`:

- Store name is `schema.name`.
- `--scope project` uses `.<name>/` under the working directory, plus `location: root` paths at the project root.
- `--scope global` uses `~/.<name>/` for every entity path.
- Entity `scope` is a filesystem-backend lock. Later backends may invent their own scope (including RBAC). They must not be assumed to reuse these words.

## Alternatives

- Generate static CLI source per schema: drifts from the installed binary; rejected in favor of runtime command trees and `skills get`.
- Hide files behind the CLI: contradicts the product stance that disk is a public, greppable contract.
- Treat scope/location as global product vocabulary: too early; they are filesystem features.

## Consequences

- Adding a backend later is a new store implementation, not a change to schema-as-source-of-truth.
- Skills must be served from the installed binary so they cannot drift from the verbs.
- Known gap: `store.Entity` enforces entity scope, but CLI verbs pass schema entities straight into read/write helpers and skip that check.

## Validation

Existing `core/schema`, `core/store`, and `cli` tests plus `examples/demo.schema.json` are the archive suite. Do not "fix" README-vs-code deviations in this ADR's name; they are listed in `docs/prd/context-layer.md`.
