package cmd

import (
	"fmt"

	"github.com/suhai-art/hourly-cli/internal/store"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Remove um registro pelo ID",
		Long: `Remove um registro de horas pelo ID.

O ID é exibido na coluna direita do comando list.

Exemplo:
  github.com/suhai-art/hourly-cli delete 20240315090000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Load()
			if err != nil {
				return err
			}

			if !s.Delete(args[0]) {
				return fmt.Errorf("registro %q não encontrado", args[0])
			}

			color.Red("✗ Registro %s removido.", args[0])
			return s.Save()
		},
	}
}
