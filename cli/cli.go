package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/skillbundle"
	"github.com/AkaraChen/ctxl/core/store"
	"github.com/spf13/cobra"
)

// Options provide a normalized schema and its prepared Skill bundle.
type Options struct {
	Schema schema.Schema
	Skills skillbundle.Bundle
}

// New builds a schema-specialized cobra command tree from opts.Schema.
func New(opts Options) *cobra.Command {
	s := opts.Schema
	var scope string
	root := &cobra.Command{
		Use:           s.CLI.Name,
		Short:         shortOf(s),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch scope {
			case string(schema.ScopeProject), string(schema.ScopeGlobal):
			default:
				return fmt.Errorf("scope must be project or global")
			}
			return nil
		},
	}
	root.PersistentFlags().StringVar(&scope, "scope", "project", "project|global")
	open := func() (store.Store, error) {
		sc := schema.ScopeProject
		if scope == "global" {
			sc = schema.ScopeGlobal
		}
		return store.Open(s, sc, "")
	}
	for _, ent := range s.Entities {
		root.AddCommand(entityCommand(ent, open))
	}
	root.AddCommand(initCommand(open))
	root.AddCommand(skillsCommand(opts.Skills))
	return root
}

func entityCommand(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	cmd := &cobra.Command{Use: e.Command.Name, Short: entityShort(e)}
	switch {
	case e.Kind == schema.KindSingular:
		cmd.AddCommand(writeCmd(e, open), showCmd(e, open))
	case e.Format == schema.FormatNDJSON:
		cmd.AddCommand(appendCmd(e, open), listNDJSONCmd(e, open), getNDJSONCmd(e, open))
	default:
		cmd.AddCommand(mdWrite(e, open, "create", false), mdWrite(e, open, "update", true), listMD(e, open), getMD(e, open), deleteMD(e, open))
	}
	return cmd
}

func entityShort(e schema.Entity) string {
	if e.Description != "" {
		return e.Description
	}
	return string(e.Kind) + " " + string(e.Format)
}

func writeCmd(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	values := map[string]*string{}
	var body string
	cmd := &cobra.Command{Use: "write", Short: writeShort(e), RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		fields := map[string]string{}
		for _, f := range e.Fields {
			if p := values[f.Name]; cmd.Flags().Changed(f.Name) || *p != "" {
				fields[f.Name] = *p
			}
		}
		return st.WriteSingular(e, store.Record{Fields: fields, Body: body})
	}}
	for _, f := range e.Fields {
		v := ""
		values[f.Name] = &v
		cmd.Flags().StringVar(values[f.Name], f.Name, "", f.Description)
		if f.Required {
			_ = cmd.MarkFlagRequired(f.Name)
		}
	}
	cmd.Flags().StringVar(&body, "body", "", "text after frontmatter")
	return cmd
}

func showCmd(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	return &cobra.Command{Use: "show", Short: "Print current " + e.Name, RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		rec, err := st.ReadSingular(e)
		if err != nil {
			return err
		}
		return printJSON(cmd, rec.Public())
	}}
}

func appendCmd(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	values := map[string]*string{}
	cmd := &cobra.Command{Use: "append", Short: "Append one " + e.Name, RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		fields := map[string]any{}
		for _, f := range e.Fields {
			p := values[f.Name]
			if *p == "" {
				continue
			}
			if f.Type == schema.TypeObject {
				var obj any
				if err := json.Unmarshal([]byte(*p), &obj); err != nil {
					return fmt.Errorf("%s must be JSON: %w", f.Name, err)
				}
				fields[f.Name] = obj
			} else if f.Type == schema.TypeInt {
				n, err := strconv.Atoi(*p)
				if err != nil {
					return fmt.Errorf("%s must be int: %w", f.Name, err)
				}
				fields[f.Name] = n
			} else {
				fields[f.Name] = *p
			}
		}
		row, err := st.AppendNDJSON(e, fields)
		if err != nil {
			return err
		}
		return printJSON(cmd, row)
	}}
	for _, f := range e.Fields {
		if f.Name == e.ID || f.Name == "ts" {
			continue
		}
		v := ""
		values[f.Name] = &v
		cmd.Flags().StringVar(values[f.Name], f.Name, "", f.Description)
		if f.Required {
			_ = cmd.MarkFlagRequired(f.Name)
		}
	}
	return cmd
}

func listNDJSONCmd(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	var full bool
	cmd := &cobra.Command{Use: "list", Short: "List " + e.Name + " rows; default is fixed fields only", RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		rows, err := st.ListNDJSON(e)
		if err != nil {
			return err
		}
		if rows == nil {
			rows = []map[string]any{}
		}
		if !full {
			rows = store.FixedRows(e, rows)
		}
		return printJSON(cmd, rows)
	}}
	cmd.Flags().BoolVar(&full, "full", false, "include object-typed fields")
	return cmd
}

func getNDJSONCmd(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	var id string
	cmd := &cobra.Command{Use: "get", Short: "Get one " + e.Name + " by id", RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		row, err := st.GetNDJSON(e, id)
		if err != nil {
			return err
		}
		return printJSON(cmd, row)
	}}
	cmd.Flags().StringVar(&id, "id", "", "row id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func mdWrite(e schema.Entity, open func() (store.Store, error), use string, mustExist bool) *cobra.Command {
	values := map[string]*string{}
	var id, body string
	cmd := &cobra.Command{Use: use, Short: use + " one " + e.Name, RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		fields := map[string]string{}
		if mustExist {
			rec, err := st.ReadMarkdownItem(e, id)
			if err != nil {
				return err
			}
			fields = rec.Fields
			if !cmd.Flags().Changed("body") {
				body = rec.Body
			}
		} else {
			ids, err := st.ListMarkdownItems(e)
			if err != nil {
				return err
			}
			for _, existing := range ids {
				if existing == id {
					return fmt.Errorf("%s %s already exists", e.Name, id)
				}
			}
		}
		for _, f := range e.Fields {
			if p := values[f.Name]; cmd.Flags().Changed(f.Name) || *p != "" {
				fields[f.Name] = *p
			}
		}
		return st.WriteMarkdownItem(e, id, store.Record{Fields: fields, Body: body})
	}}
	cmd.Flags().StringVar(&id, "id", "", "item id")
	_ = cmd.MarkFlagRequired("id")
	for _, f := range e.Fields {
		if f.Name == e.ID {
			continue
		}
		v := ""
		values[f.Name] = &v
		cmd.Flags().StringVar(values[f.Name], f.Name, "", f.Description)
		if !mustExist && f.Required {
			_ = cmd.MarkFlagRequired(f.Name)
		}
	}
	cmd.Flags().StringVar(&body, "body", "", "markdown body")
	return cmd
}

func listMD(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List " + e.Name + " ids", RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		ids, err := st.ListMarkdownItems(e)
		if err != nil {
			return err
		}
		if ids == nil {
			ids = []string{}
		}
		return printJSON(cmd, ids)
	}}
}

func getMD(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	var id string
	cmd := &cobra.Command{Use: "get", Short: "Get one " + e.Name, RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		rec, err := st.ReadMarkdownItem(e, id)
		if err != nil {
			return err
		}
		return printJSON(cmd, rec)
	}}
	cmd.Flags().StringVar(&id, "id", "", "item id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func deleteMD(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	var id string
	cmd := &cobra.Command{Use: "delete", Short: "Delete one " + e.Name, RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		return st.DeleteMarkdownItem(e, id)
	}}
	cmd.Flags().StringVar(&id, "id", "", "item id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func skillsCommand(bundle skillbundle.Bundle) *cobra.Command {
	cmd := &cobra.Command{Use: "skills", Short: "Access bundled Agent Skills"}
	if _, ok := bundle.SoleName(); !ok {
		cmd.AddCommand(&cobra.Command{Use: "list", Short: "List bundled skills", RunE: func(cmd *cobra.Command, args []string) error {
			return printJSON(cmd, bundle.Names())
		}})
	}
	cmd.AddCommand(&cobra.Command{Use: skillNameUse("get", bundle), Short: "Print bundled Skill instructions", Args: skillNameArgs(bundle), RunE: func(cmd *cobra.Command, args []string) error {
		name, err := resolveSkillName(bundle, args)
		if err != nil {
			return err
		}
		body, err := bundle.Markdown(name)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(body))
		return err
	}}, &cobra.Command{Use: skillNameUse("path", bundle), Short: "Materialize a bundled skill directory", Args: skillNameArgs(bundle), RunE: func(cmd *cobra.Command, args []string) error {
		name, err := resolveSkillName(bundle, args)
		if err != nil {
			return err
		}
		path, err := bundle.Materialize(name)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
		return err
	}})
	return cmd
}

func skillNameUse(verb string, bundle skillbundle.Bundle) string {
	if _, ok := bundle.SoleName(); ok {
		return verb + " [name]"
	}
	return verb + " <name>"
}

func skillNameArgs(bundle skillbundle.Bundle) cobra.PositionalArgs {
	if _, ok := bundle.SoleName(); ok {
		return cobra.MaximumNArgs(1)
	}
	return cobra.ExactArgs(1)
}

func resolveSkillName(bundle skillbundle.Bundle, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if name, ok := bundle.SoleName(); ok {
		return name, nil
	}
	return "", fmt.Errorf("skill name required")
}

func printJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func shortOf(s schema.Schema) string {
	if s.Description != "" {
		return s.Description
	}
	return "Schema-specialized context CLI."
}

func writeShort(e schema.Entity) string {
	if e.Format == schema.FormatSymlink {
		return "Create symlink " + e.Name
	}
	if e.Write == schema.WriteSection {
		return "Write section in " + e.Name
	}
	return "Overwrite current " + e.Name
}

func initCommand(open func() (store.Store, error)) *cobra.Command {
	var force bool
	cmd := &cobra.Command{Use: "init", Short: "Create every entity path declared in the schema", RunE: func(cmd *cobra.Command, args []string) error {
		st, err := open()
		if err != nil {
			return err
		}
		rows, err := st.Init(force)
		if err != nil {
			return err
		}
		return printJSON(cmd, rows)
	}}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	return cmd
}
