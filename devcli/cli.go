package devcli

import (
	"fmt"

	"github.com/AkaraChen/ctxl/core/codegen"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	root := &cobra.Command{
		Use:               "ctxl",
		Short:             "Generate schema-specialized context CLIs",
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.AddCommand(&cobra.Command{
		Use:   "generate <schema.json>",
		Short: "Generate the CLI declared by a ctxl schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := codegen.Generate(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "generated %s\n", output)
			return nil
		},
	})
	return root
}
