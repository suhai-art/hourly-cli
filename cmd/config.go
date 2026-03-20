package cmd

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/suhai-art/hourly-cli/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configura valor/hora e moeda",
	}

	cmd.AddCommand(newConfigSetCmd(), newConfigShowCmd())
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	var currency string

	cmd := &cobra.Command{
		Use:   "set <valor_por_hora>",
		Short: "Define o valor por hora",
		Long: `Define o valor cobrado por hora e a moeda.

Exemplos:
  hourly config set 50
  hourly config set 75.50 --currency "USD"
  hourly config set 100 --currency "€"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rate, err := strconv.ParseFloat(args[0], 64)
			if err != nil || rate <= 0 {
				return fmt.Errorf("valor inválido: %q (use um número positivo, ex: 50 ou 75.50)", args[0])
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.HourlyRate = rate
			if currency != "" {
				cfg.Currency = currency
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			color.Green("✓ Configurado: %s %.2f/hora", cfg.Currency, cfg.HourlyRate)
			return nil
		},
	}

	cmd.Flags().StringVarP(&currency, "currency", "c", "", "Símbolo da moeda (ex: R$, USD, €)")
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Exibe a configuração atual",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if !cfg.HasRate() {
				color.Yellow("Nenhum valor/hora configurado. Use: hourly config set <valor>")
				return nil
			}

			bold := color.New(color.Bold)
			muted := color.New(color.FgHiBlack)

			fmt.Printf("\n  Valor/hora:  %s\n", bold.Sprintf("%s %.2f", cfg.Currency, cfg.HourlyRate))
			if cfg.UpdatedAt != "" {
				fmt.Printf("  Atualizado:  %s\n", muted.Sprint(cfg.UpdatedAt))
			}
			fmt.Println()
			return nil
		},
	}
}
