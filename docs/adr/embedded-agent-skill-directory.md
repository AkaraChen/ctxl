# Bundle a complete Agent Skill directory

Status: Accepted

## Context

Current Skill generation returns Markdown strings. A complete Agent Skill is a directory whose required `SKILL.md` may refer to scripts, references, assets, or other files by relative path. A single Go binary cannot rely on separately installed package data, but returning only concatenated Markdown breaks those relative resources and executable files.

The generated CLI must always carry version-matched instructions for its generated commands while also allowing users to bundle several independent, arbitrary Skills without ctxl rewriting them.

## Decision

- Model `skills` as a discriminated array with `builtin` and `custom` entries.
- Normalize every schema to exactly one built-in Skill. A missing built-in entry creates the default; more than one is invalid.
- Let a built-in entry configure its name and Agent Skill frontmatter, an optional source directory, and instruction injection. Without a directory, synthesize the complete Skill.
- Generate CLI-specific instructions from the effective schema and inject them only into the built-in `SKILL.md` at the schema-selected boundary:
  - `before` means immediately after YAML frontmatter;
  - `after` means after the existing Markdown body.
- Define a custom entry as exactly `type: "custom"` plus `directory`. Allow any number of custom entries.
- Package custom Skill directories without modifying `SKILL.md`, frontmatter, name, or other contents.
- Resolve all source directories relative to the schema file and preserve their file trees and executable modes.
- Validate the transformed built-in output and each unchanged custom Skill against the normative Agent Skills specification, including required frontmatter, name constraints, directory-name agreement, and uniqueness across the bundle.
- Reject symlinks and paths that could escape the declared Skill root.
- Embed a deterministic file manifest and contents in the generated binary.
- Expose `skills list`, `skills get`, and `skills path`. `get` prints the effective `SKILL.md`; `path` atomically materializes the complete content-addressed directory into the user cache and returns it.
- Use a content digest, not a mutable CLI name alone, as the cache identity. Never expose a partially extracted directory.

The official `skills-ref` implementation may be used as a conformance oracle in fixtures, but it is explicitly a demonstration reference rather than a production SDK. The Go runtime therefore must not require Python or an external `skills-ref` executable.

## Alternatives considered

### Append an inline Markdown string from JSON

Rejected. It is unpleasant to author, cannot carry scripts or resources, and invents a sub-format below the Agent Skills directory contract.

### Merge custom Skills into the built-in Skill

Rejected. Custom Skills are independent Agent Skill packages. Merging their frontmatter, instructions, or resources would make ctxl invent precedence rules and prevent users from authoring them freely.

### Concatenate every bundled file into `skills get --full`

Rejected as the only access method. Concatenation destroys file identity, relative links, executable scripts, and non-text assets.

### Install Skill files beside the binary

Rejected as the required distribution model. `go install` installs a binary, not arbitrary sibling data, and external files can drift from the installed executable.

### Shell out to `skills-ref` at generation or runtime

Rejected as a required dependency. The official repository labels it demonstration-only, and a Go code generator must remain buildable without a Python environment.

## Consequences

- A generated binary is self-contained while agents can still work with a normal directory after materialization.
- Built-in injection must parse the YAML frontmatter boundary rather than using raw prefix/suffix concatenation. Custom Skills bypass transformation entirely.
- Binary size grows with bundled assets. That cost is visible and accepted; the schema author owns the selected directory.
- Cache cleanup is not part of this feature. Content addressing prevents version collisions and permits later garbage collection without changing the Skill contract.

## Failure bounds and validation

- Validate every input directory and the transformed built-in output before generating the binary manifest.
- Preserve exact relative paths and test text, binary assets, executable scripts, empty directories where representable, and injection order.
- Reject multiple built-in entries, extra fields on custom entries, duplicate effective Skill names, invalid frontmatter, escaping links, unsupported filesystem entries, and case-colliding paths on supported case-insensitive platforms.
- Test interrupted materialization and concurrent callers; all successful callers must receive the same complete directory.
