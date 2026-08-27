# Generated schema CLI

Status: Accepted

## Problem and context

The current `ctxl` executable accepts `--schema FILE` at runtime and builds entity commands from that file. This makes every invocation carry deployment configuration, lets one binary change identity at runtime, and forces the generic executable, downstream Go API, generated commands, and generated Skills to support the same dynamic-schema path.

A schema author instead needs a repeatable Go code-generation workflow that produces a dedicated CLI. The resulting binary must carry one schema, the commands derived from it, and version-matched Agent Skill content. Runtime users must not locate or pass the schema.

## Users and user stories

- A Go developer owns a ctxl schema in their repository and runs standard `go generate` to produce a dedicated CLI.
- The developer can generate a standalone Go module for independent packaging or integrate a generated command into an existing Go module.
- A runtime user invokes schema-derived commands without `--schema`.
- An agent discovers bundled Skill content that describes the exact installed CLI and can access the Skill's referenced files.
- An advanced downstream developer can continue constructing a CLI through the public Go API without using the generator template.

## Goals

- Make the schema JSON the single source of truth for runtime behavior and generation configuration.
- Make the generic `ctxl` executable a developer-only generator, not a generic context-store runtime.
- Generate a directly buildable, schema-specialized CLI.
- Preserve the current schema-derived entity operations in generated CLIs.
- Bundle one generated built-in Skill plus any number of user-authored custom Skills with the binary.
- Support full, deterministic regeneration rather than preserving edits to generated files.

## Non-goals

- Selecting or overriding a schema at runtime.
- Preserving the existing generic `ctxl --schema` workflow or providing a deprecation period.
- Migrating, deleting, or preserving compatibility with data created under an older schema.
- Generating GoReleaser configuration, CI workflows, installers, or other release automation.
- Providing extension hooks inside generated-owned files. Custom implementations use the public Go API or user-owned files outside generated roots.

## Product contract

### Schema ownership and defaults

- The user-owned JSON schema is the generation SSOT. Generation mode, output, module identity, CLI identity, store identity, and the complete `skills` array are declared there.
- `go:generate` identifies the schema file; it does not duplicate those settings as command-line flags.
- CLI name, store name, built-in Skill name, standalone module identity, and default output names derive from top-level `name` unless explicitly overridden in the JSON.
- Entity command names derive from the entity name unless explicitly overridden in the JSON.
- All paths declared by the schema are resolved relative to the schema file, not the caller's working directory.

### Skill array

- `skills` is a discriminated array whose entries have `type: "builtin"` or `type: "custom"`.
- Every generated CLI contains exactly one built-in Skill. If `skills` is omitted or contains no `builtin` entry, ctxl supplies the default built-in entry. At most one explicit `builtin` entry is allowed.
- A built-in entry configures the ctxl-generated command Skill. It may override its Agent Skill frontmatter, optionally provide a source directory, and choose `before` or `after` generated-instruction injection.
- A custom entry contains exactly `type` and `directory`. Its directory is a complete user-owned Agent Skill and its `SKILL.md` is never rewritten or injected by ctxl.
- Any number of custom entries is allowed. Effective Skill names across the built-in and custom entries must be unique.

The checked-in JSON Schema is the only normative field-level shape; the PRD does not duplicate it.

### Developer flow

1. The developer writes a ctxl JSON schema and any complete Agent Skill directories referenced by its `skills` array.
2. A standard `go:generate` directive invokes the developer `ctxl` generator with that schema file.
3. The generator validates every input before changing generated output.
4. The generator replaces the generated-owned output for the selected mode.
5. The developer builds the result with standard `go build`.

### Output modes

- `standalone` is the default. It produces a complete generated project with its own `go.mod`, pins the ctxl runtime version corresponding to the generator, and builds with `go build .` from that project.
- `existing-module` generates a command package under the caller's existing module and builds through that module, such as `go build ./cmd/<name>`.
- Existing-module generation does not edit the parent `go.mod` or `go.sum`. If the required ctxl dependency is absent or incompatible, generation fails with an exact remediation command.
- Both modes expose the same runtime CLI contract.

### Generated ownership

- Regeneration fully replaces generated-owned roots and removes stale generated files. The generator does not merge or preserve user edits there.
- Standalone mode owns its complete generated project directory.
- Existing-module mode owns only the declared generated command root; it never treats the surrounding Go module as generated-owned.
- A generated marker identifies an owned root. A missing or non-generated target must be empty, absent, or explicitly selected for first creation; ctxl must not recursively replace an arbitrary unmarked directory.
- Generation is atomic: invalid input or a failed render leaves the previous complete generated output in place.

### Runtime command surface

- The developer `ctxl` tool exposes generation-time behavior only. It does not expose schema entity commands.
- A generated CLI embeds exactly one active schema and does not accept `--schema`.
- A generated CLI retains `--scope`, `init`, and the commands derived from its entities:
  - singular markdown and symlink entities: `write`, `show`;
  - plural NDJSON entities: `append`, `list`, `get`;
  - plural markdown entities: `create`, `update`, `list`, `get`, `delete`.
- A generated CLI always exposes `skills get` and `skills path`. It exposes `skills list` only when the effective Skill bundle contains two or more Skills.
- When the effective bundle contains exactly one Skill, `skills get` and `skills path` accept zero arguments and operate on that Skill. Passing the correct name still succeeds.
- A generated CLI does not expose the runtime-redundant `schema validate` command or the schema-authoring Skill.

### Bundled Skills

- The built-in Skill contains ctxl-generated instructions for the effective CLI name, entity command names, scopes, fields, and Skill-loading commands.
- Without a built-in `directory`, ctxl generates the complete built-in Skill. With a directory, ctxl uses that complete Agent Skill as the base and applies the configured frontmatter overrides and instruction injection.
- The built-in schema selects `before` or `after` injection:
  - `before` inserts generated instructions immediately after the YAML frontmatter and before the source body;
  - `after` appends generated instructions after the source body.
- Injection never places content before YAML frontmatter. The resulting built-in `SKILL.md` and output directory name must satisfy the Agent Skills naming and frontmatter rules.
- Each custom Skill directory is packaged byte-for-byte except for the representation required to embed and materialize it. ctxl does not inject instructions, rewrite frontmatter, rename it, or merge it with another Skill.
- Every bundled directory follows the Agent Skills specification: `SKILL.md` is required and supporting directories or files are allowed. Relative paths are preserved; unsafe links or paths escaping the Skill root are rejected.
- Generated CLIs serve every bundled Skill through `skills get` and, when more than one Skill is present, `skills list`. `skills path` returns a real directory for the selected Skill so relative references and scripts remain usable.
- Generated built-in instructions document the zero-argument `skills get` / `skills path` forms when the bundle is a singleton, and the named forms plus `skills list` otherwise.
- Materialization from a single binary is content-addressed and atomic. Identical bundled content reuses the same extracted directory.

## User-visible failures

- Invalid JSON, a schema that violates the ctxl JSON Schema, inapplicable configuration that would otherwise be ignored, or an invalid Skill directory fails before output replacement.
- Validation errors identify the JSON instance path or Skill-relative path and the violated rule.
- Existing-module dependency mismatches fail with the required ctxl module version and a precise command to resolve it.
- Output collisions outside a marked generated root fail without deleting the target.
- A failure while materializing bundled Skill files does not return a partial path.

## Acceptance criteria

1. A schema containing only required values generates a standalone module whose effective CLI, store, built-in Skill, module, and output names derive from `name`.
2. Each derived identity and command or built-in Skill name can be overridden in the same JSON schema and appears consistently in help, storage paths, generated instructions, and built-in Skill metadata.
3. The default standalone output passes `go build .`; existing-module output passes the parent module's `go build` after its declared dependency precondition is met.
4. The developer `ctxl` command has no entity runtime commands. A generated CLI has the expected entity commands and no `--schema` flag.
5. The generated CLI's entity operations retain the existing observable store behavior for singular markdown, sections, symlinks, NDJSON collections, and markdown collections.
6. With no `skills` entry, a generated CLI contains exactly one default built-in Skill. A custom-only array still receives that implicit built-in Skill, while an explicit built-in entry replaces the default configuration; a second explicit built-in is rejected.
7. Multiple custom fixtures containing `SKILL.md`, scripts, references, assets, and other allowed files are bundled without path loss or content rewriting. `skills list`, `skills get`, and `skills path` expose every built-in and custom Skill when two or more Skills are present.
8. Built-in `before` and `after` injection preserve YAML frontmatter at the start of `SKILL.md`, place generated instructions at the configured boundary, and pass Agent Skills conformance checks. Custom `SKILL.md` files remain unchanged.
9. Regeneration after removing or renaming an entity removes stale generated commands and instructions. User edits inside generated-owned output are not preserved.
10. Invalid schema or Skill input and render failures do not partially replace a previously valid output.
11. The public Go construction API remains usable by a downstream test that embeds a schema without invoking code generation.

## Resolved product decisions

- Both standalone and existing-module output are first-class; standalone is the default.
- Schema changes may be breaking. No runtime data migration contract is introduced.
- The first release stops at standard Go buildability and version-matched Skill delivery.
