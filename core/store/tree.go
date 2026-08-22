package store

import "github.com/AkaraChen/ctxl/core/schema"

// TreeNode is one level-2 logical entity. Absent files still appear.
type TreeNode struct {
	Entity string `json:"entity"`
	Kind   string `json:"kind"`
	Format string `json:"format"`
	Path   string `json:"path"`
}

// Tree projects the selected entities as a two-level store → entity list.
func Tree(entities []schema.Entity) []TreeNode {
	out := make([]TreeNode, 0, len(entities))
	for _, e := range entities {
		out = append(out, TreeNode{
			Entity: e.Name,
			Kind:   string(e.Kind),
			Format: string(e.Format),
			Path:   e.Path,
		})
	}
	return out
}
