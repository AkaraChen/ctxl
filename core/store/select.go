package store

import "github.com/AkaraChen/ctxl/core/schema"

// Select returns schema entities available in this store's scope.
// An empty names list means every entity that st.Entity accepts.
// A named entity that is unknown or locked to another scope is an error.
func (st Store) Select(names []string) ([]schema.Entity, error) {
	if len(names) == 0 {
		out := make([]schema.Entity, 0, len(st.Schema.Entities))
		for _, e := range st.Schema.Entities {
			if _, err := st.Entity(e.Name); err != nil {
				continue
			}
			out = append(out, e)
		}
		return out, nil
	}
	out := make([]schema.Entity, 0, len(names))
	seen := map[string]struct{}{}
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		e, err := st.Entity(name)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
