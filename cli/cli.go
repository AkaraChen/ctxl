package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/skillsgen"
	"github.com/AkaraChen/ctxl/core/store"
	"github.com/spf13/cobra"
)

// Options configure a branded CLI built from a schema.
type Options struct {
	Name   string
	Schema schema.Schema
}

// Execute builds the command tree and runs it.
func Execute(opts Options) error { return New(opts).Execute() }

// New builds a cobra command tree from opts.Schema.
// If --schema is already on the process args and loads, entity commands
// come from that file so a generic binary can host any schema.
func New(opts Options) *cobra.Command {
	s := opts.Schema
	if path := peekFlag(os.Args[1:], "schema"); path != "" {
		if loaded, err := schema.LoadFile(path); err == nil {
			s = loaded
		}
	}
	bin := opts.Name
	if bin == "" {
		bin = s.Name
	}
	var schemaPath, scope string
	root := &cobra.Command{
		Use:           bin,
		Short:         shortOf(s),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if schemaPath != "" {
				loaded, err := schema.LoadFile(schemaPath)
				if err != nil {
					return err
				}
				s = loaded
			}
			switch scope {
			case "", "project", "global":
			default:
				return fmt.Errorf("scope must be project or global")
			}
			return nil
		},
	}
	root.PersistentFlags().StringVar(&schemaPath, "schema", "", "JSON schema file")
	root.PersistentFlags().StringVar(&scope, "scope", "project", "project|global")
	open := func() (store.Store, error) {
		sc := store.ScopeProject
		if scope == "global" {
			sc = store.ScopeGlobal
		}
		return store.Open(s, sc, "")
	}
	for _, ent := range s.Entities {
		root.AddCommand(entityCommand(ent, open))
	}
	root.AddCommand(initCommand(open))
	root.AddCommand(skillsCommand(func() schema.Schema { return s }))
	root.AddCommand(schemaValidate(func() schema.Schema { return s }))
	return root
}

func entityCommand(e schema.Entity, open func() (store.Store, error)) *cobra.Command {
	cmd := &cobra.Command{Use: e.Name, Short: entityShort(e)}
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
			if p := values[f.Name]; p != nil && (cmd.Flags().Changed(f.Name) || *p != "") {
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
		return printJSON(rec.Public())
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
			if p == nil || *p == "" {
				continue
			}
			if f.Type == schema.TypeObject {
				var obj any
				if err := json.Unmarshal([]byte(*p), &obj); err != nil {
					return fmt.Errorf("%s must be JSON: %w", f.Name, err)
				}
				fields[f.Name] = obj
			} else if f.Type == schema.TypeInt {
				var n int
				if _, err := fmt.Sscan(*p, &n); err != nil {
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
		return printJSON(row)
	}}
	for _, f := range e.Fields {
		if f.Name == "id" || f.Name == "ts" {
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
		return printJSON(rows)
	}}
	cmd.Flags().BoolVar(&full, "full", false, "include object fields such as custom_data")
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
		return printJSON(row)
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
			if fields == nil {
				fields = map[string]string{}
			}
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
			if p := values[f.Name]; p != nil && (cmd.Flags().Changed(f.Name) || *p != "") {
				fields[f.Name] = *p
			}
		}
		return st.WriteMarkdownItem(e, id, store.Record{Fields: fields, Body: body})
	}}
	cmd.Flags().StringVar(&id, "id", "", "item id")
	_ = cmd.MarkFlagRequired("id")
	for _, f := range e.Fields {
		if f.Name == e.IDField() {
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
		return printJSON(ids)
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
		return printJSON(rec)
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

func skillsCommand(current func() schema.Schema) *cobra.Command {
	cmd := &cobra.Command{Use: "skills", Short: "Serve generated skill markdown"}
	cmd.AddCommand(&cobra.Command{Use: "list", Short: "List generated skills", RunE: func(cmd *cobra.Command, args []string) error {
		all := skillsgen.All(current())
		names := make([]string, 0, len(all))
		for n := range all {
			names = append(names, n)
		}
		return printJSON(names)
	}}, &cobra.Command{Use: "get <name>", Short: "Print a generated skill", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		all := skillsgen.All(current())
		body, ok := all[args[0]]
		if !ok {
			return fmt.Errorf("unknown skill %q", args[0])
		}
		fmt.Print(body)
		return nil
	}})
	return cmd
}

func schemaValidate(current func() schema.Schema) *cobra.Command {
	cmd := &cobra.Command{Use: "schema", Short: "Schema utilities"}
	cmd.AddCommand(&cobra.Command{Use: "validate", Short: "Validate the active schema", RunE: func(cmd *cobra.Command, args []string) error {
		if err := current().Validate(); err != nil {
			return err
		}
		fmt.Println("ok")
		return nil
	}})
	return cmd
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func peekFlag(args []string, name string) string {
	prefix := "--" + name
	for i, a := range args {
		if a == prefix && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, prefix+"=") {
			return strings.TrimPrefix(a, prefix+"=")
		}
	}
	return ""
}

func shortOf(s schema.Schema) string {
	if s.Description != "" {
		return s.Description
	}
	return "Context layer. Pass --schema FILE."
}

func writeShort(e schema.Entity) string {
	if e.Format == schema.FormatSymlink {
		return "Create symlink " + e.Name
	}
	if e.ResolvedWrite() == schema.WriteSection {
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
		return printJSON(rows)
	}}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	return cmd
}
