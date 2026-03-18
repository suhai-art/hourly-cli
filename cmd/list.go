package cmd

import (
	"time"

	"github.com/suhai-art/hourly-cli/internal/config"
	"github.com/suhai-art/hourly-cli/internal/report"
	"github.com/suhai-art/hourly-cli/internal/store"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		week  bool
		month bool
		day   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista registros de horas",
		Long: `Lista registros de horas trabalhadas.

Exemplos:
  hourly list           # hoje
  hourly list --week    # semana atual
  hourly list --month   # mês atual
  hourly list --day 2024-03-15`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Load()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			now := time.Now()
			var entries []store.Entry
			var title string

			switch {
			case day != "":
				t, err := time.ParseInLocation("2006-01-02", day, time.Local)
				if err != nil {
					return err
				}
				entries = s.ByDate(t)
				title = "Dia " + t.Format("02/01/2006")

			case week:
				entries = s.ByWeek(now)
				_, w := now.ISOWeek()
				title = now.Format("Semana ") + string(rune('0'+w/10)) + string(rune('0'+w%10)) + now.Format(" de January 2006")

			case month:
				entries = s.ByMonth(now)
				title = now.Format("January 2006")

			default:
				entries = s.ByDate(now)
				title = "Hoje, " + now.Format("02/01/2006")
			}

			report.PrintEntries(entries, title, cfg)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&week, "week", "w", false, "Semana atual")
	cmd.Flags().BoolVarP(&month, "month", "m", false, "Mês atual")
	cmd.Flags().StringVarP(&day, "day", "d", "", "Data específica (YYYY-MM-DD)")
	return cmd
}
