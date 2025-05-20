package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sectore/fit-sum-tui/internal/fit"
	"github.com/sectore/fit-sum-tui/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "fit-sum-tui",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("import")
		// if no path provided, try to use current directory
		if path == "" {
			d, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %v", err)
			}
			path = d
		}

		fitFiles, err := fit.GetFitFiles(path)
		if err != nil {
			return err
		}

		program := tea.NewProgram(
			tui.InitialModel(path, fitFiles),
			tea.WithAltScreen(),
		)
		model, err := program.Run()
		if err != nil {
			return fmt.Errorf("Error running program: %v", err)
		}
		fmt.Printf("TUI exited -> state -> %v", model)

		return nil

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	rootCmd.PersistentFlags().StringP("import", "i", "", "Import directory")
}
