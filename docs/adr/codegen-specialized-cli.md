# Generate schema-specialized CLIs

Status: Accepted

## Context

The current root command mixes two roles: a generic runtime accepts `--schema`, while the Go package can also receive a schema directly through `cli.Options`. Runtime schema selection forces command discovery to inspect process arguments before Cobra constructs the tree and makes generated Skill text repeat `--schema <file>`.

The desired product has a developer generator and schema-specialized runtime binaries. It must support a portable standalone project, integration into an existing Go module, and advanced downstream construction through the Go API.

## Decision

- Recast the generic `ctxl` executable as a developer-only code generator. Remove its entity runtime and `--schema` behavior.
- Invoke generation through standard `go generate`; the invocation identifies one schema file, while the schema owns all generation settings.
- Generate a thin entrypoint whose source embeds the normalized typed schema and bundled Skill data as Go literals, then call the shared ctxl CLI/runtime packages. The generated executable does not parse or default the user JSON again at startup.
- Keep schema-derived command assembly in the shared runtime rather than expanding separate Cobra implementation source for every entity.
- Preserve the public Go construction API for downstream projects that do not use the generated template.
- Keep normalization at `schema/loader.Parse` and `schema/loader.LoadFile`. Runtime constructors consume the normalized schema directly; they do not retain alternate identity options or compatibility defaults for manually assembled partial structs.
- Support two generation layouts:
  - default standalone project with an owned `go.mod` and pinned ctxl runtime dependency;
  - existing-module command root that relies on a compatible parent dependency without mutating the parent's module files.
- Replace generated-owned output as one atomic unit. Render and verify in a temporary sibling, then exchange it with the marked target.
- In standalone mode, run `go mod tidy` inside the temporary project before replacement so the emitted `go.mod` and `go.sum` are directly buildable. Dependency-resolution or network failure leaves the prior output untouched.

## Alternatives considered

### Keep the generic runtime with an optional embedded default

Rejected. It preserves two runtime identities, keeps `--schema` and its dynamic command-tree path alive, and leaves every Skill responsible for describing both invocation forms.

### Generate complete Cobra source for every entity

Rejected. It duplicates stable command behavior in generated files, increases template and compatibility surface, and weakens the retained public runtime API. A thin generated entrypoint is sufficient because one embedded schema fixes the command tree before the process starts.

### Support only a standalone module

Rejected. Standalone output is the clean packaging default, but existing Go projects need a normal `cmd/<name>` integration path and may compose other user-owned packages around the generated command.

### Modify parent `go.mod` automatically

Rejected. Module files are user-owned and generation could alter unrelated dependency resolution. Existing-module mode instead validates the precondition and returns an exact remediation command.

## Consequences

- Runtime users get a smaller, unambiguous command surface and never locate a schema file.
- The generated project remains small and receives command fixes through its pinned ctxl dependency.
- Standalone generated modules require a resolvable ctxl version. A development build whose own module version cannot be identified must require an explicit version in the schema rather than guessing.
- Standalone generation resolves Go module dependencies and can therefore require network access when the module cache is cold.
- Existing-module generation can fail before writing output when the parent dependency is missing or incompatible.
- Removing the old generic runtime is intentionally breaking; no compatibility alias or deprecation cycle is maintained.
- The generator must test both layouts and the retained public Go API.

## Failure bounds and validation

- Never remove or replace a directory that is not positively identified as the exact generated target.
- Do not change generated output until schema validation, Skill validation, rendering, formatting, dependency checks, and a build-oriented structural check succeed.
- Tests must cover default derivation, every override, stale-file removal, failed-regeneration preservation, both module layouts, generated help, and absence of `--schema`.
