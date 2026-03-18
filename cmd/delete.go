package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/suhai-art/hourly-cli/internal/report"
	"github.com/suhai-art/hourly-cli/internal/store"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Remove registros de horas",
		Long: `Remove registros de horas.

Sem argumentos: abre seletor interativo.
Com --all: apaga todos os registros (pede confirmação).

Exemplos:
  hourly delete               # seletor interativo
  hourly delete 20240315090000
  hourly delete --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Load()
			if err != nil {
				return err
			}

			// -- delete --all
			if all {
				if len(s.Entries) == 0 {
					color.Yellow("Nenhum registro para apagar.")
					return nil
				}
				fmt.Printf("Isso vai apagar %s registros. Confirma? (y/N) ",
					color.RedString("%d", len(s.Entries)))
				reader := bufio.NewReader(os.Stdin)
				resp, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(resp)) != "y" {
					color.Yellow("Cancelado.")
					return nil
				}
				s.DeleteAll()
				if err := s.Save(); err != nil {
					return err
				}
				color.Red("✗ Todos os registros foram removidos.")
				return nil
			}

			// -- delete <id>
			if len(args) == 1 {
				if !s.Delete(args[0]) {
					return fmt.Errorf("registro %q não encontrado", args[0])
				}
				color.Red("✗ Registro %s removido.", args[0])
				return s.Save()
			}

			// -- modo interativo
			if len(s.Entries) == 0 {
				color.Yellow("Nenhum registro encontrado.")
				return nil
			}

			// Monta as opções do seletor
			options := make([]string, len(s.Entries))
			for i, e := range s.Entries {
				label := e.In.Format("02/01/2006  15:04")
				if e.Out != nil {
					label += " → " + e.Out.Format("15:04")
					label += "  " + report.FormatDuration(e.Duration())
				} else {
					label += "  (em aberto)"
				}
				if e.Note != "" {
					label += "  · " + e.Note
				}
				options[i] = label
			}

			var selected []string
			prompt := &survey.MultiSelect{
				Message:  "Selecione os registros para remover (espaço = marcar, enter = confirmar):",
				Options:  options,
				PageSize: 15,
			}
			if err := survey.AskOne(prompt, &selected); err != nil {
				return err
			}

			if len(selected) == 0 {
				color.Yellow("Nenhum registro selecionado.")
				return nil
			}

			// Mapeia opção selecionada → ID
			selectedSet := map[string]bool{}
			for _, sel := range selected {
				selectedSet[sel] = true
			}
			var toDelete []string
			for i, opt := range options {
				if selectedSet[opt] {
					toDelete = append(toDelete, s.Entries[i].ID)
				}
			}

			for _, id := range toDelete {
				s.Delete(id)
				color.Red("✗ Registro %s removido.", id)
			}
			return s.Save()
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Remove todos os registros")
	return cmd
}
