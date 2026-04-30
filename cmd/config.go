package cmd

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/suhai-art/hourly-cli/internal/config"
	"github.com/suhai-art/hourly-cli/internal/report"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configura valor/hora, moeda e meta de horas/dia",
	}

	cmd.AddCommand(newConfigSetCmd(), newConfigShowCmd())
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	var (
		currency   string
		dailyHours float64
	)

	cmd := &cobra.Command{
		Use:   "set <valor_por_hora>",
		Short: "Define o valor por hora e/ou meta de horas/dia",
		Long: `Define o valor cobrado por hora, a moeda e/ou a meta de horas trabalhadas por dia.

Exemplos:
  hourly config set 50
  hourly config set 75.50 --currency "USD"
  hourly config set 100 --currency "€"
  hourly config set 50 --daily-hours 8
  hourly config set --daily-hours 6`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				rate, err := strconv.ParseFloat(args[0], 64)
				if err != nil || rate <= 0 {
					return fmt.Errorf("valor inválido: %q (use um número positivo, ex: 50 ou 75.50)", args[0])
				}
				cfg.HourlyRate = rate
			}

			if currency != "" {
				cfg.Currency = currency
			}

			if cmd.Flags().Changed("daily-hours") {
				if dailyHours <= 0 {
					return fmt.Errorf("horas/dia inválido: deve ser um número positivo")
				}
				cfg.DailyHours = dailyHours
			}

			if len(args) == 0 && !cmd.Flags().Changed("daily-hours") && currency == "" {
				return fmt.Errorf("nenhuma configuração fornecida. Use um valor/hora, --daily-hours ou --currency")
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			printConfigSaved(cfg)
			return nil
		},
	}

	cmd.Flags().StringVarP(&currency, "currency", "c", "", "Símbolo da moeda (ex: R$, USD, €)")
	cmd.Flags().Float64Var(&dailyHours, "daily-hours", 0, "Meta de horas trabalhadas por dia (ex: 8, 6.5)")
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

			if !cfg.HasRate() && !cfg.HasDailyHours() {
				color.Yellow("Nenhuma configuração encontrada. Use: hourly config set")
				return nil
			}

			bold := color.New(color.Bold)
			muted := color.New(color.FgHiBlack)

			fmt.Println()
			if cfg.HasRate() {
				fmt.Printf("  Valor/hora:   %s\n", bold.Sprintf("%s %.2f", cfg.Currency, cfg.HourlyRate))
			}
			if cfg.HasDailyHours() {
				fmt.Printf("  Horas/dia:    %s\n", bold.Sprint(report.FormatDuration(cfg.DailyDuration())))
			}
			if cfg.UpdatedAt != "" {
				fmt.Printf("  Atualizado:   %s\n", muted.Sprint(cfg.UpdatedAt))
			}
			fmt.Println()
			return nil
		},
	}
}

func printConfigSaved(cfg *config.Config) {
	bold := color.New(color.Bold)

	fmt.Println()
	if cfg.HasRate() {
		color.Green("  ✓ Valor/hora:  %s", bold.Sprintf("%s %.2f", cfg.Currency, cfg.HourlyRate))
	}
	if cfg.HasDailyHours() {
		color.Green("  ✓ Horas/dia:   %s", bold.Sprintf("%.1fh", cfg.DailyHours))
	}
	fmt.Println()
}
