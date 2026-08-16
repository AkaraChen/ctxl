# ctxl

A reusable context layer: a JSON schema names a store, declares singular and plural entities, and the same binary serves a cobra CLI plus generated agent skills.

## Store layout

The schema `name` is the store name. Project files live under `.<name>/` in the working directory. Global files live under `~/.<name>/`. An entity may set `location: root` to keep a file or a plural folder at the project root. `--scope project|global` selects which tree to use; an entity may also lock itself to one scope.

## Three entity shapes

- **Singular markdown (`write: replace`)** — one file, current state. `write` overwrites; `show` prints the frontmatter keys. If a `last_green` field exists and is empty, write fills the current RFC3339 time.
- **Singular markdown (`write: section`)** — one heading in an existing file. Headings inside fences are ignored. Missing file is created, then the section is written.
- **Singular symlink** — `path` points at `target`. Already linked is a no-op. Missing target is refused. An existing non-link is refused.
- **Plural NDJSON** — one append-only file. `append` writes a line (`id` and `ts` are filled in); `list` prints fixed fields unless `--full`; `get --id` reads one row. Object fields are JSON.
- **Plural markdown** — a folder of `<id>.md` files. `location: root` keeps that folder at the project root.

## CLI and Go package

```bash
go install github.com/AkaraChen/ctxl/cmd/ctxl@latest
```

The CLI does not embed a schema. Pass `--schema FILE`. The example is `examples/demo.schema.json`. The same tree is a Go package: parse your own JSON and ship a branded binary.

```go
s, _ := schema.Parse(embedded)
app.New(app.Options{Name: "mycli", Schema: s}).Execute()
```

That binary gets one command per entity, plus `skills` and `schema validate`.

```bash
ctxl --schema examples/demo.schema.json --scope project status write --service hermes --start up --stop down
ctxl --schema examples/demo.schema.json status show
ctxl --schema examples/demo.schema.json log append --result green --cmd up
ctxl --schema examples/demo.schema.json log list
ctxl --schema examples/demo.schema.json log list --full
ctxl --schema examples/demo.schema.json note create --id n1 --title hello
ctxl --schema examples/demo.schema.json guide write --body "installed notes"
ctxl --schema examples/demo.schema.json alias write
ctxl --schema examples/demo.schema.json schema validate
```

## Skills

Skills are generated from the schema, not checked in as long guides.

- `overview` — index of every entity; load this first
- one skill per entity — load exactly one before writing or reading
- `schema` — how to author the JSON

```bash
ctxl --schema FILE skills get overview
ctxl --schema FILE skills get schema
ctxl --schema FILE skills get <entity>
```

`skills/ctxl/SKILL.md` is a thin stub that only points at those commands.

## Distribution

Serve instructions from the installed binary so `skills get` always matches the installed code. A cached markdown file will drift; the generated skill will not.
