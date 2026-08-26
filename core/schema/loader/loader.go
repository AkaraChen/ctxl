package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/schema/internal/schemajson"
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
	var generated schemajson.Document
	if err := json.Unmarshal(raw, &generated); err != nil {
		return schema.Schema{}, &ValidationError{Rule: "generated", Message: err.Error()}
	}
	s, err := fromGenerated(generated)
	if err != nil {
		return schema.Schema{}, err
	}
	s, err = derive(s)
	if err != nil {
		return schema.Schema{}, err
	}
	return s, nil
}

func derive(s schema.Schema) (schema.Schema, error) {
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
		if s.Entities[i].Command.Name == "" {
			s.Entities[i].Command.Name = s.Entities[i].Name
		}
	}

	var builtin schema.Skill
	found := false
	custom := make([]schema.Skill, 0, len(s.Skills))
	for _, item := range s.Skills {
		if item.Type == schema.SkillBuiltin {
			builtin = item
			found = true
			continue
		}
		custom = append(custom, item)
	}
	if !found {
		var err error
		builtin, err = defaultBuiltin()
		if err != nil {
			return schema.Schema{}, err
		}
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
	return s, nil
}

func fromGenerated(doc schemajson.Document) (schema.Schema, error) {
	s := schema.Schema{
		SchemaURL:   stringValue(doc.Schema),
		Name:        string(doc.Name),
		Description: stringValue(doc.Description),
		Generation: schema.Generation{
			Mode:        schema.GenerationMode(doc.Generation.Mode),
			Output:      aliasValue(doc.Generation.Output),
			Module:      stringValue(doc.Generation.Module),
			CtxlVersion: stringValue(doc.Generation.CtxlVersion),
		},
	}
	if doc.Cli != nil {
		s.CLI.Name = aliasValue(doc.Cli.Name)
	}
	if doc.Store != nil {
		s.Store.Name = aliasValue(doc.Store.Name)
	}
	for _, item := range doc.Entities {
		entity := schema.Entity{
			Name:        string(item.Name),
			Kind:        schema.Kind(item.Kind),
			Format:      schema.Format(item.Format),
			Path:        string(item.Path),
			Location:    schema.Location(item.Location),
			Scope:       schema.Scope(item.Scope),
			ID:          string(item.Id),
			Write:       schema.WriteMode(item.Write),
			Section:     stringValue(item.Section),
			Target:      stringValue(item.Target),
			Body:        stringValue(item.Body),
			Description: stringValue(item.Description),
		}
		if item.Command != nil {
			entity.Command.Name = aliasValue(item.Command.Name)
		}
		for _, item := range item.Fields {
			entity.Fields = append(entity.Fields, schema.Field{
				Name:        string(item.Name),
				Type:        schema.FieldType(item.Type),
				Required:    boolValue(item.Required),
				Description: stringValue(item.Description),
			})
		}
		s.Entities = append(s.Entities, entity)
	}
	for i, item := range doc.Skills {
		object, ok := item.(map[string]any)
		if !ok {
			return schema.Schema{}, &ValidationError{Path: fmt.Sprintf("/skills/%d", i), Rule: "generated", Message: "skill must be an object"}
		}
		raw, err := json.Marshal(item)
		if err != nil {
			return schema.Schema{}, fmt.Errorf("encode generated skill %d: %w", i, err)
		}
		kind, _ := object["type"].(string)
		switch schema.SkillType(kind) {
		case schema.SkillBuiltin:
			var value schemajson.BuiltinSkill
			if err := json.Unmarshal(raw, &value); err != nil {
				return schema.Schema{}, &ValidationError{Path: fmt.Sprintf("/skills/%d", i), Rule: "generated", Message: err.Error()}
			}
			s.Skills = append(s.Skills, builtinFromGenerated(value))
		case schema.SkillCustom:
			var value schemajson.CustomSkill
			if err := json.Unmarshal(raw, &value); err != nil {
				return schema.Schema{}, &ValidationError{Path: fmt.Sprintf("/skills/%d", i), Rule: "generated", Message: err.Error()}
			}
			s.Skills = append(s.Skills, schema.Skill{Type: schema.SkillCustom, Directory: string(value.Directory)})
		default:
			return schema.Schema{}, &ValidationError{Path: fmt.Sprintf("/skills/%d/type", i), Rule: "generated", Message: "unsupported skill type"}
		}
	}
	return s, nil
}

func defaultBuiltin() (schema.Skill, error) {
	var value schemajson.BuiltinSkill
	if err := json.Unmarshal([]byte(`{"type":"builtin"}`), &value); err != nil {
		return schema.Skill{}, fmt.Errorf("decode generated builtin defaults: %w", err)
	}
	return builtinFromGenerated(value), nil
}

func builtinFromGenerated(value schemajson.BuiltinSkill) schema.Skill {
	return schema.Skill{
		Type:          schema.SkillBuiltin,
		Directory:     aliasValue(value.Directory),
		Inject:        schema.InjectPosition(value.Inject),
		Name:          aliasValue(value.Name),
		Description:   stringValue(value.Description),
		License:       stringValue(value.License),
		Compatibility: stringValue(value.Compatibility),
		Metadata:      map[string]string(value.Metadata),
		AllowedTools:  stringValue(value.AllowedTools),
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func aliasValue[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

func boolValue(value *bool) bool {
	return value != nil && *value
}
