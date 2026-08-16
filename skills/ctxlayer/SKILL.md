---
name: ctxlayer
description: Generic context layer. Load current instructions from the installed binary before reading or writing any entity. Use when authoring a schema, inspecting a named store, or shipping a branded CLI.
---

# ctxlayer

This file is a discovery stub, not the usage guide. Before running any `ctxlayer` command, load the skill that matches the schema you are using:

```bash
ctxlayer --schema ... skills get overview
ctxlayer --schema ... skills get schema
ctxlayer --schema ... skills get <entity>
```

`overview` is the index. `schema` is how to author a JSON schema. Each entity is its own skill — load exactly one before writing or reading. The CLI generates these from the schema, so do not rely on a cached copy of this stub.
