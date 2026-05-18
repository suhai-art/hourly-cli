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
		Short: "Configura valor/hora, moeda e metas de horas/dia",
	}

	cmd.AddCommand(newConfigSetCmd(), newConfigShowCmd())
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	var (
		currency               string
		dailyHours             float64
		balanceDailyHours      float64
		previewDailyHours      float64
		resetBalanceDailyHours bool
		resetPreviewDailyHours bool
	)

	cmd := &cobra.Command{
		Use:   "set <valor_por_hora>",
		Short: "Define valor por hora, moeda e/ou metas de horas/dia",
		Long: `Define o valor cobrado por hora, a moeda e/ou as metas de horas trabalhadas por dia.

As flags --balance-daily-hours e --preview-daily-hours permitem metas independentes
para cada comando. Quando não definidas, ambos usam --daily-hours como fallback.

Para remover uma meta específica e voltar ao comportamento padrão (daily-hours):
  hourly config set --reset-balance-daily-hours
  hourly config set --reset-preview-daily-hours

Exemplos:
  hourly config set 50
  hourly config set 75.50 --currency "USD"
  hourly config set --daily-hours 8
  hourly config set --balance-daily-hours 8 --preview-daily-hours 6
  hourly config set --reset-balance-daily-hours`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := applyRateArg(cfg, args); err != nil {
				return err
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

			if cmd.Flags().Changed("balance-daily-hours") {
				if balanceDailyHours <= 0 {
					return fmt.Errorf("balance-daily-hours inválido: deve ser um número positivo")
				}
				cfg.BalanceDailyHours = balanceDailyHours
			}

			if cmd.Flags().Changed("preview-daily-hours") {
				if previewDailyHours <= 0 {
					return fmt.Errorf("preview-daily-hours inválido: deve ser um número positivo")
				}
				cfg.PreviewDailyHours = previewDailyHours
			}

			if resetBalanceDailyHours {
				cfg.BalanceDailyHours = 0
			}

			if resetPreviewDailyHours {
				cfg.PreviewDailyHours = 0
			}

			if !anyFlagChanged(cmd, args) {
				return fmt.Errorf(
					"nenhuma configuração fornecida. Use um valor/hora, --daily-hours, " +
						"--balance-daily-hours, --preview-daily-hours, --currency, " +
						"--reset-balance-daily-hours ou --reset-preview-daily-hours",
				)
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			printConfigSaved(cfg)
			return nil
		},
	}

	cmd.Flags().StringVarP(&currency, "currency", "c", "", "Símbolo da moeda (ex: R$, USD, €)")
	cmd.Flags().Float64Var(&dailyHours, "daily-hours", 0,
		"Meta de horas/dia padrão — usada por balance e preview quando não há meta específica")
	cmd.Flags().Float64Var(&balanceDailyHours, "balance-daily-hours", 0,
		"Meta de horas/dia exclusiva para o comando balance (sobrescreve --daily-hours)")
	cmd.Flags().Float64Var(&previewDailyHours, "preview-daily-hours", 0,
		"Meta de horas/dia exclusiva para o comando preview (sobrescreve --daily-hours)")
	cmd.Flags().BoolVar(&resetBalanceDailyHours, "reset-balance-daily-hours", false,
		"Remove a meta específica do balance, voltando a usar --daily-hours")
	cmd.Flags().BoolVar(&resetPreviewDailyHours, "reset-preview-daily-hours", false,
		"Remove a meta específica do preview, voltando a usar --daily-hours")
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

			if !cfg.HasRate() && !cfg.HasDailyHours() && !cfg.HasBalanceDailyHours() {
				color.Yellow("Nenhuma configuração encontrada. Use: hourly config set")
				return nil
			}

			bold := color.New(color.Bold)
			muted := color.New(color.FgHiBlack)

			fmt.Println()
			if cfg.HasRate() {
				fmt.Printf("  Valor/hora:            %s\n",
					bold.Sprintf("%s %.2f", cfg.Currency, cfg.HourlyRate))
			}
			if cfg.HasDailyHours() {
				fmt.Printf("  Horas/dia (padrão):    %s\n",
					bold.Sprint(report.FormatDuration(cfg.DailyDuration())))
			}
			if cfg.BalanceDailyHours > 0 {
				fmt.Printf("  Horas/dia (balance):   %s\n",
					bold.Sprint(report.FormatDuration(cfg.BalanceDailyDuration())))
			}
			if cfg.PreviewDailyHours > 0 {
				fmt.Printf("  Horas/dia (preview):   %s\n",
					bold.Sprint(report.FormatDuration(cfg.PreviewDailyDuration())))
			}
			if cfg.UpdatedAt != "" {
				fmt.Printf("  Atualizado:            %s\n", muted.Sprint(cfg.UpdatedAt))
			}
			fmt.Println()
			return nil
		},
	}
}

// applyRateArg parses and applies the optional positional hourly-rate argument.
func applyRateArg(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return nil
	}
	rate, err := strconv.ParseFloat(args[0], 64)
	if err != nil || rate <= 0 {
		return fmt.Errorf("valor inválido: %q (use um número positivo, ex: 50 ou 75.50)", args[0])
	}
	cfg.HourlyRate = rate
	return nil
}

func anyFlagChanged(cmd *cobra.Command, args []string) bool {
	if len(args) > 0 {
		return true
	}
	for _, name := range []string{
		"daily-hours", "balance-daily-hours", "preview-daily-hours", "currency",
		"reset-balance-daily-hours", "reset-preview-daily-hours",
	} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func printConfigSaved(cfg *config.Config) {
	bold := color.New(color.Bold)

	fmt.Println()
	if cfg.HasRate() {
		color.Green("  ✓ Valor/hora:            %s",
			bold.Sprintf("%s %.2f", cfg.Currency, cfg.HourlyRate))
	}
	if cfg.HasDailyHours() {
		color.Green("  ✓ Horas/dia (padrão):    %s",
			bold.Sprintf("%.1fh", cfg.DailyHours))
	}
	if cfg.BalanceDailyHours > 0 {
		color.Green("  ✓ Horas/dia (balance):   %s",
			bold.Sprintf("%.1fh", cfg.BalanceDailyHours))
	}
	if cfg.PreviewDailyHours > 0 {
		color.Green("  ✓ Horas/dia (preview):   %s",
			bold.Sprintf("%.1fh", cfg.PreviewDailyHours))
	}
	fmt.Println()
}
