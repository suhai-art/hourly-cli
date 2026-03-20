package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/suhai-art/hourly-cli/internal/store"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "import <arquivo.csv>",
		Short: "Importa horas de um CSV do Jira",
		Long: `Importa registros de horas de um CSV exportado pelo Jira.

Colunas esperadas: Date, Time Seconds (ou Time), Member, Project, Issue, Comment

Se já existir um registro importado no mesmo dia, as horas são somadas.

Exemplos:
  hourly import jira.csv
  hourly import jira.csv --dry-run   # mostra o que seria importado sem salvar`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("não foi possível abrir o arquivo: %w", err)
			}
			defer f.Close()

			rows, err := parseJiraCSV(f)
			if err != nil {
				return err
			}

			if len(rows) == 0 {
				color.Yellow("Nenhum registro encontrado no CSV.")
				return nil
			}

			if dryRun {
				printDryRun(rows)
				return nil
			}

			s, err := store.Load()
			if err != nil {
				return err
			}

			added, merged := 0, 0
			for _, row := range rows {
				if s.AddOrMerge(row) {
					merged++
				} else {
					added++
				}
			}

			if err := s.Save(); err != nil {
				return err
			}

			fmt.Printf("\n")
			if added > 0 {
				color.Green("✓ %d registro(s) adicionado(s)", added)
			}
			if merged > 0 {
				color.Cyan("⊕ %d registro(s) somado(s) em dias existentes", merged)
			}
			fmt.Printf("  Total importado: %d linha(s) do CSV\n\n", len(rows))
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Mostra o que seria importado sem salvar")
	return cmd
}

type jiraRow struct {
	Date    time.Time
	Seconds int
	Member  string
	Project string
	Issue   string
	Comment string
}

func parseJiraCSV(r io.Reader) ([]store.Entry, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler cabeçalho: %w", err)
	}

	// Map header names to indices (case-insensitive)
	idx := map[string]int{}
	for i, h := range headers {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	required := []string{"date", "time seconds"}
	for _, req := range required {
		if _, ok := idx[req]; !ok {
			// Try fallback for "time seconds"
			if req == "time seconds" {
				if _, ok2 := idx["time"]; !ok2 {
					return nil, fmt.Errorf("coluna obrigatória ausente: %q", req)
				}
			}
		}
	}

	muted := color.New(color.FgHiBlack)
	warn := color.New(color.FgYellow)

	var entries []store.Entry
	lineNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		lineNum++
		if err != nil {
			warn.Printf("  linha %d ignorada: %v\n", lineNum, err)
			continue
		}

		// Parse date
		dateStr := strings.Trim(getField(record, idx, "date"), `"`)
		date, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			warn.Printf("  linha %d ignorada: data inválida %q\n", lineNum, dateStr)
			continue
		}

		// Parse duration in seconds
		var seconds int
		if secIdx, ok := idx["time seconds"]; ok {
			secStr := strings.Trim(record[secIdx], `"`)
			seconds, _ = strconv.Atoi(secStr)
		}
		// Fallback: parse "8h" or "1h 30m" from Time column
		if seconds == 0 {
			timeStr := strings.Trim(getField(record, idx, "time"), `"`)
			seconds = parseHumanDuration(timeStr)
		}

		if seconds <= 0 {
			muted.Printf("  linha %d ignorada: duração zero\n", lineNum)
			continue
		}

		// Build entry: In = date 00:00, Out = In + seconds
		in := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		out := in.Add(time.Duration(seconds) * time.Second)

		entry := store.Entry{
			ID:   store.NewID() + fmt.Sprintf("%04d", lineNum),
			In:   in,
			Out:  &out,
			Note: "jira-import",
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func getField(record []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(record) {
		return ""
	}
	return record[i]
}

// parseHumanDuration parses strings like "8h", "1h 30m", "45m" into seconds.
func parseHumanDuration(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	total := 0
	// Remove quotes
	s = strings.Trim(s, `"`)

	parts := strings.Fields(s)
	for _, p := range parts {
		if strings.HasSuffix(p, "h") {
			v, _ := strconv.Atoi(strings.TrimSuffix(p, "h"))
			total += v * 3600
		} else if strings.HasSuffix(p, "m") {
			v, _ := strconv.Atoi(strings.TrimSuffix(p, "m"))
			total += v * 60
		}
	}
	return total
}

func printDryRun(entries []store.Entry) {
	header := color.New(color.FgCyan, color.Bold)
	muted := color.New(color.FgHiBlack)
	success := color.New(color.FgGreen)

	header.Printf("\n  Dry-run — %d registro(s) seriam importados:\n\n", len(entries))

	for _, e := range entries {
		dur := e.Duration()
		h := int(dur.Hours())
		m := int(dur.Minutes()) % 60
		fmt.Printf("  %s  %s\n",
			muted.Sprint(e.In.Format("02/01/2006")),
			success.Sprintf("%dh%02dm", h, m),
		)
	}
	fmt.Println()
}
