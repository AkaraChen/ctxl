package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkaraChen/ctxl/core/schema"
)

type Scope schema.Scope

const (
	ScopeProject = Scope(schema.ScopeProject)
	ScopeGlobal  = Scope(schema.ScopeGlobal)
)

type Store struct {
	Schema schema.Schema
	Scope  Scope
	Root   string
}

func Open(s schema.Schema, scope Scope, projectRoot string) (Store, error) {
	if projectRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Store{}, err
		}
		projectRoot = wd
	}
	if scope == "" {
		scope = ScopeProject
	}
	return Store{Schema: s, Scope: scope, Root: projectRoot}, nil
}

func (st Store) StoreDir() (string, error) {
	name := st.Schema.Name
	if st.Scope == ScopeGlobal {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "."+name), nil
	}
	return filepath.Join(st.Root, "."+name), nil
}

func (st Store) EntityPath(e schema.Entity) (string, error) {
	loc := e.ResolvedLocation()
	if st.Scope == ScopeGlobal {
		dir, err := st.StoreDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, e.Path), nil
	}
	if loc == schema.LocationRoot {
		return filepath.Join(st.Root, e.Path), nil
	}
	dir, err := st.StoreDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, e.Path), nil
}

func (st Store) EnsureParent(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func (st Store) Entity(name string) (schema.Entity, error) {
	e, err := st.Schema.Entity(name)
	if err != nil {
		return e, err
	}
	allowed := e.ResolvedScope()
	if allowed != schema.ScopeBoth && string(allowed) != string(st.Scope) {
		return e, fmt.Errorf("entity %q is not available in %s scope", name, st.Scope)
	}
	return e, nil
}
