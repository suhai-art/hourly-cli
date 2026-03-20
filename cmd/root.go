package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hourly",
	Short: "Calculadora de horas trabalhadas",
	Long: color.New(color.FgCyan, color.Bold).Sprint("⏱  hourly") + `

Registre entradas e saídas e calcule automaticamente
as horas trabalhadas por dia, semana ou mês.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Erro: %v", err))
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		newAddCmd(),
		newListCmd(),
		newReportCmd(),
		newDeleteCmd(),
		newConfigCmd(),
		newImportCmd(),
	)
}
