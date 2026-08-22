# ADR: Store-wide read verbs

Status: accepted and implemented for the filesystem backend.

## Context

Per-entity verbs cannot orient or search the whole store. Host `grep` / `tree` are the wrong surface: they scan arbitrary files and return text that is hard to optimize. The product wants two Unix-shaped verbs that still share ctxl's selection rules.

`grep` and `tree` are orthogonal: same selected entity set, different projections.

## Decision

Add two root verbs in `cli`, both built on one selection walk in `core` (not on `os/exec` of host tools).

Invocation:

```text
ctxl [--schema FILE] [--scope project|global] [--entity NAME ...] grep [-E] [-i] PATTERN
ctxl [--schema FILE] [--scope project|global] [--entity NAME ...] tree
```

- Product flags stay on `ctxl`. Entity filter is a root persistent flag, not a `grep`/`tree` flag.
- The `grep` tail looks like grep: default fixed-string match; `-E` uses Go `regexp`; `-i` is case-insensitive. No context-line flags in this cut.
- `tree` needs no tool flags in this cut.

Selection (shared):

- Start from schema entities.
- Apply current `--scope` (and entity filesystem scope when the walk consults the store).
- If `--entity` is present, keep only those names; unknown name fails the command before any search.

Projections:

- `grep` walks identity fields and content lines of the selected set; emits a JSON array of hit objects.
- `tree` emits a JSON array of node objects, one per selected entity, whether or not the path exists. Depth is two: store → entity.

Stdout stays in the existing CLI family: indented JSON via the same print helper as `list` / `show`. Not NDJSON-unless-we-change-all-reads; not ASCII trees.

Skills: `skillsgen` must instruct agents to use these verbs. Do not teach host `grep` / `tree`.

## Alternatives

- Exec host `grep`/`tree` and wrap stdout: rejected; the product is first-party retrieval, and host results are the problem.
- Put `--entity` on each verb: rejected; product flags converge in front.
- Workspace-wide search: rejected; corpus is the ctxl store.
- ASCII `tree` plus optional JSON: rejected; one structured shape for both verbs.

## Consequences

- No schema change.
- Selection logic belongs in `core` so branded binaries and tests share it. `cli` only binds flags and prints.
- Later tree depth (plural items as children) extends the tree projection only; it should not fork a second selection API.
- `grep` no-match is an empty JSON array and a non-zero exit, like grep.

## Failure bounds

- Host binaries are never a fallback.
- Invalid `-E` pattern fails closed.
- Missing files do not fail `tree` and do not fail `grep` by themselves.

## Validation

Acceptance criteria in `docs/prd/cli-grep-tree.md`. Tests should cover literal vs `-E`, identity vs content hits, shared `--entity` for both verbs, absent files on `tree`, and a guard that the command path does not spawn `grep`/`tree`.
