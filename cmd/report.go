package cmd

import (
	"fmt"
	"time"

	"github.com/suhai-art/hourly-cli/internal/config"
	"github.com/suhai-art/hourly-cli/internal/report"
	"github.com/suhai-art/hourly-cli/internal/store"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var month string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Relatório de horas por dia no mês",
		Long: `Exibe um relatório consolidado do mês.

Exemplos:
  hourly report              # mês atual
  hourly report --month 2024-03`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Load()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			ref, err := parseMonthRef(month)
			if err != nil {
				return err
			}

			entries := s.ByMonth(ref)
			printMonthReport(entries, ref, cfg)
			return nil
		},
	}

	cmd.Flags().StringVarP(&month, "month", "m", "", "Mês (YYYY-MM)")
	return cmd
}

// parseMonthRef parses the --month flag or returns the current month.
func parseMonthRef(month string) (time.Time, error) {
	if month == "" {
		return time.Now(), nil
	}
	ref, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("formato inválido, use YYYY-MM")
	}
	return ref, nil
}

// printMonthReport renders the consolidated monthly report.
func printMonthReport(entries []store.Entry, ref time.Time, cfg *config.Config) {
	headerStyle := color.New(color.FgCyan, color.Bold)
	muted := color.New(color.FgHiBlack)
	successBold := color.New(color.FgGreen, color.Bold)
	earnStyle := color.New(color.FgMagenta, color.Bold)

	title := ref.Format("Relatório — January 2006")
	headerStyle.Printf("\n  %s\n", title)
	fmt.Println(muted.Sprint("  " + dashes(62)))

	days, order := groupEntriesByDay(entries)

	var grandTotal time.Duration
	var grandEarnings float64

	for _, key := range order {
		dayTotal := sumDayDuration(days[key])
		grandTotal += dayTotal

		t, _ := time.ParseInLocation("2006-01-02", key, time.Local)
		dayLabel := t.Format("Mon 02/01")

		line := fmt.Sprintf("  %s  %s",
			muted.Sprint(dayLabel),
			color.GreenString("%-9s", report.FormatDuration(dayTotal)),
		)

		if cfg.HasRate() {
			hours := dayTotal.Hours()
			grandEarnings += hours * cfg.HourlyRate
			line += fmt.Sprintf("  %s", earnStyle.Sprintf("%s", cfg.Earn(hours)))
		}

		fmt.Println(line)
	}

	fmt.Println(muted.Sprint("  " + dashes(62)))

	totalStr := fmt.Sprintf("%-9s", report.FormatDuration(grandTotal))
	if cfg.HasRate() {
		fmt.Printf("  Total no mês: %s  %s  %s\n\n",
			successBold.Sprint(totalStr),
			earnStyle.Sprintf("%s %.2f", cfg.Currency, grandEarnings),
			muted.Sprintf("(%d dias)", len(order)),
		)
	} else {
		fmt.Printf("  Total no mês: %s  %s\n\n",
			successBold.Sprint(totalStr),
			muted.Sprintf("(%d dias)", len(order)),
		)
	}
}

// groupEntriesByDay groups entries by their date key (YYYY-MM-DD),
// preserving insertion order.
func groupEntriesByDay(entries []store.Entry) (map[string][]store.Entry, []string) {
	days := map[string][]store.Entry{}
	order := []string{}

	for _, e := range entries {
		key := e.In.Format("2006-01-02")
		if _, exists := days[key]; !exists {
			order = append(order, key)
		}
		days[key] = append(days[key], e)
	}

	return days, order
}

// sumDayDuration sums the duration of all completed entries in a day.
func sumDayDuration(entries []store.Entry) time.Duration {
	var total time.Duration
	for _, e := range entries {
		total += e.Duration()
	}
	return total
}

func dashes(n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += "─"
	}
	return out
}
