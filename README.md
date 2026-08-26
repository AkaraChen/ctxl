# ctxl

`ctxl` is a Go code generator and runtime library for schema-specialized context CLIs. A user-owned JSON schema is the single source of truth for the generated command name, store, entities, output layout, and bundled Agent Skills.

The generic `ctxl` command is a developer tool. Runtime users execute the generated CLI, which embeds exactly one schema and never accepts `--schema`.

## Generate a CLI

```go
//go:generate go run github.com/AkaraChen/ctxl/cmd/ctxl@latest generate context.schema.json
```

```bash
go generate ./...
cd generated/contextctl && go build -o contextctl .
```

All generation configuration lives in the schema, not in `go:generate`. A minimal schema:

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
      "fields": [{"name": "service", "type": "string", "required": true}]
    }
  ]
}
```

The default mode creates `generated/<name>/` as an independent, buildable Go module. Output is generated-owned and replaced in full; `ctxl` refuses to replace an unmarked non-empty directory. Set `generation.mode` to `"existing-module"` to generate only `cmd/<name>` under the surrounding module instead — the parent must already `go get github.com/AkaraChen/ctxl`, and the generator never edits its `go.mod`.

Top-level `name` supplies the CLI, store, module, built-in Skill, and output names; each can be overridden via `cli.name`, `store.name`, `generation.*`, and per-entity `command.name`. The canonical Draft 2020-12 contract is `core/schema/loader/ctxl.schema.json`.

## Skills

`skills` is an array with two entry types:

- `builtin`: the one ctxl-generated command Skill. May supply a source directory, frontmatter overrides, and `inject: "before" | "after"`. If omitted, ctxl creates it implicitly.
- `custom`: exactly `type` and `directory`. Each directory is a complete Agent Skill bundled without rewriting any file.

Generated CLIs expose the bundle:

```bash
contextctl skills list
contextctl skills get context-agent
contextctl skills path deploy   # materializes the full directory in a content-addressed cache
```

## Runtime entities

Generated CLIs keep `--scope project|global`, `init`, and schema-derived commands:

- singular markdown or symlink: `write`, `show`;
- plural NDJSON: `append`, `list`, `get`;
- plural markdown: `create`, `update`, `list`, `get`, `delete`.

Project store files live under `.<store-name>/`; global files under `~/.<store-name>/`. `location: "root"` uses the project root in project scope.

## Go API

Downstream code can construct a specialized CLI without code generation:

```go
s, err := loader.Parse(embeddedSchema)          // core/schema/loader
bundle, err := skillsgen.DefaultBundle(s)        // core/skillsgen
return cli.New(cli.Options{Schema: s, Skills: bundle}).Execute()
```

## Development

This repository contains no generated source; there is nothing to regenerate before building.

```bash
go build ./...
go test ./...
go vet ./...
```
