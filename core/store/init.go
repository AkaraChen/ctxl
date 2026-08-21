package store

import (
	"fmt"
	"os"

	"github.com/AkaraChen/ctxl/core/schema"
)

type InitResult struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

func (st Store) Init(force bool) ([]InitResult, error) {
	var out []InitResult
	for _, e := range st.Schema.Entities {
		action, err := st.initEntity(e, force)
		if err != nil {
			return out, err
		}
		out = append(out, InitResult{Name: e.Name, Action: action})
	}
	return out, nil
}

func (st Store) initEntity(e schema.Entity, force bool) (string, error) {
	path, err := st.EntityPath(e)
	if err != nil {
		return "", err
	}
	exists, err := pathExists(path)
	if err != nil {
		return "", err
	}
	switch {
	case e.Format == schema.FormatSymlink:
		if exists && !force {
			return "skipped", nil
		}
		if exists && force {
			if err := os.Remove(path); err != nil {
				return "", err
			}
		}
		if err := st.WriteSymlink(e); err != nil {
			return "", err
		}
		return "created", nil
	case e.Kind == schema.KindSingular && e.ResolvedWrite() == schema.WriteSection:
		if exists && !force {
			return "skipped", nil
		}
		if err := st.WriteSingular(e, Record{Body: e.Body}); err != nil {
			return "", err
		}
		return "created", nil
	case e.Kind == schema.KindSingular:
		if exists && !force {
			return "skipped", nil
		}
		if err := st.WriteSingular(e, Record{}); err != nil {
			return "", err
		}
		return "created", nil
	case e.Format == schema.FormatNDJSON:
		if exists && !force {
			return "skipped", nil
		}
		if err := st.EnsureParent(path); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			return "", err
		}
		return "created", nil
	default:
		if exists && !force {
			return "skipped", nil
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", err
		}
		return "created", nil
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("%s: %w", path, err)
}
