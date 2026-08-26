# Example: OpenSpec as a ctxl schema

Models the storage core of [Fission-AI/OpenSpec](https://github.com/Fission-AI/OpenSpec), an AI-native CLI for spec-driven development. OpenSpec manages an `openspec/` directory holding change proposals, capability specs, and a workflow config, and ships one Agent Skill per workflow.

Mapping:

- `openspec new change <name>` → `openspec change create --id <name> --schema spec-driven ...`; `openspec list` / `show` → `change list` / `change get --id <name>`. Change identity is the kebab-case name; proposal structure (`## Why`, `## What Changes`, ...) lives in the markdown body.
- `openspec/specs/<capability>/spec.md` → the `spec` plural markdown collection (`spec create --id <capability> --body ...`).
- `openspec/config.yaml` (`schema:`, `store:`, `context:`) → the singular `config` entity: frontmatter fields plus a context body.
- The `event` NDJSON log is an addition this example uses to demonstrate append-only collections; OpenSpec itself tracks lifecycle by moving directories.
- Two of OpenSpec's twelve generated skills (`openspec-propose`, `openspec-apply-change`) are bundled verbatim as custom Skills; the built-in Skill documents the store commands. In real OpenSpec these files are generated per tool by `openspec init`; here `skills get` / `skills path` serve them from the binary.

Simplified relative to real OpenSpec: changes are one markdown file, not a directory of artifacts (`proposal.md`, `design.md`, `tasks.md`, delta specs); capability ids cannot be nested paths (`area/cap`); `archive` (moving a change into `changes/archive/<date>-<id>/` and merging deltas into main specs) is an operation ctxl does not model; the store directory is `.openspec/` rather than `openspec/`; config is markdown frontmatter rather than YAML.

Generate with a released ctxl (or in-repo via the `TestExamples` fixture):

```go
//go:generate go run github.com/AkaraChen/ctxl/cmd/ctxl@latest generate context.schema.json
```
