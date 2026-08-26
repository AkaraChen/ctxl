# ctxl

`ctxl` is a Go code generator and runtime library for schema-specialized context CLIs. A user-owned JSON schema is the single source of truth for the generated command name, store, entities, output layout, and bundled Agent Skills.

The generic `ctxl` command is a developer tool. Runtime users execute the generated CLI, which embeds exactly one schema and never accepts `--schema`.

## Generate a CLI

Install a released generator or invoke the same version directly from `go:generate`:

```go
//go:generate go run github.com/AkaraChen/ctxl/cmd/ctxl@latest generate context.schema.json
```

For reproducible generation, replace `latest` with the same pinned ctxl version used by the project.

```bash
go generate ./...
cd generated/contextctl
go build -o contextctl .
```

Generation configuration belongs in `context.schema.json`, not in `go:generate`. The default mode creates `generated/<name>/` as an independent Go module:

```json
{
  "name": "contextctl",
  "description": "Manage project context.",
  "entities": [
    {
      "name": "status",
      "kind": "singular",
      "format": "markdown",
      "path": "STATUS.md",
      "location": "root",
      "scope": "project",
      "fields": [
        {"name": "service", "type": "string", "required": true}
      ]
    }
  ]
}
```

The output is generated-owned and replaced in full. `ctxl` refuses to replace an unmarked non-empty directory.

## Existing Go module mode

Set `generation.mode` to generate only a command package under the surrounding module. The default output becomes `cmd/<name>`:

```json
{
  "name": "contextctl",
  "generation": {
    "mode": "existing-module"
  },
  "entities": [
    {"name": "status", "kind": "singular", "format": "markdown", "path": "STATUS.md"}
  ]
}
```

The parent module must already require that ctxl version:

```bash
go get github.com/AkaraChen/ctxl@latest
go generate ./...
go build ./cmd/contextctl
```

The generator never edits the parent `go.mod` or `go.sum`.

## Names and overrides

By default, the top-level `name` supplies the generated CLI name, store name, standalone module name, built-in Skill name, and output name. Every identity can be overridden in the schema:

```json
{
  "name": "context",
  "cli": {"name": "contextctl"},
  "store": {"name": "context-data"},
  "generation": {
    "mode": "standalone",
    "output": "tools/contextctl",
    "module": "example.com/tools/contextctl"
  },
  "skills": [
    {"type": "builtin", "name": "context-agent"}
  ],
  "entities": [
    {
      "name": "status",
      "command": {"name": "current"},
      "kind": "singular",
      "format": "markdown",
      "path": "STATUS.md"
    }
  ]
}
```

All relative paths resolve from the schema file.

## Skills

`skills` is an array with two entry types:

- `builtin` configures the one ctxl-generated command Skill. It may supply a complete source directory, frontmatter overrides, and `inject: "before" | "after"`. If omitted, ctxl creates the default built-in Skill.
- `custom` contains exactly `type` and `directory`. Any number may be declared. Each directory is a complete Agent Skill and is bundled without rewriting any file.

```json
{
  "skills": [
    {
      "type": "builtin",
      "directory": "skills/context-agent",
      "inject": "after",
      "name": "context-agent",
      "description": "Use contextctl safely.",
      "license": "Apache-2.0",
      "compatibility": "Requires contextctl on PATH.",
      "metadata": {"owner": "platform"},
      "allowed-tools": "Bash(contextctl:*)"
    },
    {"type": "custom", "directory": "skills/deploy"},
    {"type": "custom", "directory": "skills/debug"}
  ]
}
```

Every source directory follows the Agent Skills layout and contains `SKILL.md`. Supporting scripts, references, assets, binary files, empty directories, and executable modes are embedded in the generated binary. Generated CLIs expose:

```bash
contextctl skills list
contextctl skills get context-agent
contextctl skills path deploy
```

`skills path` atomically materializes the complete selected Skill in a content-addressed user cache so relative references and scripts work normally.

## Runtime entities

Generated CLIs keep `--scope project|global`, `init`, and schema-derived commands:

- singular markdown or symlink: `write`, `show`;
- plural NDJSON: `append`, `list`, `get`;
- plural markdown: `create`, `update`, `list`, `get`, `delete`.

Project store files live under `.<effective-store-name>/`; global files live under `~/.<effective-store-name>/`. An entity with `location: "root"` uses the project root in project scope.

## Go API

Downstream code can construct a specialized CLI without code generation:

```go
import (
    "github.com/AkaraChen/ctxl/cli"
    "github.com/AkaraChen/ctxl/core/schema/loader"
    "github.com/AkaraChen/ctxl/core/skillsgen"
)

s, err := loader.Parse(embeddedSchema)
if err != nil {
    return err
}
bundle, err := skillsgen.DefaultBundle(s)
if err != nil {
    return err
}
return cli.New(cli.Options{Schema: s, Skills: bundle}).Execute()
```

The canonical Draft 2020-12 schema is checked in at `core/schema/loader/ctxl.schema.json`. Regenerate its Go decoder and validation with `go generate ./core/schema/internal/schemajson`.

## Development

```bash
go build ./...
go test ./...
go vet ./...
```
