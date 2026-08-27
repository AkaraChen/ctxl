# Single-skill command surface

Status: Accepted

## Problem and context

A generated CLI that bundles exactly one Skill — the common case of only the built-in command-reference Skill — currently forces callers to discover that Skill's name before they can read it:

```bash
yoi-server skills list
yoi-server skills get yoi-server-store
yoi-server skills path yoi-server-store
```

For an agent on first contact, the name carries no information and `list` is a one-element enumeration. The extra round trip is ceremony.

## Users and user stories

- An agent operating a generated CLI with a singleton Skill bundle asks the CLI for its Skill and reads the command reference without first listing names.
- A human or agent that already knows the Skill name can still pass it to `get` or `path`.
- An agent using a multi-skill CLI continues to list names and address Skills by name.

## Goals

- Collapse the `skills` command surface when the effective Skill bundle contains exactly one Skill.
- Keep the named `get` / `path` forms working for the correct name.
- Teach agents the short form from the generated built-in instructions when the bundle is a singleton.

## Non-goals

- Changing how Skills are bundled, named, embedded, or materialized.
- Auto-selecting a Skill when two or more Skills are bundled.
- Removing `skills list` from multi-skill CLIs.
- Changing the developer `ctxl` generator command surface.

## Scope and user flow

When the effective Skill bundle contains exactly one Skill:

1. The caller runs `skills get` with no arguments and receives that Skill's instructions.
2. The caller runs `skills path` with no arguments and receives that Skill's materialized directory.
3. The command tree does not include `skills list`.
4. `skills get <name>` and `skills path <name>` with the correct name still succeed; a wrong name fails as today.

When the effective Skill bundle contains two or more Skills, behavior is unchanged: `list` enumerates, and `get` / `path` require a name.

## User-visible states and failure behavior

- Singleton, zero arguments: `get` and `path` operate on the sole Skill.
- Singleton, correct name: `get` and `path` succeed as today.
- Singleton, wrong name: `get` and `path` fail with the existing unknown-Skill error.
- Singleton: `skills list` is not a command.
- Multi-skill, missing name: `get` and `path` fail as today (name required).
- Multi-skill, unknown name: existing unknown-Skill error.
- Empty bundle via the Go API is not a generated-CLI case (every generated CLI has a built-in Skill). `list` remains available; `get` / `path` still require a name.

## Minimum acceptance criteria

1. A generated CLI whose effective bundle is the implicit built-in Skill alone accepts `skills get` and `skills path` with no arguments and returns that Skill's instructions and materialized directory.
2. The same CLI does not expose `skills list` in its command tree or help.
3. `skills get <correct-name>` and `skills path <correct-name>` still succeed on a singleton CLI; a wrong name errors as today.
4. A CLI with two or more bundled Skills still exposes `skills list` and requires a name for `get` and `path`.
5. Generated built-in instructions for a singleton bundle document `skills get` and `skills path` without a name and do not instruct the agent to run `skills list`.
6. Generated built-in instructions for a multi-skill bundle continue to document `skills list` and the named `get` / `path` forms.

## Exclusions and resolved product decisions

- Collapse is based on the effective runtime bundle length, not on whether the schema author declared an explicit built-in entry.
- Compatibility is additive for singleton CLIs (zero-argument `get` / `path`) and subtractive only for `skills list` on those CLIs.
- Multi-skill CLIs are fully unchanged.
