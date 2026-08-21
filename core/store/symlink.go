package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkaraChen/ctxl/core/schema"
)

func (st Store) WriteSymlink(e schema.Entity) error {
	link, err := st.EntityPath(e)
	if err != nil {
		return err
	}
	targetAbs, err := st.symlinkTarget(e)
	if err != nil {
		return err
	}
	if _, err := os.Stat(targetAbs); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("symlink target %s does not exist", e.Target)
		}
		return err
	}
	if info, err := os.Lstat(link); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			cur, err := os.Readlink(link)
			if err != nil {
				return err
			}
			if sameLink(link, cur, targetAbs) {
				return nil
			}
		}
		return fmt.Errorf("%s exists and is not this symlink", e.Path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := st.EnsureParent(link); err != nil {
		return err
	}
	rel, err := filepath.Rel(filepath.Dir(link), targetAbs)
	if err != nil {
		rel = targetAbs
	}
	return os.Symlink(rel, link)
}

func (st Store) ReadSymlink(e schema.Entity) (map[string]any, error) {
	link, err := st.EntityPath(e)
	if err != nil {
		return nil, err
	}
	cur, err := os.Readlink(link)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s (treat as never written)", e.Path)
		}
		return nil, err
	}
	return map[string]any{"path": e.Path, "target": e.Target, "linked": cur}, nil
}

func (st Store) symlinkTarget(e schema.Entity) (string, error) {
	if filepath.IsAbs(e.Target) {
		return e.Target, nil
	}
	return filepath.Join(st.Root, e.Target), nil
}

func sameLink(linkPath, current, targetAbs string) bool {
	if current == targetAbs {
		return true
	}
	resolved := current
	if !filepath.IsAbs(current) {
		resolved = filepath.Join(filepath.Dir(linkPath), current)
	}
	a, err1 := filepath.Abs(resolved)
	b, err2 := filepath.Abs(targetAbs)
	return err1 == nil && err2 == nil && a == b
}
