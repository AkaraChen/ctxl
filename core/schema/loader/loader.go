package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkaraChen/ctxl/core/schema"
)

func LoadFile(path string) (schema.Schema, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return schema.Schema{}, err
	}
	s, err := Parse(raw)
	if err != nil {
		var validation *ValidationError
		if errors.As(err, &validation) {
			withSource := *validation
			withSource.Source = path
			return schema.Schema{}, &withSource
		}
		return schema.Schema{}, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

func Parse(raw []byte) (schema.Schema, error) {
	if err := validateJSON(raw); err != nil {
		return schema.Schema{}, err
	}
	var s schema.Schema
	if err := json.Unmarshal(raw, &s); err != nil {
		return schema.Schema{}, &ValidationError{Rule: "decode", Message: err.Error()}
	}
	return derive(s), nil
}

// derive applies the canonical schema's literal defaults and the cross-property
// derivations (effective names, output paths, implicit built-in Skill).
func derive(s schema.Schema) schema.Schema {
	if s.Generation.Mode == "" {
		s.Generation.Mode = schema.GenerationStandalone
	}
	if s.CLI.Name == "" {
		s.CLI.Name = s.Name
	}
	if s.Store.Name == "" {
		s.Store.Name = s.Name
	}
	if s.Generation.Mode == schema.GenerationStandalone && s.Generation.Module == "" {
		s.Generation.Module = s.Name
	}
	if s.Generation.Output == "" {
		if s.Generation.Mode == schema.GenerationExistingModule {
			s.Generation.Output = filepath.Join("cmd", s.CLI.Name)
		} else {
			s.Generation.Output = filepath.Join("generated", s.CLI.Name)
		}
	}
	for i := range s.Entities {
		e := &s.Entities[i]
		if e.Command.Name == "" {
			e.Command.Name = e.Name
		}
		if e.Location == "" {
			e.Location = schema.LocationStore
		}
		if e.Scope == "" {
			e.Scope = schema.ScopeBoth
		}
		if e.ID == "" {
			e.ID = "id"
		}
		if e.Write == "" {
			e.Write = schema.WriteReplace
		}
	}

	builtin := schema.Skill{Type: schema.SkillBuiltin}
	custom := make([]schema.Skill, 0, len(s.Skills))
	for _, item := range s.Skills {
		if item.Type == schema.SkillBuiltin {
			builtin = item
			continue
		}
		custom = append(custom, item)
	}
	if builtin.Inject == "" {
		builtin.Inject = schema.InjectAfter
	}
	if builtin.Name == "" {
		builtin.Name = s.Name
	}
	if builtin.Description == "" && builtin.Directory == "" {
		builtin.Description = s.Description
		if builtin.Description == "" {
			builtin.Description = "Use the " + s.CLI.Name + " context CLI."
		}
	}
	s.Skills = append([]schema.Skill{builtin}, custom...)
	return s
}
