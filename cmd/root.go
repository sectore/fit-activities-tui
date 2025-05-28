package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sectore/fit-activities-tui/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "fit-activities-tui",
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

		program := tea.NewProgram(
			tui.InitialModel(path),
			tea.WithAltScreen(),
		)
		_, err := program.Run()
		if err != nil {
			return fmt.Errorf("Could not run program: %v", err)
		}

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
