# Specification

## Product scope

`ctxl` is a Go library and developer code generator for schema-defined context stores. A user-owned JSON schema defines storage entities and generation settings. The generator produces a dedicated Go CLI that embeds one schema, exposes its context operations, and carries version-matched Agent Skill content.

`ctxl` itself is not a generic runtime for arbitrary schemas. Runtime schema switching, release automation, and data migration between schema versions are out of scope.

## Terminology

- **ctxl schema**: the user-owned JSON document that is the single source of truth for entities, generation layout, effective names, and bundled Skill inputs.
- **Developer generator**: the generic `ctxl` development command invoked through `go generate` to validate a schema and replace generated output.
- **Generated CLI**: a dedicated executable containing exactly one ctxl schema and the commands derived from it.
- **Generated-owned root**: a marked output boundary that regeneration may replace completely.
- **Built-in Skill**: the single ctxl-owned Skill that documents the generated CLI's schema-derived commands.
- **Custom Skill**: a user-owned Agent Skills-compatible directory bundled without ctxl rewriting its contents.
- **Effective Skill bundle**: the built-in Skill plus every configured custom Skill carried by the generated CLI.

## Observable contracts

### Schema and generation

- The ctxl schema is the generation SSOT. `go:generate` identifies it but does not duplicate generation configuration.
- Effective CLI, store, built-in Skill, module, output, and entity command names derive from schema names and may be overridden in the schema JSON.
- Declared relative paths resolve from the schema file's directory.
- `standalone` is the default output mode and creates an independently buildable Go module.
- `existing-module` creates a generated command inside an existing Go module without modifying its `go.mod` or `go.sum`.
- Regeneration fully replaces marked generated-owned output and removes stale files. It does not preserve edits within that boundary.
- Generation validates all inputs and renders a complete replacement before changing the current output.
- Invalid input, missing dependencies, unsafe output targets, or render failures leave the previous complete output unchanged.

### Developer and runtime boundaries

- The generic developer `ctxl` command provides generation-time behavior only and has no schema entity commands.
- Generated CLIs embed exactly one normalized typed schema, do not parse schema JSON at startup, and never expose `--schema`.
- The Go API continues to allow downstream code to construct a schema-specialized command without using the standard generated project.
- Schema parsing normalizes once: it applies the canonical schema's literal defaults and resolves defaults that depend on other values, such as effective names and output paths. Runtime and Go constructors consume that normalized model and do not reinterpret incomplete structs or accept out-of-schema identity overrides.
- Schema evolution is allowed to break commands and stored data compatibility. ctxl provides no automatic data migration contract.

### Generated CLI commands

- Generated CLIs retain `--scope project|global` and `init`.
- Singular markdown and symlink entities expose `write` and `show`.
- Plural NDJSON entities expose `append`, `list`, and `get`.
- Plural markdown entities expose `create`, `update`, `list`, `get`, and `delete`.
- Generated CLIs always expose `skills get` and `skills path`. They expose `skills list` only when the effective Skill bundle contains two or more Skills. Invoking an omitted `skills` subcommand fails as an unknown command.
- When the effective Skill bundle contains exactly one Skill, `skills get` and `skills path` accept zero arguments and operate on that Skill. Passing the correct name still succeeds; a wrong name fails as an unknown Skill.
- Generated CLIs do not expose `schema validate` or a schema-authoring Skill; schema validation belongs to generation.

### Store behavior

- Project store files default to `.<effective-store-name>/` under the working directory; global files default to `~/.<effective-store-name>/`.
- An entity with `location: root` uses the project root in project scope. Global scope always uses the global store directory.
- An entity may restrict itself to project scope, global scope, or allow both.
- Singular markdown `replace` overwrites one file; `show` returns flattened declared fields and a non-empty body.
- Singular markdown `section` replaces only the configured heading body, ignores headings inside code fences, preserves the matched heading line, and creates a missing file.
- Singular symlink creation is idempotent for the same target, refuses a missing target, and refuses an existing different filesystem entry.
- Plural NDJSON appends one JSON object per line, supplies a missing sequential `id` and RFC3339 `ts`, and hides object fields from default list output unless full output is requested.
- Plural markdown stores one `<id>.md` file per item and exposes create, update, list, get, and delete behavior.
- `init` processes entities in schema order and leaves existing paths unchanged unless its explicit force behavior is selected.

### Validation

- A checked-in Draft 2020-12 JSON Schema is the canonical machine-readable contract for ctxl schema documents.
- Raw schema input is validated with a standards-compliant JSON Schema library before decoding and derivation.
- Validated input decodes directly into the runtime model types; they carry no parallel validator tags.
- Unknown schema properties are rejected unless the canonical schema defines an extension point.
- Configuration that cannot affect its selected mode or entity shape is rejected instead of being silently ignored.
- Validation failures identify the source and instance path and occur before generated output is changed.

### Agent Skills

- `skills` is a discriminated array of `builtin` and `custom` entries.
- Every generated CLI has exactly one built-in Skill. If no built-in entry is declared, the default is implicit; declaring more than one is invalid.
- A built-in entry may configure generated frontmatter, an optional Agent Skill source directory, and `before` or `after` generated-instruction injection. `before` is after YAML frontmatter; `after` is after the existing Markdown body.
- Generated built-in instructions use the effective generated CLI and command names and never instruct an agent to pass a schema file.
- A custom entry has exactly `type` and `directory`. Any number of custom entries is allowed, and ctxl does not rewrite, inject, rename, or merge their contents.
- Built-in and custom Skill directories follow the Agent Skills directory specification: `SKILL.md` is required and supporting files retain their relative paths.
- The complete effective Skill bundle is embedded in the generated CLI. `skills get` returns a selected Skill's instructions, and `skills path` returns its complete content-addressed materialization for relative file access. `skills list` enumerates the bundle only when it contains two or more Skills.
- Generated built-in instructions document the zero-argument `skills get` and `skills path` forms when the bundle is a singleton, and `skills list` plus the named forms otherwise.
- Skill materialization is atomic and safe under concurrent callers.
- Unsafe paths, escaping symlinks, invalid frontmatter, name mismatches, duplicate effective names, multiple built-in entries, and extra custom-entry fields are rejected during generation.

## System-wide constraints

- Core store behavior and schema-specialized CLI construction remain reusable Go packages; code generation is an adapter over them.
- Generated code is intentionally replaceable and must carry an unambiguous generated marker.
- Generated standalone modules pin the ctxl runtime that matches their generator. Existing-module generation verifies rather than mutates the parent dependency.
- The first packaging contract is standard `go build`; CI and release scaffolding are separate features.
- Repository agent entrypoint is root `AGENTS.md`; `CLAUDE.md` is a symlink to it.
- New behavior follows the PRD/ADR/spec workflow before implementation.

## Current implementation status

- The generic `ctxl` executable is a developer-only generator.
- Generated standalone and existing-module commands embed their validated schema and complete Skill bundle.
- The runtime command builder and store remain reusable through the public Go API.
