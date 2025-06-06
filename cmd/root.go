package cmd

import (
	"fmt"
	"io"
	"log"
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

		doLogging, _ := cmd.Flags().GetBool("log")
		if doLogging {
			f, err := tea.LogToFile("debug.log", "debug")
			if err != nil {
				fmt.Println("failed to initial logging:", err)
				os.Exit(1)
			}
			log.Printf("logging enabled")
			defer f.Close()
		} else {
			log.SetOutput(io.Discard)
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
	rootCmd.PersistentFlags().StringP("import", "i", "", "path to single FIT file or directory of FIT files to import")
	rootCmd.PersistentFlags().Bool("log", false, "enable logging to store logs into 'debug.log'")
}
