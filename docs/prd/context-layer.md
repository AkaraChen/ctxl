# PRD: Context layer (step 1, as shipped)

Status: accepted (archive of current behavior). This document records what already ships. It does not change feature code.

## Problem and context

People who want an agent to understand a project have different context types. Context is a searchable string. Without a regular shape, those strings are hard to grep, hard to teach an agent, and hard to reuse across projects.

ctxl sells a generated scheme: input one JSON schema, get a complete context read/write surface. Step 1 is the filesystem backend's smallest real-world cut: disk layout, cobra CLI, and runtime-generated skills. Later surfaces (more backends, sync) are out of this PRD. Filesystem `grep` / `tree` live in `docs/prd/cli-grep-tree.md`.

## Users

Primary user: anyone who wants an agent to understand project context. Their business decides which context types exist. ctxl's job is to put those types on entity-like rails so they stay regular and greppable.

The generic `ctxl` CLI is the main surface. Embedding the same schema into a branded binary is a shipped distribution path, not the main story.

## Goals

- One schema names a store and declares entities; the same module materializes disk files, a CLI, and agent skills.
- Context stays a searchable string on disk. The CLI is a more structured retrieval layer over those files, not a second data source.
- Humans may grep or hand-edit the files. Agents should prefer the CLI and generated skills because raw grep results are hard to optimize.
- Five entity shapes are enough for step 1.

## Non-goals

- Multi-person sync or merge (git, CRDT, or otherwise).
- Multi-machine ownership or consensus (including Raft).
- Additional backends (MongoDB, S3, or anything that is not `backend.type: filesystem`).
- Collaboration scope or RBAC. Those belong to later backends.
- A product-level "pick a backend" action. Step 1 always writes local files.

## Scope

In scope, as shipped:

- Schema authoring (`name`, `backend`, `entities`) and `schema validate`.
- `init` creating every declared path.
- Per-entity read/write verbs for the five shapes.
- Runtime `skills list` / `skills get` (`overview`, one skill per entity, `schema`).
- Generic CLI via `--schema FILE`, and the Go embed API for a branded binary.
- Filesystem trees selected by `--scope project|global`.

Out of scope: anything in Non-goals.

## Product stance

- Disk files are a public contract. Project-scope files are expected to live in the project tree and can be committed. Global-scope files are a fixed location on this machine (local memory) and are not a repository.
- `--scope` and entity `location` / `scope` are filesystem-backend features, not universal product words that later backends must keep.
- The five shapes are closed for now, not forever. New backends will need new shapes later.

## User flow

1. Author a JSON schema (`name` + `entities`) and run `schema validate`.
2. Run `init` to create declared paths (existing files stay unless `--force`).
3. Read and write through per-entity commands, or hand-edit the same files.
4. An agent loads `skills get overview`, then exactly one entity skill, and calls the CLI rather than opening files itself.

Completion for step 1 is the current shipped cut: schema → disk + CLI read/write + runtime skills. That step is done.

## Entity shapes (closed for step 1)

1. Singular markdown, `write: replace` — one file, current state. `write` overwrites; `show` prints current values.
2. Singular markdown, `write: section` — one heading in a file. Headings inside fences are ignored. Missing file is created, then the section is written.
3. Singular symlink — `path` points at `target`. Already linked is a no-op. Missing target is refused. An existing non-link is refused.
4. Plural NDJSON — one append-only file. `append` writes a line (`id` and `ts` filled when absent); `list` prints fixed fields unless `--full`; `get --id` reads one row.
5. Plural markdown — a folder of `<id>.md` files. `create` / `update` / `list` / `get` / `delete`.

## Filesystem layout (backend-specific)

- Schema `name` is the store name.
- Project files default to `.<name>/` in the working directory. Global files live under `~/.<name>/`.
- `location: root` keeps a file or plural folder at the project root (project scope). Global scope still places the entity path under `~/.<name>/`.
- `--scope` defaults to `project`.
- An entity may declare `scope: project|global|both` (default `both`). The store API can refuse the wrong scope; see known deviations.

## CLI surface (shipped)

Root persistent flags: `--schema FILE`, `--scope project|global`.

| Shape | Verbs |
| --- | --- |
| Singular replace / section / symlink | `write`, `show` |
| Plural NDJSON | `append`, `list`, `list --full`, `get --id` |
| Plural markdown | `create`, `update`, `list`, `get`, `delete` |
| Store | `init`, `init --force` |
| Meta | `skills list`, `skills get <name>`, `schema validate` |

Read commands print JSON on stdout. `schema validate` prints `ok` on success.

Skills are generated from the live schema, not checked in as long guides. `skills/ctxl/SKILL.md` is a stub that points at `skills get`.

## Failure behavior (as shipped)

- Invalid schema JSON or failed validation is an error; entity commands are not a substitute for `schema validate`.
- Symlink write refuses a missing target and refuses an existing non-link.
- Plural markdown `create` refuses an existing id.
- Plural `get` / `update` / `delete` fail when the id is missing.
- `init` without `--force` skips existing paths.
- Unknown `--scope` is an error.

## Known deviations (do not treat as new product decisions)

These are current facts. They are not cleaned up in this archive.

- README once said "Three entity shapes"; five shapes ship.
- Entity `scope` is enforced by `store.Entity`; the CLI entity verbs do not call that check, so a locked entity can still be read or written under the other `--scope`.
- `status show` (singular) prints a flattened JSON object. `note get` (plural markdown) prints the internal `{Fields, Body}` wrapper.
- `last_green` is a hardcoded field name: empty on singular replace write is filled with the current RFC3339 time.
- Generated entity skills say "Do not open the files yourself." The product stance still allows humans to grep and hand-edit those files.

## Acceptance criteria (archive)

These already hold in the current tree:

- A valid schema includes `backend.type: "filesystem"` and can be loaded by `ctxl --schema FILE` and by `schema.Parse` / `schema.LoadFile`.
- A schema without `backend.type`, or with an unsupported type, fails validation.
- `schema validate` exits 0 and prints `ok` for a valid schema; invalid schema is non-zero.
- `init` creates every declared path in schema order and leaves existing files alone unless `--force`.
- Each of the five shapes has the verbs in the table above, and writes land on the paths implied by `name`, `--scope`, `location`, and `path`.
- `skills get overview|schema|<entity>` prints generated markdown that matches the active schema.
- The Go embed API (`cli.New` with an embedded schema) exposes the same command tree without requiring `--schema`.
- Project-scope default files sit under `.<name>/` or at root when `location: root`. Global-scope files sit under `~/.<name>/`.

## Resolved product decisions

- Archive shipped behavior; conflicts are deviations, not silent rewrites.
- The product being sold is a generated context scheme; CLI + skills are the first surfaces.
- Users are people who want agents to understand project context.
- Context is a searchable string given entity-like regularity.
- Disk is a public contract. CLI is structured retrieval over the same files.
- Five shapes are temporarily closed.
- `--scope` / `location` belong to the filesystem backend.
- Generic CLI is the primary surface; branded embed is secondary.
- Project-scope files are for the repo; global-scope files are machine-local memory.
