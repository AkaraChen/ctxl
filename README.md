# ctxlayer

Generic context layer: a JSON schema names a store, declares singular and plural entities, and the same binary serves a cobra CLI plus generated agent skills.

**yoi is unchanged.** `examples/yoi.schema.json` only shows how that deploy state and audit log would look as a schema. The yoi binary, its files, and its skills stay where they are.

## Store layout

The schema `name` is the store name. Project files live under `.<name>/` in the working directory. Global files live under `~/.<name>/`. An entity may set `location: root` to keep a singular file at the project root (yoi's `DEPLOY.md` does this). `--scope project|global` selects which tree to use; an entity may also lock itself to one scope.

## Three entity shapes

- **Singular markdown** — one file, current state. `write` overwrites; `show` reads. Frontmatter keys come from `fields`.
- **Plural NDJSON** — one append-only file. `append` writes a line (`id` and `ts` are filled in); `list` and `get --id` read. Object fields are JSON.
- **Plural markdown** — a folder of `<id>.md` files. `create` / `update --id`, `list`, `get --id`, `delete --id`.

## CLI and Go package

```bash
go install github.com/AkaraChen/ctxlayer/cmd/ctxlayer@latest
```

The default binary embeds the yoi example schema. Pass `--schema FILE` to drive any other schema (see `examples/notes.schema.json`). The same tree is a Go package: embed your own JSON and ship a branded binary.

```go
s, _ := schema.Parse(embedded)
app.New(app.Options{Name: "mycli", Schema: s}).Execute()
```

That binary gets one command per entity, plus `skills` and `schema validate`.

```bash
ctxlayer --schema examples/yoi.schema.json --scope project deploy write --service hermes --start up --stop down
ctxlayer --schema examples/yoi.schema.json deploy show
ctxlayer --schema examples/yoi.schema.json log append --result green --cmd up
ctxlayer --schema examples/yoi.schema.json log list
ctxlayer --schema examples/notes.schema.json note create --id n1 --title hello
ctxlayer schema validate --schema examples/yoi.schema.json
```

## Skills

Skills are generated from the schema, not checked in as long guides.

- `overview` — index of every entity; load this first
- one skill per entity — load exactly one before writing or reading
- `schema` — how to author the JSON

```bash
ctxlayer --schema FILE skills get overview
ctxlayer --schema FILE skills get schema
ctxlayer --schema FILE skills get <entity>
```

`skills/ctxlayer/SKILL.md` is a thin stub that only points at those commands.

## Lessons from yoi and hnm

yoi taught the file shapes: one current-state markdown, one append-only log, and skills that call the CLI instead of parsing files. [AkaraChen/hnm](https://github.com/AkaraChen/hnm) taught the distribution rule: embed the instructions in the binary so `skills get` always matches the installed code. ctxlayer keeps both. A cached markdown file will drift; the generated skill will not.
