package skillbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Path      string
	Mode      uint32
	Directory bool
	Data      []byte
}

type Skill struct {
	Name    string
	Entries []Entry
}

type Bundle struct {
	Skills []Skill
}

func (b Bundle) Names() []string {
	names := make([]string, 0, len(b.Skills))
	for _, skill := range b.Skills {
		names = append(names, skill.Name)
	}
	return names
}

// SoleName returns the only Skill name when the bundle contains exactly one Skill.
func (b Bundle) SoleName() (string, bool) {
	if len(b.Skills) != 1 {
		return "", false
	}
	return b.Skills[0].Name, true
}

func (b Bundle) skill(name string) (Skill, error) {
	for _, skill := range b.Skills {
		if skill.Name == name {
			return skill, nil
		}
	}
	return Skill{}, fmt.Errorf("unknown skill %q", name)
}

func (b Bundle) Markdown(name string) ([]byte, error) {
	skill, err := b.skill(name)
	if err != nil {
		return nil, err
	}
	for _, entry := range skill.Entries {
		if entry.Path == "SKILL.md" && !entry.Directory {
			return append([]byte(nil), entry.Data...), nil
		}
	}
	return nil, fmt.Errorf("skill %q has no SKILL.md", name)
}

func (b Bundle) Materialize(name string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return b.materializeAt(name, filepath.Join(cache, "ctxl", "skills"))
}

func (b Bundle) materializeAt(name, root string) (string, error) {
	skill, err := b.skill(name)
	if err != nil {
		return "", err
	}
	cleanName, err := cleanEntryPath(skill.Name)
	if err != nil || cleanName != skill.Name || strings.Contains(skill.Name, "/") || strings.Contains(skill.Name, `\`) {
		return "", fmt.Errorf("unsafe skill name %q", skill.Name)
	}
	digest := skill.digest()
	parent := filepath.Join(root, digest)
	target := filepath.Join(parent, skill.Name)
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return target, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(parent, "."+skill.Name+"-tmp-")
	if err != nil {
		return "", err
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.RemoveAll(tmp)
		}
	}()
	for _, entry := range sortedEntries(skill.Entries) {
		rel, err := cleanEntryPath(entry.Path)
		if err != nil {
			return "", fmt.Errorf("skill %q: %w", skill.Name, err)
		}
		path := filepath.Join(tmp, filepath.FromSlash(rel))
		if entry.Directory {
			if err := os.MkdirAll(path, entry.fileMode(0o755)); err != nil {
				return "", err
			}
			if err := os.Chmod(path, entry.fileMode(0o755)); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, entry.Data, entry.fileMode(0o644)); err != nil {
			return "", err
		}
		if err := os.Chmod(path, entry.fileMode(0o644)); err != nil {
			return "", err
		}
	}
	if err := os.Rename(tmp, target); err != nil {
		if info, statErr := os.Stat(target); statErr == nil && info.IsDir() {
			return target, nil
		}
		return "", err
	}
	keep = true
	return target, nil
}

func (s Skill) digest() string {
	h := sha256.New()
	_, _ = h.Write([]byte(s.Name))
	_, _ = h.Write([]byte{0})
	for _, entry := range sortedEntries(s.Entries) {
		_, _ = h.Write([]byte(entry.Path))
		_, _ = h.Write([]byte{0})
		_, _ = fmt.Fprintf(h, "%d:%t", entry.Mode, entry.Directory)
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(entry.Data)
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func sortedEntries(entries []Entry) []Entry {
	out := append([]Entry(nil), entries...)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func cleanEntryPath(path string) (string, error) {
	path = filepath.ToSlash(path)
	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || filepath.IsAbs(path) {
		return "", fmt.Errorf("unsafe bundled path %q", path)
	}
	return clean, nil
}

func (e Entry) fileMode(fallback os.FileMode) os.FileMode {
	mode := os.FileMode(e.Mode).Perm()
	if mode == 0 {
		return fallback
	}
	return mode
}
