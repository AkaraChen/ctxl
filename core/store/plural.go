package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AkaraChen/ctxl/core/schema"
)

func (st Store) AppendNDJSON(e schema.Entity, fields map[string]any) (map[string]any, error) {
	path, err := st.EntityPath(e)
	if err != nil {
		return nil, err
	}
	if err := st.ensureParent(path); err != nil {
		return nil, err
	}
	existing, err := st.ListNDJSON(e)
	if err != nil {
		return nil, err
	}
	idField := e.ID
	next := 1
	if n := len(existing); n > 0 {
		if v, ok := asInt(existing[n-1][idField]); ok {
			next = v + 1
		} else {
			return nil, fmt.Errorf("%s: last %s must be an integer", e.Name, idField)
		}
	}
	if fields == nil {
		fields = map[string]any{}
	}
	if _, ok := fields[idField]; !ok {
		fields[idField] = next
	}
	if _, ok := fields["ts"]; !ok {
		fields["ts"] = time.Now().Format(time.RFC3339)
	}
	line, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return nil, err
	}
	return fields, nil
}

func (st Store) ListNDJSON(e schema.Entity) ([]map[string]any, error) {
	path, err := st.EntityPath(e)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []map[string]any
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, sc.Err()
}

func (st Store) GetNDJSON(e schema.Entity, id string) (map[string]any, error) {
	all, err := st.ListNDJSON(e)
	if err != nil {
		return nil, err
	}
	idField := e.ID
	for _, row := range all {
		if fmt.Sprint(row[idField]) == id {
			return row, nil
		}
	}
	return nil, fmt.Errorf("%s %s not found", e.Name, id)
}

func (st Store) WriteMarkdownItem(e schema.Entity, id string, rec Record) error {
	dir, err := st.EntityPath(e)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if rec.Fields == nil {
		rec.Fields = map[string]string{}
	}
	rec.Fields[e.ID] = id
	path := filepath.Join(dir, id+".md")
	var b strings.Builder
	b.WriteString("---\n")
	for _, f := range e.Fields {
		fmt.Fprintf(&b, "%s: %s\n", f.Name, rec.Fields[f.Name])
	}
	if rec.Fields[e.ID] != "" && !hasField(e, e.ID) {
		fmt.Fprintf(&b, "%s: %s\n", e.ID, id)
	}
	b.WriteString("---\n")
	if body := strings.TrimSpace(rec.Body); body != "" {
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func (st Store) ReadMarkdownItem(e schema.Entity, id string) (Record, error) {
	dir, err := st.EntityPath(e)
	if err != nil {
		return Record{}, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, id+".md"))
	if err != nil {
		if os.IsNotExist(err) {
			return Record{}, fmt.Errorf("%s %s not found", e.Name, id)
		}
		return Record{}, err
	}
	return parseFrontmatter(string(raw))
}

func (st Store) ListMarkdownItems(e schema.Entity) ([]string, error) {
	dir, err := st.EntityPath(e)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if strings.HasSuffix(name, ".md") {
			ids = append(ids, strings.TrimSuffix(name, ".md"))
		}
	}
	return ids, nil
}

func (st Store) DeleteMarkdownItem(e schema.Entity, id string) error {
	dir, err := st.EntityPath(e)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, id+".md")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s %s not found", e.Name, id)
		}
		return err
	}
	return nil
}

func hasField(e schema.Entity, name string) bool {
	for _, f := range e.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		if math.Trunc(n) != n {
			return 0, false
		}
		return int(n), true
	case int:
		return n, true
	default:
		return 0, false
	}
}

// FixedRows drops object-typed fields so list defaults to scalar columns.
// Pass the original rows when --full is set.
func FixedRows(e schema.Entity, rows []map[string]any) []map[string]any {
	keep := map[string]bool{}
	for _, f := range e.Fields {
		if f.Type != schema.TypeObject {
			keep[f.Name] = true
		}
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		slim := map[string]any{}
		for k, v := range row {
			if keep[k] {
				slim[k] = v
			}
		}
		out = append(out, slim)
	}
	return out
}
