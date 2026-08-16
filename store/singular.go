package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AkaraChen/ctxl/schema"
)

type Record struct {
	Fields map[string]string
	Body   string
}

// Public flattens frontmatter keys to the top level so show prints
// service/port/start/stop/last_green, not an internal Fields/Body wrapper.
func (r Record) Public() map[string]any {
	out := make(map[string]any, len(r.Fields)+1)
	for k, v := range r.Fields {
		out[k] = v
	}
	if body := strings.TrimSpace(r.Body); body != "" {
		out["body"] = body
	}
	return out
}

func (st Store) WriteSingular(e schema.Entity, rec Record) error {
	path, err := st.EntityPath(e)
	if err != nil {
		return err
	}
	if err := st.EnsureParent(path); err != nil {
		return err
	}
	if rec.Fields == nil {
		rec.Fields = map[string]string{}
	}
	for _, f := range e.Fields {
		if f.Name == "last_green" && strings.TrimSpace(rec.Fields["last_green"]) == "" {
			rec.Fields["last_green"] = time.Now().Format(time.RFC3339)
		}
	}
	var b strings.Builder
	b.WriteString("---\n")
	for _, f := range e.Fields {
		fmt.Fprintf(&b, "%s: %s\n", f.Name, rec.Fields[f.Name])
	}
	b.WriteString("---\n")
	if body := strings.TrimSpace(rec.Body); body != "" {
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func (st Store) ReadSingular(e schema.Entity) (Record, error) {
	path, err := st.EntityPath(e)
	if err != nil {
		return Record{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Record{}, fmt.Errorf("no %s (treat as never written)", e.Path)
		}
		return Record{}, err
	}
	return parseFrontmatter(string(raw))
}

func parseFrontmatter(text string) (Record, error) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "---") {
		return Record{}, fmt.Errorf("missing frontmatter")
	}
	rest := strings.TrimPrefix(text, "---")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Record{}, fmt.Errorf("unclosed frontmatter")
	}
	head := rest[:end]
	body := strings.TrimSpace(rest[end+len("\n---"):])
	fields := map[string]string{}
	for _, line := range strings.Split(head, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return Record{Fields: fields, Body: body}, nil
}
