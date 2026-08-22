package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AkaraChen/ctxl/core/schema"
)

const (
	MatchIdentity = "identity"
	MatchContent  = "content"
)

// ErrNoMatch is returned after a successful empty grep so the CLI can exit 1
// without treating it as a hard failure message.
var ErrNoMatch = errors.New("no matches")

type GrepOptions struct {
	Regexp     bool
	IgnoreCase bool
}

type Hit struct {
	Entity  string `json:"entity"`
	ID      string `json:"id,omitempty"`
	Kind    string `json:"kind"`
	Snippet string `json:"snippet"`
}

type matcher struct {
	needle string
	re     *regexp.Regexp
	fold   bool
}

func newMatcher(pattern string, regex, fold bool) (matcher, error) {
	if regex {
		pat := pattern
		if fold {
			pat = "(?i)" + pattern
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return matcher{}, fmt.Errorf("invalid regular expression: %w", err)
		}
		return matcher{re: re}, nil
	}
	needle := pattern
	if fold {
		needle = strings.ToLower(pattern)
	}
	return matcher{needle: needle, fold: fold}, nil
}

func (m matcher) find(s string) bool {
	if m.re != nil {
		return m.re.MatchString(s)
	}
	if m.fold {
		return strings.Contains(strings.ToLower(s), m.needle)
	}
	return strings.Contains(s, m.needle)
}

// Grep searches identity and content of the selected entities.
func (st Store) Grep(entities []schema.Entity, pattern string, opts GrepOptions) ([]Hit, error) {
	m, err := newMatcher(pattern, opts.Regexp, opts.IgnoreCase)
	if err != nil {
		return nil, err
	}
	var hits []Hit
	for _, e := range entities {
		hits = append(hits, identityHit(e, "", e.Name, m)...)
		hits = append(hits, identityHit(e, "", e.Path, m)...)
		more, err := st.grepEntity(e, m)
		if err != nil {
			return nil, err
		}
		hits = append(hits, more...)
	}
	return hits, nil
}

func (st Store) grepEntity(e schema.Entity, m matcher) ([]Hit, error) {
	switch {
	case e.Format == schema.FormatSymlink:
		row, err := st.ReadSymlink(e)
		if err != nil {
			if isAbsent(err) {
				return nil, nil
			}
			return nil, err
		}
		var hits []Hit
		for _, key := range []string{"target", "linked", "path"} {
			if v, ok := row[key]; ok {
				hits = append(hits, contentHit(e, "", fmt.Sprint(v), m)...)
			}
		}
		return hits, nil
	case e.Kind == schema.KindSingular:
		rec, err := st.ReadSingular(e)
		if err != nil {
			if isAbsent(err) {
				return nil, nil
			}
			return nil, err
		}
		return grepRecord(e, "", rec, m), nil
	case e.Format == schema.FormatNDJSON:
		rows, err := st.ListNDJSON(e)
		if err != nil {
			return nil, err
		}
		var hits []Hit
		idField := e.IDField()
		for _, row := range rows {
			id := fmt.Sprint(row[idField])
			hits = append(hits, identityHit(e, id, id, m)...)
			line, err := json.Marshal(row)
			if err != nil {
				return nil, err
			}
			hits = append(hits, contentHit(e, id, string(line), m)...)
		}
		return hits, nil
	default:
		ids, err := st.ListMarkdownItems(e)
		if err != nil {
			return nil, err
		}
		var hits []Hit
		for _, id := range ids {
			hits = append(hits, identityHit(e, id, id, m)...)
			rec, err := st.ReadMarkdownItem(e, id)
			if err != nil {
				return nil, err
			}
			hits = append(hits, grepRecord(e, id, rec, m)...)
		}
		return hits, nil
	}
}

func grepRecord(e schema.Entity, id string, rec Record, m matcher) []Hit {
	var hits []Hit
	seen := map[string]struct{}{}
	for _, f := range e.Fields {
		seen[f.Name] = struct{}{}
		hits = append(hits, contentHit(e, id, f.Name+": "+rec.Fields[f.Name], m)...)
	}
	for k, v := range rec.Fields {
		if _, ok := seen[k]; ok {
			continue
		}
		hits = append(hits, contentHit(e, id, k+": "+v, m)...)
	}
	for _, line := range strings.Split(rec.Body, "\n") {
		hits = append(hits, contentHit(e, id, line, m)...)
	}
	return hits
}

func identityHit(e schema.Entity, id, value string, m matcher) []Hit {
	if !m.find(value) {
		return nil
	}
	return []Hit{{Entity: e.Name, ID: id, Kind: MatchIdentity, Snippet: value}}
}

func contentHit(e schema.Entity, id, line string, m matcher) []Hit {
	if !m.find(line) {
		return nil
	}
	return []Hit{{Entity: e.Name, ID: id, Kind: MatchContent, Snippet: line}}
}

func isAbsent(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "never written") ||
		strings.Contains(s, "not found") ||
		strings.Contains(s, "no section")
}
