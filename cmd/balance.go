package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/suhai-art/hourly-cli/internal/config"
	"github.com/suhai-art/hourly-cli/internal/report"
	"github.com/suhai-art/hourly-cli/internal/store"
	"github.com/suhai-art/hourly-cli/internal/workdays"
)

func newBalanceCmd() *cobra.Command {
	var month string

	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Balanço de horas do mês vs. dias úteis",
		Long: `Compara as horas trabalhadas no mês com a meta esperada,
considerando apenas os dias úteis (segunda a sexta).

Mostra:
  - Total de dias úteis no mês
  - Horas esperadas até hoje
  - Horas trabalhadas até hoje
  - Saldo (positivo = adiantado, negativo = devendo)
  - Projeção para o fim do mês

Exemplos:
  hourly balance
  hourly balance --month 2024-03`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if !cfg.HasDailyHours() {
				return fmt.Errorf(
					"nenhuma meta de horas/dia configurada. Use: hourly config set --daily-hours <horas>",
				)
			}

			s, err := store.Load()
			if err != nil {
				return err
			}

			ref, err := parseMonthRef(month)
			if err != nil {
				return err
			}

			printBalance(s, cfg, ref)
			return nil
		},
	}

	cmd.Flags().StringVarP(&month, "month", "m", "", "Mês de referência (YYYY-MM)")
	return cmd
}

func printBalance(s *store.Store, cfg *config.Config, ref time.Time) {
	now := time.Now()
	isCurrentMonth := ref.Year() == now.Year() && ref.Month() == now.Month()

	// Workday counts
	totalWorkdays := workdays.CountInMonth(ref)
	var workedWorkdays int
	if isCurrentMonth {
		workedWorkdays = workdays.CountUntilToday(now)
	} else {
		workedWorkdays = totalWorkdays
	}
	remainingWorkdays := totalWorkdays - workedWorkdays

	// Expected and actual hours
	dailyGoal := cfg.DailyDuration()
	expectedSoFar := time.Duration(workedWorkdays) * dailyGoal
	monthGoal := time.Duration(totalWorkdays) * dailyGoal

	entries := s.ByMonth(ref)
	workedSoFar := sumCompletedDuration(entries)

	balance := workedSoFar - expectedSoFar

	// Projection: worked + (remaining workdays × daily goal)
	projected := workedSoFar + time.Duration(remainingWorkdays)*dailyGoal

	// Styles
	headerStyle := color.New(color.FgCyan, color.Bold)
	muted := color.New(color.FgHiBlack)
	boldStyle := color.New(color.Bold)
	successBold := color.New(color.FgGreen, color.Bold)
	warnBold := color.New(color.FgYellow, color.Bold)
	redBold := color.New(color.FgRed, color.Bold)
	earnStyle := color.New(color.FgMagenta, color.Bold)

	sep := muted.Sprint("  " + dashes(54))

	fmt.Println()
	headerStyle.Printf("  Balanço de horas — %s\n", ref.Format("January 2006"))
	fmt.Println(sep)

	// Workdays summary
	fmt.Printf("  Dias úteis no mês:     %s\n",
		boldStyle.Sprintf("%d dias", totalWorkdays))
	fmt.Printf("  Dias úteis até hoje:   %s\n",
		boldStyle.Sprintf("%d dias", workedWorkdays))
	fmt.Printf("  Dias úteis restantes:  %s\n",
		boldStyle.Sprintf("%d dias", remainingWorkdays))

	fmt.Println(sep)

	// Hours summary
	fmt.Printf("  Meta do mês:           %s\n",
		boldStyle.Sprint(report.FormatDuration(monthGoal)))
	fmt.Printf("  Esperado até hoje:     %s\n",
		boldStyle.Sprint(report.FormatDuration(expectedSoFar)))
	fmt.Printf("  Trabalhado até hoje:   %s\n",
		successBold.Sprint(report.FormatDuration(workedSoFar)))

	fmt.Println(sep)

	// Balance
	fmt.Printf("  Saldo atual:           ")
	switch {
	case balance > 0:
		successBold.Printf("+%s adiantado\n", report.FormatDuration(balance))
	case balance < 0:
		redBold.Printf("-%s devendo\n", report.FormatDuration(-balance))
	default:
		boldStyle.Println("em dia ✓")
	}

	fmt.Println(sep)

	// Projection
	fmt.Printf("  Projeção ao fim do mês: %s", boldStyle.Sprint(report.FormatDuration(projected)))
	projBalance := projected - monthGoal
	switch {
	case projBalance > 0:
		successBold.Printf("  (+%s)\n", report.FormatDuration(projBalance))
	case projBalance < 0:
		warnBold.Printf("  (-%s)\n", report.FormatDuration(-projBalance))
	default:
		fmt.Println()
	}

	// Earnings
	if cfg.HasRate() {
		fmt.Println(sep)
		fmt.Printf("  Ganho até hoje:         %s\n",
			earnStyle.Sprint(cfg.Earn(workedSoFar.Hours())))
		fmt.Printf("  Ganho projetado:        %s\n",
			earnStyle.Sprint(cfg.Earn(projected.Hours())))
	}

	fmt.Println()
}

// sumCompletedDuration sums the duration of all completed (non-open) entries.
func sumCompletedDuration(entries []store.Entry) time.Duration {
	var total time.Duration
	for _, e := range entries {
		total += e.Duration()
	}
	return total
}
