package skillsgen

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/skillbundle"
	"gopkg.in/yaml.v3"
)

var skillNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type skillFrontmatter struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`
}

func DefaultBundle(s schema.Schema) (skillbundle.Bundle, error) {
	if len(s.Skills) == 0 || s.Skills[0].Type != schema.SkillBuiltin {
		return skillbundle.Bundle{}, fmt.Errorf("default bundle requires a normalized schema with a built-in Skill")
	}
	builtin := s.Skills[0]
	builtin.Directory = ""
	s.Skills = []schema.Skill{builtin}
	return Build(s, "")
}

func Build(s schema.Schema, schemaDir string) (skillbundle.Bundle, error) {
	all := make([]skillbundle.Skill, 0, len(s.Skills))
	names := map[string]bool{}
	for _, item := range s.Skills {
		var (
			skill skillbundle.Skill
			err   error
		)
		switch item.Type {
		case schema.SkillBuiltin:
			skill, err = buildBuiltin(s, item, schemaDir)
		case schema.SkillCustom:
			skill, err = buildCustom(item, schemaDir)
		default:
			err = fmt.Errorf("unsupported skill type %q", item.Type)
		}
		if err != nil {
			return skillbundle.Bundle{}, err
		}
		if names[skill.Name] {
			return skillbundle.Bundle{}, fmt.Errorf("duplicate effective skill %q", skill.Name)
		}
		names[skill.Name] = true
		all = append(all, skill)
	}
	return skillbundle.Bundle{Skills: all}, nil
}

func buildBuiltin(s schema.Schema, cfg schema.Skill, schemaDir string) (skillbundle.Skill, error) {
	meta := skillFrontmatter{}
	body := ""
	var entries []skillbundle.Entry
	if cfg.Directory != "" {
		var err error
		entries, err = readDirectory(filepath.Join(schemaDir, cfg.Directory))
		if err != nil {
			return skillbundle.Skill{}, fmt.Errorf("builtin skill %s: %w", cfg.Directory, err)
		}
		raw, err := entryData(entries, "SKILL.md")
		if err != nil {
			return skillbundle.Skill{}, err
		}
		meta, body, err = parseSkillMarkdown(raw)
		if err != nil {
			return skillbundle.Skill{}, fmt.Errorf("builtin skill %s: %w", cfg.Directory, err)
		}
		if err := validateFrontmatter(meta); err != nil {
			return skillbundle.Skill{}, fmt.Errorf("builtin skill %s: %w", cfg.Directory, err)
		}
		if filepath.Base(filepath.Clean(cfg.Directory)) != meta.Name {
			return skillbundle.Skill{}, fmt.Errorf("builtin skill %s: frontmatter name %q must match directory name", cfg.Directory, meta.Name)
		}
	}
	meta.Name = cfg.Name
	if cfg.Description != "" {
		meta.Description = cfg.Description
	}
	if meta.Description == "" {
		meta.Description = s.Description
		if meta.Description == "" {
			meta.Description = "Use the " + s.CLI.Name + " context CLI."
		}
	}
	if cfg.License != "" {
		meta.License = cfg.License
	}
	if cfg.Compatibility != "" {
		meta.Compatibility = cfg.Compatibility
	}
	if cfg.Metadata != nil {
		meta.Metadata = cfg.Metadata
	}
	if cfg.AllowedTools != "" {
		meta.AllowedTools = cfg.AllowedTools
	}
	if meta.AllowedTools == "" {
		meta.AllowedTools = "Bash(" + s.CLI.Name + ":*)"
	}
	if err := validateFrontmatter(meta); err != nil {
		return skillbundle.Skill{}, fmt.Errorf("builtin skill: %w", err)
	}
	generated := instructions(s)
	if cfg.Inject == schema.InjectBefore {
		body = joinMarkdown(generated, body)
	} else {
		body = joinMarkdown(body, generated)
	}
	raw, err := renderSkillMarkdown(meta, body)
	if err != nil {
		return skillbundle.Skill{}, err
	}
	entries = replaceEntry(entries, skillbundle.Entry{Path: "SKILL.md", Mode: 0o644, Data: raw})
	return skillbundle.Skill{Name: meta.Name, Entries: entries}, nil
}

func buildCustom(cfg schema.Skill, schemaDir string) (skillbundle.Skill, error) {
	dir := filepath.Join(schemaDir, cfg.Directory)
	entries, err := readDirectory(dir)
	if err != nil {
		return skillbundle.Skill{}, fmt.Errorf("custom skill %s: %w", cfg.Directory, err)
	}
	raw, err := entryData(entries, "SKILL.md")
	if err != nil {
		return skillbundle.Skill{}, fmt.Errorf("custom skill %s: %w", cfg.Directory, err)
	}
	meta, _, err := parseSkillMarkdown(raw)
	if err != nil {
		return skillbundle.Skill{}, fmt.Errorf("custom skill %s: %w", cfg.Directory, err)
	}
	if err := validateFrontmatter(meta); err != nil {
		return skillbundle.Skill{}, fmt.Errorf("custom skill %s: %w", cfg.Directory, err)
	}
	if filepath.Base(filepath.Clean(dir)) != meta.Name {
		return skillbundle.Skill{}, fmt.Errorf("custom skill %s: frontmatter name %q must match directory name", cfg.Directory, meta.Name)
	}
	return skillbundle.Skill{Name: meta.Name, Entries: entries}, nil
}

func instructions(s schema.Schema) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- ctxl:generated:start -->\n# %s command guide\n\n", s.CLI.Name)
	if s.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", s.Description)
	}
	fmt.Fprintf(&b, "Use `%s`; the schema is built into this executable. Never pass a schema file and do not edit store files directly.\n\n", s.CLI.Name)
	fmt.Fprintf(&b, "- Project scope: `%s --scope project ...` stores data under `.%s/` unless an entity uses a root location.\n", s.CLI.Name, s.Store.Name)
	fmt.Fprintf(&b, "- Global scope: `%s --scope global ...` stores data under `~/.%s/`.\n", s.CLI.Name, s.Store.Name)
	fmt.Fprintf(&b, "- Initialize declared paths: `%s init`.\n\n", s.CLI.Name)
	fmt.Fprintf(&b, "Bundled Skills:\n\n- List: `%s skills list`.\n- Read instructions: `%s skills get NAME`.\n- Materialize a complete Skill directory: `%s skills path NAME`.\n\n", s.CLI.Name, s.CLI.Name, s.CLI.Name)
	for _, e := range s.Entities {
		fmt.Fprintf(&b, "## %s\n\n", e.Command.Name)
		if e.Description != "" {
			fmt.Fprintf(&b, "%s\n\n", e.Description)
		}
		writeEntityCommands(&b, s.CLI.Name, e)
		if len(e.Fields) > 0 {
			b.WriteString("\nFields:\n\n")
			for _, field := range e.Fields {
				required := ""
				if field.Required {
					required = ", required"
				}
				fmt.Fprintf(&b, "- `--%s` (%s%s)", field.Name, field.Type, required)
				if field.Description != "" {
					fmt.Fprintf(&b, ": %s", field.Description)
				}
				b.WriteByte('\n')
			}
		}
		b.WriteByte('\n')
	}
	b.WriteString("<!-- ctxl:generated:end -->")
	return b.String()
}

func writeEntityCommands(b *strings.Builder, cliName string, e schema.Entity) {
	cmd := e.Command.Name
	b.WriteString("```bash\n")
	switch {
	case e.Format == schema.FormatSymlink:
		fmt.Fprintf(b, "%s %s show\n%s %s write\n", cliName, cmd, cliName, cmd)
	case e.Kind == schema.KindSingular:
		fmt.Fprintf(b, "%s %s show\n%s %s write", cliName, cmd, cliName, cmd)
		for _, field := range e.Fields {
			if field.Required {
				fmt.Fprintf(b, " --%s VALUE", field.Name)
			}
		}
		b.WriteByte('\n')
	case e.Format == schema.FormatNDJSON:
		fmt.Fprintf(b, "%s %s append", cliName, cmd)
		for _, field := range e.Fields {
			if field.Required && field.Name != "id" && field.Name != "ts" {
				fmt.Fprintf(b, " --%s VALUE", field.Name)
			}
		}
		fmt.Fprintf(b, "\n%s %s list\n%s %s list --full\n%s %s get --id ID\n", cliName, cmd, cliName, cmd, cliName, cmd)
	default:
		fmt.Fprintf(b, "%s %s create --id ID\n%s %s list\n%s %s get --id ID\n%s %s update --id ID\n%s %s delete --id ID\n", cliName, cmd, cliName, cmd, cliName, cmd, cliName, cmd, cliName, cmd)
	}
	b.WriteString("```\n")
}

func readDirectory(root string) ([]skillbundle.Entry, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("symlink roots are not supported")
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("is not a directory")
	}
	var entries []skillbundle.Entry
	seen := map[string]string{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are not supported", path)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("%s: unsupported filesystem entry", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		folded := strings.ToLower(rel)
		if prior, ok := seen[folded]; ok && prior != rel {
			return fmt.Errorf("%s: path collides with %s on case-insensitive filesystems", rel, prior)
		}
		seen[folded] = rel
		item := skillbundle.Entry{Path: rel, Mode: uint32(info.Mode().Perm()), Directory: info.IsDir()}
		if !info.IsDir() {
			item.Data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		}
		entries = append(entries, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func parseSkillMarkdown(raw []byte) (skillFrontmatter, string, error) {
	head, body, err := splitFrontmatter(raw)
	if err != nil {
		return skillFrontmatter{}, "", err
	}
	var meta skillFrontmatter
	if err := yaml.Unmarshal(head, &meta); err != nil {
		return skillFrontmatter{}, "", fmt.Errorf("frontmatter yaml: %w", err)
	}
	return meta, strings.TrimSpace(string(body)), nil
}

func renderSkillMarkdown(meta skillFrontmatter, body string) ([]byte, error) {
	head, err := yaml.Marshal(meta)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(head)
	out.WriteString("---\n")
	if strings.TrimSpace(body) != "" {
		out.WriteByte('\n')
		out.WriteString(strings.TrimSpace(body))
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

func splitFrontmatter(raw []byte) ([]byte, []byte, error) {
	normalized := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return nil, nil, fmt.Errorf("SKILL.md must start with YAML frontmatter")
	}
	rest := normalized[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		if strings.HasSuffix(rest, "\n---") {
			end = len(rest) - len("\n---")
		} else {
			return nil, nil, fmt.Errorf("SKILL.md has unclosed YAML frontmatter")
		}
	}
	head := []byte(rest[:end])
	bodyStart := end + len("\n---")
	body := []byte(strings.TrimPrefix(rest[bodyStart:], "\n"))
	return head, body, nil
}

func validateFrontmatter(meta skillFrontmatter) error {
	switch {
	case len(meta.Name) > 64 || !skillNamePattern.MatchString(meta.Name):
		return fmt.Errorf("frontmatter name %q is not a valid Agent Skill name", meta.Name)
	case meta.Description == "" || len(meta.Description) > 1024:
		return fmt.Errorf("frontmatter description is required and at most 1024 characters")
	case len(meta.Compatibility) > 500:
		return fmt.Errorf("frontmatter compatibility is at most 500 characters")
	}
	return nil
}

func entryData(entries []skillbundle.Entry, name string) ([]byte, error) {
	for _, entry := range entries {
		if entry.Path == name && !entry.Directory {
			return append([]byte(nil), entry.Data...), nil
		}
	}
	return nil, fmt.Errorf("missing %s", name)
}

func replaceEntry(entries []skillbundle.Entry, replacement skillbundle.Entry) []skillbundle.Entry {
	out := make([]skillbundle.Entry, 0, len(entries)+1)
	replaced := false
	for _, entry := range entries {
		if entry.Path == replacement.Path {
			if !replaced {
				out = append(out, replacement)
				replaced = true
			}
			continue
		}
		out = append(out, entry)
	}
	if !replaced {
		out = append(out, replacement)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func joinMarkdown(parts ...string) string {
	nonempty := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			nonempty = append(nonempty, trimmed)
		}
	}
	return strings.Join(nonempty, "\n\n")
}
