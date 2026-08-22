# Specification

## Product scope

ctxl is a generated context layer. A JSON schema names a store and declares entities. The same module materializes a filesystem store, a cobra CLI, and runtime agent skills.

The generic `ctxl` CLI is the primary surface. A branded binary may embed a schema through the same CLI constructor.

Out of scope until an accepted PRD says otherwise: extra backends, multi-person sync, multi-machine consensus, collaboration RBAC, and workspace-wide search.

## Terminology

- **Feature 质问**: the mandatory product-then-technical clarification loop driven by `$feature-dev` before implementation.
- **PRD**: a product requirements document under `docs/prd/`.
- **ADR**: an architecture decision record under `docs/adr/`.
- **Spec**: this file — shared terminology, observable contracts, and invariants.
- **Context**: a searchable string that an agent should be able to retrieve.
- **Schema**: one JSON object with `name`, `backend`, and a non-empty `entities` array.
- **Backend**: the persistence and retrieval implementation selected by `backend.type`. The only accepted type today is `filesystem`. Other fields on `backend` are reserved for later.
- **Store**: the named context collection implied by a schema. On the filesystem backend it is a directory tree, not a separate database.
- **Entity**: one regular context type declared in the schema. Step 1 has five shapes: singular replace, singular section, singular symlink, plural NDJSON, plural markdown.
- **Filesystem backend**: the only shipped backend. `--scope` and `location` belong to it.
- **Project scope**: files under `.<name>/` in the working directory, or at the project root when `location` is `root`. Expected to live in the project tree.
- **Global scope**: files under `~/.<name>/`. Machine-local memory, not a repository.
- **Product flag**: a flag on `ctxl` (schema, scope, entity filter). Not a flag on `grep` or `tree`.
- **Identity**: entity name, item id, or entity path.
- **Content**: the persisted text of an entity (singular text, NDJSON line, markdown frontmatter and body).

## Observable contracts

### Documentation harness

- New product behavior is defined in `docs/prd/` before feature code lands.
- Material technical choices are recorded in `docs/adr/` before or with the code that depends on them.
- Stable, implementation-independent rules merge into this file.
- Agents must not implement feature code during 质问.

### Schema

- `schema.name` is required and is the store name.
- `schema.backend` is required. `backend.type` must be `filesystem` until another backend is specified.
- `schema.entities` is a non-empty array of uniquely named entities.
- Each entity has `name`, `kind` (`singular` | `plural`), `format` (`markdown` | `ndjson` | `symlink`), and `path`.
- Singular supports markdown or symlink. Plural supports markdown or ndjson. Symlink is singular only.
- `write` is `replace` (default) or `section`. Section requires singular markdown and a heading.
- `location` is `root` or `store` (default `store`).
- Entity `scope` is `project`, `global`, or `both` (default `both`).
- `schema validate` prints `ok` and exits 0 when the active schema is valid.

### Filesystem store

- Disk files are a public contract. Humans may read, grep, and hand-edit them. The CLI reads and writes the same files.
- `--scope project|global` selects which filesystem tree to use. Default is `project`. Any other value is an error.
- Project `location: store` paths live under `.<name>/`. Project `location: root` paths live at the working-directory root.
- Global scope places entity paths under `~/.<name>/` regardless of `location`.
- `init` creates every declared path in schema order. Existing paths are left alone unless `--force`.

### CLI verbs (shipped)

- Singular shapes expose `write` and `show`.
- Plural NDJSON exposes `append`, `list`, `list --full`, and `get --id`. `id` and `ts` are filled on append when absent. Default `list` omits object fields.
- Plural markdown exposes `create`, `update`, `list`, `get`, and `delete` by `--id`.
- Symlink `write` is a no-op if already linked, errors if the target is missing, and refuses an existing non-link.
- Singular section write ignores headings inside fences. A missing file is created first.
- Read verbs that return data print JSON on stdout.
- `skills get overview`, `skills get schema`, and `skills get <entity>` print markdown generated from the active schema. Skills are not long checked-in guides.
- The generic binary does not embed a schema. `--schema FILE` selects one. A branded binary may supply the schema in process.

### Store-wide read verbs

See `docs/prd/cli-grep-tree.md`. These verbs belong to the filesystem backend.

- `grep` and `tree` operate only on the ctxl store, never on the whole workspace, and never by execing host `grep` or `tree`.
- Both verbs share one entity-selection rule: current `--scope`, then optional `--entity` flags on `ctxl`. Unknown `--entity` is an error.
- `grep` default match is a literal substring. `-E` on the grep tail selects regular expressions.
- `grep` searches identity and content. Each content matching line is one hit. Each matching identity field is one hit.
- A grep hit is a JSON object with entity, optional item id, match kind (`identity` or `content`), and snippet.
- `grep` with no hits prints an empty list and exits non-zero.
- `tree` is a two-level logical list: store → entities. Absent files still appear. Plural items are not children.
- `tree` prints one JSON object per selected entity.
- Generated skills tell agents to use these verbs instead of host `grep` / `tree`.

## System-wide constraints

- Repository agent entrypoint is root `AGENTS.md` (`CLAUDE.md` is a symlink to it).
- Feature development workflow skill lives at `.agents/skills/feature-dev/` (also linked from `.claude/skills/`).
- `cli` may depend on `core`. `core` does not import `cli` or cobra.
- Commit attempts re-check the working tree against this specification and relevant PRDs/ADRs.

## Current implementation status

- Step 1 (schema → filesystem store + CLI + runtime skills) is shipped. Archive: `docs/prd/context-layer.md`.
- Filesystem `grep` and `tree` are specified in `docs/prd/cli-grep-tree.md` and implemented.
- Known shipped deviations are listed in `docs/prd/context-layer.md` and must not be silently "corrected" by documentation.
