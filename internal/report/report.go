package report

import (
	"fmt"
	"time"

	"github.com/suhai-art/hourly-cli/internal/config"
	"github.com/suhai-art/hourly-cli/internal/store"

	"github.com/fatih/color"
)

var (
	header  = color.New(color.FgCyan, color.Bold)
	success = color.New(color.FgGreen, color.Bold)
	earn    = color.New(color.FgMagenta, color.Bold)
	warn    = color.New(color.FgYellow)
	muted   = color.New(color.FgHiBlack)
	bold    = color.New(color.Bold)
)

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%02dm", h, m)
}

func PrintEntries(entries []store.Entry, title string, cfg *config.Config) {
	if len(entries) == 0 {
		warn.Printf("Nenhum registro encontrado para %s.\n", title)
		return
	}

	header.Printf("\n  %s\n", title)
	fmt.Println(muted.Sprint("  " + repeat("─", 60)))

	var total time.Duration
	for _, e := range entries {
		dateStr := e.In.Format("02/01 Mon")
		inStr := e.In.Format("15:04")

		if e.IsOpen() {
			fmt.Printf("  %s  %s  %s  %s",
				muted.Sprint(dateStr),
				bold.Sprint(inStr),
				warn.Sprint("→  aberto  "),
				muted.Sprint(e.ID),
			)
		} else {
			dur := e.Duration()
			total += dur
			outStr := e.Out.Format("15:04")
			durStr := fmt.Sprintf("%-9s", FormatDuration(dur))

			fmt.Printf("  %s  %s → %s  %s",
				muted.Sprint(dateStr),
				bold.Sprint(inStr),
				bold.Sprint(outStr),
				success.Sprint(durStr),
			)

			if cfg != nil && cfg.HasRate() {
				hours := dur.Hours()
				fmt.Printf("  %s", earn.Sprintf("%-12s", cfg.Earn(hours)))
			}
		}
		if e.Note != "" {
			fmt.Printf("  %s", muted.Sprint(e.Note))
		}
		fmt.Println()
	}

	fmt.Println(muted.Sprint("  " + repeat("─", 60)))

	totalStr := fmt.Sprintf("%-9s", FormatDuration(total))
	if cfg != nil && cfg.HasRate() {
		fmt.Printf("  Total: %s  %s\n\n",
			success.Sprint(totalStr),
			earn.Sprintf("%s", cfg.Earn(total.Hours())),
		)
	} else {
		fmt.Printf("  Total: %s\n\n", success.Sprint(totalStr))
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
