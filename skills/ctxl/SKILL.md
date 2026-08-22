---
name: ctxl
description: Generic context layer CLI. Load current instructions from the installed binary before running commands.
---

# ctxl

This file is a discovery stub, not the usage guide. Before running any `ctxl` command, load the skill that matches the schema you are using:

```bash
ctxl --schema FILE skills get overview
ctxl --schema FILE skills get schema
ctxl --schema FILE skills get <entity>
```

Use `ctxl tree` and `ctxl grep` for store orientation and search. Do not call host `grep` or `tree`.

One entity is one skill. The overview lists every entity. Do not preload all of them.
