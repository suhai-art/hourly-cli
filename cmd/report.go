package cmd

import (
	"fmt"
	"time"

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
  workhours report              # mês atual
  workhours report --month 2024-03`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Load()
			if err != nil {
				return err
			}

			var ref time.Time
			if month != "" {
				ref, err = time.ParseInLocation("2006-01", month, time.Local)
				if err != nil {
					return fmt.Errorf("formato inválido, use YYYY-MM")
				}
			} else {
				ref = time.Now()
			}

			entries := s.ByMonth(ref)
			title := ref.Format("Relatório — January 2006")

			color.New(color.FgCyan, color.Bold).Printf("\n  %s\n", title)
			fmt.Println(color.HiBlackString("  " + dashes(52)))

			// Group by day
			days := map[string][]store.Entry{}
			order := []string{}
			for _, e := range entries {
				key := e.In.Format("2006-01-02")
				if _, ok := days[key]; !ok {
					order = append(order, key)
				}
				days[key] = append(days[key], e)
			}

			var grandTotal time.Duration
			for _, key := range order {
				dayEntries := days[key]
				t, _ := time.ParseInLocation("2006-01-02", key, time.Local)
				dayLabel := t.Format("Mon 02/01")

				var dayTotal time.Duration
				for _, e := range dayEntries {
					dayTotal += e.Duration()
				}
				grandTotal += dayTotal

				fmt.Printf("  %s  %s\n",
					color.HiBlackString(dayLabel),
					color.GreenString(report.FormatDuration(dayTotal)),
				)
			}

			fmt.Println(color.HiBlackString("  " + dashes(52)))
			fmt.Printf("  Total no mês: %s  (%d dias)\n\n",
				color.New(color.FgGreen, color.Bold).Sprint(report.FormatDuration(grandTotal)),
				len(order),
			)
			return nil
		},
	}

	cmd.Flags().StringVarP(&month, "month", "m", "", "Mês (YYYY-MM)")
	return cmd
}

func dashes(n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += "─"
	}
	return out
}
