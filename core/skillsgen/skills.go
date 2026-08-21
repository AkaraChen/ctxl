package skillsgen

import (
	"fmt"
	"strings"

	"github.com/AkaraChen/ctxl/core/schema"
)

func Overview(s schema.Schema) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\nname: overview\ndescription: All %s context entities. Load this first, then load one entity skill.\n---\n\n", s.Name)
	fmt.Fprintf(&b, "# %s context layer\n\n", s.Name)
	if s.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", s.Description)
	}
	b.WriteString("This file is an index. Load exactly one entity skill before writing or reading.\n\n")
	b.WriteString("```bash\n")
	for _, e := range s.Entities {
		fmt.Fprintf(&b, "ctxl --schema <file> skills get %s\n", e.Name)
	}
	b.WriteString("```\n\n")
	b.WriteString("| Entity | Kind | Format | When to load |\n| --- | --- | --- | --- |\n")
	for _, e := range s.Entities {
		when := e.Description
		if when == "" {
			when = string(e.Kind) + " " + string(e.Format)
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", e.Name, e.Kind, e.Format, when)
	}
	b.WriteString("\nDo not preload every entity. Project files live under `.<name>/` unless an entity uses `location: root`. Global files live under `~/.<name>/`.\n")
	return b.String()
}

func Entity(s schema.Schema, e schema.Entity) string {
	var b strings.Builder
	desc := e.Description
	if desc == "" {
		desc = fmt.Sprintf("%s %s %s", e.Kind, e.Format, e.Name)
	}
	fmt.Fprintf(&b, "---\nname: %s\ndescription: %s. Do not load other entity skills at the same time.\n---\n\n", e.Name, desc)
	fmt.Fprintf(&b, "# %s\n\n", e.Name)
	fmt.Fprintf(&b, "Call `ctxl`. Do not open the files yourself.\n\n")
	fmt.Fprintf(&b, "Scope: `--scope project` writes `./.%s/`. `--scope global` writes `~/.%s/`.\n", s.Name, s.Name)
	if e.ResolvedLocation() == schema.LocationRoot {
		fmt.Fprintf(&b, "This entity is at the project root: `%s`.\n", e.Path)
	}
	b.WriteString("\n")
	switch {
	case e.Format == schema.FormatSymlink:
		b.WriteString("```bash\n")
		b.WriteString("ctxl --schema <file> --scope project " + e.Name + " show\n")
		b.WriteString("ctxl --schema <file> --scope project " + e.Name + " write\n```\n")
	case e.Kind == schema.KindSingular:
		b.WriteString("```bash\n")
		b.WriteString("ctxl --schema <file> --scope project " + e.Name + " show\n")
		b.WriteString("ctxl --schema <file> --scope project " + e.Name + " write")
		for _, f := range e.Fields {
			if f.Required {
				fmt.Fprintf(&b, " --%s VALUE", f.Name)
			}
		}
		b.WriteString("\n```\n")
	case e.Format == schema.FormatNDJSON:
		b.WriteString("```bash\n")
		fmt.Fprintf(&b, "ctxl --schema <file> %s append", e.Name)
		for _, f := range e.Fields {
			if f.Required {
				fmt.Fprintf(&b, " --%s VALUE", f.Name)
			}
		}
		fmt.Fprintf(&b, "\nctxl --schema <file> %s list\nctxl --schema <file> %s list --full\nctxl --schema <file> %s get --id ID\n```\n", e.Name, e.Name, e.Name)
	default:
		b.WriteString("```bash\n")
		fmt.Fprintf(&b, "ctxl --schema <file> %s create --id ID\n", e.Name)
		fmt.Fprintf(&b, "ctxl --schema <file> %s list\n", e.Name)
		fmt.Fprintf(&b, "ctxl --schema <file> %s get --id ID\n", e.Name)
		fmt.Fprintf(&b, "ctxl --schema <file> %s update --id ID\n", e.Name)
		fmt.Fprintf(&b, "ctxl --schema <file> %s delete --id ID\n```\n", e.Name)
	}
	if len(e.Fields) > 0 {
		b.WriteString("\nFields:\n\n")
		for _, f := range e.Fields {
			req := ""
			if f.Required {
				req = " (required)"
			}
			fmt.Fprintf(&b, "- `%s` (%s)%s", f.Name, f.Type, req)
			if f.Description != "" {
				fmt.Fprintf(&b, ": %s", f.Description)
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\nDo not dump the conversation into object fields.\n")
	return b.String()
}

func SchemaAuthoring() string {
	return `---
name: schema
description: Write or edit a ctxl JSON schema. Use when creating a new named context layer or adding an entity.
---

# Author a ctxl schema

Produce one JSON object. Do not invent file formats outside this shape.

Required keys:

- ` + "`name`" + `: store name. Project files go to ` + "`.${name}/`" + `, global files go to ` + "`~/.${name}/`" + `.
- ` + "`entities`" + `: array of entities.

Each entity:

- ` + "`name`" + `, ` + "`kind`" + ` (` + "`singular`" + ` or ` + "`plural`" + `), ` + "`format`" + ` (` + "`markdown`" + `, ` + "`ndjson`" + `, or ` + "`symlink`" + `), ` + "`path`" + `
- ` + "`location`" + `: ` + "`root`" + ` (project root, including plural folders) or ` + "`store`" + ` (under the dotted directory). Default ` + "`store`" + `.
- ` + "`scope`" + `: ` + "`project`" + `, ` + "`global`" + `, or ` + "`both`" + `. Default ` + "`both`" + `.
- ` + "`write`" + `: ` + "`replace`" + ` (overwrite the file) or ` + "`section`" + ` (one heading in an existing file). Default ` + "`replace`" + `.
- ` + "`section`" + `: heading text when ` + "`write`" + ` is ` + "`section`" + `.
- ` + "`body`" + `: default section text used by ` + "`init`" + ` the first time.
- ` + "`target`" + `: destination path when ` + "`format`" + ` is ` + "`symlink`" + `.
- ` + "`fields`" + `: ` + "`name`" + `, ` + "`type`" + ` (` + "`string`" + `|` + "`int`" + `|` + "`object`" + `), optional ` + "`required`" + `

Rules:

- Singular markdown ` + "`replace`" + ` overwrites the file.
- Singular markdown ` + "`section`" + ` updates one heading. Headings inside fences are ignored. Missing file is created first.
- Singular ` + "`symlink`" + ` creates a link. Already linked is a no-op. Missing target is an error. An existing non-link is refused.
- Plural ` + "`ndjson`" + ` is one append-only file.
- Plural ` + "`markdown`" + ` is a folder of ` + "`<id>.md`" + ` files.
- One entity is one skill. After writing the JSON, tell the user to run ` + "`ctxl --schema FILE skills get overview`" + `.

Create every declared path with ` + "`ctxl --schema FILE init`" + `. Existing files are left alone unless ` + "`--force`" + `.

Validate with ` + "`ctxl schema validate --schema FILE`" + `.
`
}

func All(s schema.Schema) map[string]string {
	out := map[string]string{
		"overview": Overview(s),
		"schema":   SchemaAuthoring(),
	}
	for _, e := range s.Entities {
		out[e.Name] = Entity(s, e)
	}
	return out
}
