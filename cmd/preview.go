package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/suhai-art/hourly-cli/internal/config"
	"github.com/suhai-art/hourly-cli/internal/report"
	"github.com/suhai-art/hourly-cli/internal/store"
)

func newPreviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview",
		Short: "Mostra o horário previsto de saída com base no registro em aberto",
		Long: `Calcula e exibe o horário de saída previsto para o dia.

Soma todas as horas já concluídas no dia ao tempo do registro em aberto,
e projeta quando você atingirá a meta de horas/dia configurada.

Exemplos:
  hourly preview`,
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

			openEntry := s.FindOpenEntry()
			if openEntry == nil {
				color.Yellow("Nenhum registro em aberto encontrado.")
				return nil
			}

			now := time.Now()
			completedToday := sumCompletedHoursToday(s, openEntry.ID, now)
			elapsed := now.Sub(openEntry.In)
			totalSoFar := completedToday + elapsed

			goal := cfg.DailyDuration()
			remaining := goal - completedToday

			if remaining <= 0 {
				printGoalAlreadyReached(completedToday, goal)
				return nil
			}

			exitTime := openEntry.In.Add(remaining)
			printPreview(openEntry, now, completedToday, totalSoFar, goal, remaining, exitTime, cfg)
			return nil
		},
	}
}

func sumCompletedHoursToday(s *store.Store, openID string, ref time.Time) time.Duration {
	var total time.Duration
	for _, e := range s.ByDate(ref) {
		if e.ID == openID || e.IsOpen() {
			continue
		}
		total += e.Duration()
	}
	return total
}

func printPreview(
	open *store.Entry,
	now time.Time,
	completedToday time.Duration,
	totalSoFar time.Duration,
	goal time.Duration,
	remaining time.Duration,
	exitTime time.Time,
	cfg *config.Config,
) {
	headerStyle := color.New(color.FgCyan, color.Bold)
	muted := color.New(color.FgHiBlack)
	successBold := color.New(color.FgGreen, color.Bold)
	warnStyle := color.New(color.FgYellow, color.Bold)
	earnStyle := color.New(color.FgMagenta, color.Bold)
	bold := color.New(color.Bold)

	dashes := func(n int) string {
		out := ""
		for i := 0; i < n; i++ {
			out += "─"
		}
		return out
	}

	fmt.Println()
	headerStyle.Printf("  Preview do dia — %s\n", now.Format("02/01/2006"))
	fmt.Println(muted.Sprint("  " + dashes(52)))

	fmt.Printf("  Entrada atual:     %s\n", bold.Sprint(open.In.Format("15:04")))
	fmt.Printf("  Horas já feitas:   %s\n", successBold.Sprint(report.FormatDuration(completedToday)))
	fmt.Printf("  Total acumulado:   %s\n", successBold.Sprint(report.FormatDuration(totalSoFar)))
	fmt.Printf("  Meta do dia:       %s\n", bold.Sprint(report.FormatDuration(goal)))
	fmt.Printf("  Faltam:            %s\n", warnStyle.Sprint(report.FormatDuration(remaining)))

	fmt.Println(muted.Sprint("  " + dashes(52)))
	fmt.Printf("  Saída prevista:    %s\n", successBold.Sprintf("⏰  %s", exitTime.Format("15:04")))

	if cfg.HasRate() {
		totalEarnings := cfg.Earn(goal.Hours())
		currentEarnings := cfg.Earn(totalSoFar.Hours())
		fmt.Printf("  Ganho atual:       %s\n", earnStyle.Sprint(currentEarnings))
		fmt.Printf("  Ganho ao finalizar:%s\n", earnStyle.Sprint(" "+totalEarnings))
	}

	fmt.Println()
}

func printGoalAlreadyReached(completed time.Duration, goal time.Duration) {
	successBold := color.New(color.FgGreen, color.Bold)
	muted := color.New(color.FgHiBlack)

	fmt.Println()
	successBold.Printf("  ✓ Meta do dia já atingida!\n")
	fmt.Printf("  Trabalhado hoje: %s  (meta: %s)\n",
		successBold.Sprint(report.FormatDuration(completed)),
		muted.Sprint(report.FormatDuration(goal)),
	)
	fmt.Println()
}
