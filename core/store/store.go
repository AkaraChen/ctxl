package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkaraChen/ctxl/core/schema"
)

type Store struct {
	schema schema.Schema
	scope  schema.Scope
	root   string
}

func Open(s schema.Schema, scope schema.Scope, projectRoot string) (Store, error) {
	if projectRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Store{}, err
		}
		projectRoot = wd
	}
	if scope != schema.ScopeProject && scope != schema.ScopeGlobal {
		return Store{}, fmt.Errorf("scope must be project or global")
	}
	return Store{schema: s, scope: scope, root: projectRoot}, nil
}

func (st Store) StoreDir() (string, error) {
	name := st.schema.Store.Name
	if st.scope == schema.ScopeGlobal {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "."+name), nil
	}
	return filepath.Join(st.root, "."+name), nil
}

func (st Store) EntityPath(e schema.Entity) (string, error) {
	if e.Scope != schema.ScopeBoth && e.Scope != st.scope {
		return "", fmt.Errorf("entity %q is not available in %s scope", e.Name, st.scope)
	}
	if st.scope == schema.ScopeGlobal {
		dir, err := st.StoreDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, e.Path), nil
	}
	if e.Location == schema.LocationRoot {
		return filepath.Join(st.root, e.Path), nil
	}
	dir, err := st.StoreDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, e.Path), nil
}

func (st Store) ensureParent(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}
