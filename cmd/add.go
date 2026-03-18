package cmd

import (
	"fmt"

	"github.com/suhai-art/hourly-cli/internal/report"
	"github.com/suhai-art/hourly-cli/internal/store"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var note string

	cmd := &cobra.Command{
		Use:   "add <entrada> [saída]",
		Short: "Registra entrada (e opcionalmente saída)",
		Long: `Registra um período de trabalho.

Exemplos:
  github.com/suhai-art/hourly-cli add 09:00
  github.com/suhai-art/hourly-cli add 09:00 18:00
  github.com/suhai-art/hourly-cli add "2024-03-15 08:30" "2024-03-15 17:45" --note "reunião manhã"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Load()
			if err != nil {
				return err
			}

			in, err := store.ParseTime(args[0])
			if err != nil {
				return fmt.Errorf("horário de entrada inválido: %w", err)
			}

			entry := store.Entry{
				ID:   store.NewID(),
				In:   in,
				Note: note,
			}

			if len(args) == 2 {
				out, err := store.ParseTime(args[1])
				if err != nil {
					return fmt.Errorf("horário de saída inválido: %w", err)
				}
				if out.Before(in) {
					return fmt.Errorf("saída não pode ser antes da entrada")
				}
				entry.Out = &out
				color.Green("✓ Registrado: %s → %s  (%s)",
					in.Format("02/01 15:04"),
					out.Format("15:04"),
					report.FormatDuration(entry.Duration()),
				)
			} else {
				color.Yellow("✓ Entrada registrada: %s  (em aberto)", in.Format("02/01 15:04"))
			}

			s.Add(entry)
			return s.Save()
		},
	}

	cmd.Flags().StringVarP(&note, "note", "n", "", "Nota opcional para o registro")
	return cmd
}
