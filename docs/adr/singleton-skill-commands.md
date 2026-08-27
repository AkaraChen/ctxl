# Collapse singleton Skill commands at runtime

Status: Accepted

## Context

Generated CLIs assemble their command tree from a normalized schema and an already-built Skill bundle in the shared `cli` package. The built-in Skill's instructions are generated from that same schema in `skillsgen`. The product contract collapses `skills` when the effective bundle contains exactly one Skill.

Two implementation sites must agree: the command tree an agent can invoke, and the instructions that tell the agent which commands exist.

## Decision

- Detect a singleton from the effective `skillbundle.Bundle` length when constructing the `skills` command tree.
- Omit the `list` subcommand when the bundle has exactly one Skill; keep it otherwise.
- Accept zero or one positional argument for `get` and `path` on a singleton. Zero arguments resolve to the sole Skill name. One argument is validated against the bundle as today.
- Keep `ExactArgs(1)` for `get` and `path` when the bundle is empty or has two or more Skills.
- Generate built-in instruction text from the normalized schema's Skill count (the same set that becomes the runtime bundle). Singleton instructions document the zero-argument forms and omit `list`. Multi-skill instructions keep the existing named forms.

## Considered alternatives

### Generate a different command tree in the standalone/existing-module templates

Rejected. Command assembly already lives in the shared runtime so the public Go API and generated binaries stay identical. Duplicating the collapse in templates would split the contract.

### Keep `list` always and only add optional names

Rejected by the product contract. A one-element list adds no value, and omitting it is the signal that no discovery step is required.

### Treat an omitted name as "first Skill" in every bundle

Rejected. Multi-skill CLIs must address Skills by name. Implicit first-Skill selection would hide collisions and change multi-skill behavior.

## Trade-offs and consequences

- The command tree is a function of the bundle actually embedded in the process, not of schema authoring choices such as implicit vs explicit built-in.
- Generated instruction text must stay aligned with that tree. Counting `schema.Skills` after normalization is sufficient because generation embeds exactly that set.
- Removing `skills list` from singleton CLIs is a breaking help-surface change for those binaries; named `get` / `path` remain compatible.
- Empty bundles constructed through the Go API keep the multi-skill command shape (`list` present, name required) because they are not a generated-CLI case.

## Validation

- Unit-test the command tree and zero-argument `get` / `path` through `cli.New`.
- Unit-test generated instruction wording for singleton and multi-skill schemas.
- Keep the existing multi-skill generation end-to-end coverage of `list` and named `get` / `path`.
