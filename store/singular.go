package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AkaraChen/ctxlayer/schema"
)

type Record struct {
	Fields map[string]string
	Body   string
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
	if _, ok := rec.Fields["last_green"]; !ok {
		for _, f := range e.Fields {
			if f.Name == "last_green" {
				rec.Fields["last_green"] = time.Now().Format(time.RFC3339)
			}
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
