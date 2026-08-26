# Example: hnm as a ctxl schema

Models [AkaraChen/hnm](https://github.com/AkaraChen/hnm), a Go CLI that installs an agent documentation harness (`AGENTS.md`, `CLAUDE.md` symlink, `docs/spec.md`, `docs/prd/`, `docs/adr/`, plus workflow skills) into a project. hnm already declares this data model in ctxl's schema shape (`schema/harness.json`); this example completes it into a generatable CLI.

Mapping:

- `hnm init` → generated `hnmctl init`: creates `AGENTS.md`, the `CLAUDE.md → AGENTS.md` symlink, `docs/spec.md`, and the `docs/prd/` / `docs/adr/` directories, skipping paths that already exist (hnm's non-`--force` behavior).
- `docs/prd/` and `docs/adr/` are plural markdown collections keyed by kebab-case filename → `hnmctl prd create --id cli-run ...`, `hnmctl adr list`, and so on.
- hnm's `feature-dev` and `git-commit` skills are bundled verbatim as custom Skills (`hnmctl skills get feature-dev`), taken from hnm's `templates/.agents/skills/`.

Not expressible in ctxl today, kept in hnm itself: template rendering of `AGENTS.md`/`docs/spec.md` content (`--name`, `--stack` presets), frontmatter-free singular files (ctxl always writes YAML frontmatter), static config artifacts (`.claude/settings.json`, `.codex/hooks*`), and the directory-level `.claude/skills → ../.agents/skills` symlink (ctxl symlink targets resolve from the project root).

Generate with a released ctxl (or in-repo via the `TestExamples` fixture):

```go
//go:generate go run github.com/AkaraChen/ctxl/cmd/ctxl@latest generate context.schema.json
```
