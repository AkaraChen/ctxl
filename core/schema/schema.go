package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Kind string

const (
	KindSingular Kind = "singular"
	KindPlural   Kind = "plural"
)

type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatNDJSON   Format = "ndjson"
	FormatSymlink  Format = "symlink"
)

type WriteMode string

const (
	WriteReplace WriteMode = "replace"
	WriteSection WriteMode = "section"
)

type Scope string

const (
	ScopeProject Scope = "project"
	ScopeGlobal  Scope = "global"
	ScopeBoth    Scope = "both"
)

type Location string

const (
	LocationRoot  Location = "root"
	LocationStore Location = "store"
)

type BackendType string

const (
	BackendFilesystem BackendType = "filesystem"
)

type Backend struct {
	Type BackendType `json:"type"`
}

type FieldType string

const (
	TypeString FieldType = "string"
	TypeInt    FieldType = "int"
	TypeObject FieldType = "object"
)

type Field struct {
	Name        string    `json:"name"`
	Type        FieldType `json:"type"`
	Required    bool      `json:"required,omitempty"`
	Description string    `json:"description,omitempty"`
}

type Entity struct {
	Name        string   `json:"name"`
	Kind        Kind     `json:"kind"`
	Format      Format   `json:"format"`
	Path        string   `json:"path"`
	Location    Location `json:"location,omitempty"`
	Scope       Scope    `json:"scope,omitempty"`
	ID          string    `json:"id,omitempty"`
	Write       WriteMode `json:"write,omitempty"`
	Section     string    `json:"section,omitempty"`
	Target      string    `json:"target,omitempty"`
	Body        string    `json:"body,omitempty"`
	Description string    `json:"description,omitempty"`
	Fields      []Field   `json:"fields"`
}

type Schema struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Backend     Backend  `json:"backend"`
	Entities    []Entity `json:"entities"`
}

func LoadFile(path string) (Schema, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Schema{}, err
	}
	return Parse(raw)
}

func Parse(raw []byte) (Schema, error) {
	var s Schema
	if err := json.Unmarshal(raw, &s); err != nil {
		return Schema{}, fmt.Errorf("schema json: %w", err)
	}
	if err := s.Validate(); err != nil {
		return Schema{}, err
	}
	return s, nil
}

func (s Schema) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("schema.name is required")
	}
	if s.Backend.Type == "" {
		return fmt.Errorf("backend.type is required")
	}
	if s.Backend.Type != BackendFilesystem {
		return fmt.Errorf("backend.type %q is not supported", s.Backend.Type)
	}
	if len(s.Entities) == 0 {
		return fmt.Errorf("schema.entities must not be empty")
	}
	seen := map[string]bool{}
	for i, e := range s.Entities {
		if e.Name == "" {
			return fmt.Errorf("entities[%d].name is required", i)
		}
		if seen[e.Name] {
			return fmt.Errorf("duplicate entity %q", e.Name)
		}
		seen[e.Name] = true
		if e.Kind != KindSingular && e.Kind != KindPlural {
			return fmt.Errorf("entity %q: kind must be singular or plural", e.Name)
		}
		if e.Format != FormatMarkdown && e.Format != FormatNDJSON && e.Format != FormatSymlink {
			return fmt.Errorf("entity %q: format must be markdown, ndjson, or symlink", e.Name)
		}
		if e.Kind == KindSingular && e.Format != FormatMarkdown && e.Format != FormatSymlink {
			return fmt.Errorf("entity %q: singular supports markdown or symlink", e.Name)
		}
		if e.Kind == KindPlural && e.Format == FormatSymlink {
			return fmt.Errorf("entity %q: symlink is singular", e.Name)
		}
		if e.Write != "" && e.Write != WriteReplace && e.Write != WriteSection {
			return fmt.Errorf("entity %q: write must be replace or section", e.Name)
		}
		if e.ResolvedWrite() == WriteSection {
			if e.Kind != KindSingular || e.Format != FormatMarkdown {
				return fmt.Errorf("entity %q: write section is singular markdown", e.Name)
			}
			if strings.TrimSpace(e.Section) == "" {
				return fmt.Errorf("entity %q: section heading is required", e.Name)
			}
		}
		if e.Format == FormatSymlink && strings.TrimSpace(e.Target) == "" {
			return fmt.Errorf("entity %q: symlink target is required", e.Name)
		}
		if e.Kind == KindPlural && e.Format == FormatNDJSON && e.Path == "" {
			return fmt.Errorf("entity %q: ndjson path is required", e.Name)
		}
		if e.Kind == KindPlural && e.Format == FormatMarkdown && e.Path == "" {
			return fmt.Errorf("entity %q: markdown collection path is required", e.Name)
		}
		if e.Kind == KindSingular && e.Path == "" {
			return fmt.Errorf("entity %q: path is required", e.Name)
		}
	}
	return nil
}

func (s Schema) Entity(name string) (Entity, error) {
	for _, e := range s.Entities {
		if e.Name == name {
			return e, nil
		}
	}
	return Entity{}, fmt.Errorf("unknown entity %q", name)
}

func (e Entity) ResolvedLocation() Location {
	if e.Location != "" {
		return e.Location
	}
	return LocationStore
}

func (e Entity) ResolvedScope() Scope {
	if e.Scope != "" {
		return e.Scope
	}
	return ScopeBoth
}

func (e Entity) IDField() string {
	if e.ID != "" {
		return e.ID
	}
	return "id"
}

func (e Entity) ResolvedWrite() WriteMode {
	if e.Write != "" {
		return e.Write
	}
	return WriteReplace
}
