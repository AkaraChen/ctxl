# PRD: CLI grep and tree

Status: accepted and implemented for `backend.type: filesystem`. Do not treat host `grep` / `tree` as a substitute.

## Problem and context

Step 1 left disk files public and gave per-entity `show` / `list` / `get`. Agents who need store-wide orientation or search still fall back to the host's `grep` and `tree`. Those results are hard to optimize, and they are not the CLI.

This feature adds first-party `grep` and `tree` over the ctxl store only. They are the next structured-retrieval verbs, not a general project search.

## Users

Same as `docs/prd/context-layer.md`: anyone who wants an agent to understand project context. The agent (or a human) calls the CLI instead of the host tools.

## Goals

- Search identity and content inside the current filesystem store without calling the host `grep`.
- List the store as a logical entity tree without calling the host `tree`.
- Keep product flags on `ctxl`; keep the verb tail looking like the Unix tool.
- Return machine-readable records so the caller does not have to parse host-tool text.

## Non-goals

- Searching the whole working tree, or any path outside schema-declared entities.
- Invoking the host `grep`, `tree`, or equivalent binaries.
- Changing the schema language or adding a sixth entity shape.
- Descending tree past two levels (store → entity). Plural item ids are not tree children in this cut.
- A second output mode (no parallel `--json` vs text-tree).
- Sync, extra backends, or host-tool feature parity (`-A`/`-B`/`-C`, color, etc.).

## User stories

- As an agent, I run `ctxl tree` and see every entity the schema declares, even if `init` has not created the file yet.
- As an agent, I run `ctxl grep PATTERN` and get structured hits from entity names, ids, paths, and file contents.
- As an agent, I put `--entity` on `ctxl` (not on `grep`/`tree`) when I only care about some entities.
- As a human, I can still open the same files; I use these commands when I want structured retrieval.

## Command shape

```text
ctxl [product flags] grep [grep-like flags] PATTERN
ctxl [product flags] tree [tree-like flags if any; none required in this cut]
```

Product flags stay in front and are shared:

- `--schema FILE` (already shipped)
- `--scope project|global` (already shipped; default `project`)
- entity filter on `ctxl` (repeatable; unknown name is an error)

`grep` and `tree` are orthogonal verbs over the same selected entity set. They share selection rules and should share selection code.

## Grep

Default match is a literal substring (like `grep -F`). `-E` on the grep tail switches to regular expressions.

Corpus, in the selected entities:

- Identity: entity name, item id (when the shape has items), and the entity path.
- Content: singular body / section text, NDJSON row text, plural markdown body and frontmatter.

A hit is one structured record:

- `entity`
- optional item id
- whether the match was identity or content
- a snippet of the match

Granularity follows grep: one record per matching content line; one record per matching identity field. One item with three matching lines yields three hits.

No matches: empty result list, non-zero exit (grep-like).

## Tree

Logical entity tree, two levels: store → each selected entity. Files that do not exist still appear.

Each node is one structured record in the same family as a grep hit (JSON object), not an ASCII `tree` drawing. The record identifies the entity; extra fields may come from the shared selection walk so long as the output stays one object per entity.

This cut does not nest plural items under an entity. Later cuts may add levels without changing the two-verb, shared-selection rule.

## Skills

Generated skills must tell the agent to call `ctxl grep` and `ctxl tree`, not the host tools. `overview` (and any dedicated store-wide skill) is in scope. Per-entity skills stay the place to load before a write.

## Failure behavior

- Unknown entity filter: error, no partial walk.
- Invalid `--scope`: same as today.
- Invalid regex with `-E`: error.
- Entity locked to the other filesystem scope: skip or refuse consistently with the shared store walk used by other verbs; do not search a path that does not belong to this scope.
- Missing content files: no content hits; identity can still match; tree still lists the entity.

## Acceptance criteria

- `ctxl grep PATTERN` does not exec host `grep`. `ctxl tree` does not exec host `tree`.
- With no entity filter, both verbs consider every schema entity available in the current `--scope`.
- `ctxl --entity note grep PATTERN` and `ctxl --entity note tree` use the same selected set. `ctxl grep --entity note PATTERN` is not the product form.
- Literal `PATTERN` matches a substring in identity and in content. `-E` treats `PATTERN` as a regular expression.
- A content line that matches three times on separate lines yields three hits. A matching entity name yields one identity hit.
- Each grep hit JSON object includes entity, match kind (identity or content), snippet, and item id when the match belongs to a plural item.
- `ctxl tree` prints one JSON object per selected entity, including entities whose files are absent.
- `ctxl tree` does not list plural item ids as children.
- `skills get` materials tell the agent to use these verbs instead of host `grep`/`tree`.
- Branded binaries built with `cli.New` expose the same two verbs.

## Resolved product decisions

- Store-only, not workspace-wide.
- Tree is logical entities, two levels, reserved for later depth.
- Grep searches identity and content.
- Hits are structured records, grep-like granularity.
- Default literal; `-E` for regex; grep tail looks like grep.
- Product flags on `ctxl`; `grep`/`tree` share those rules and code.
- Output is structured for both verbs (already decided with "结构化一条" + shared verbs). ASCII tree text is not a second mode.
