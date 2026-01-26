package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grokify/releasekit/output"
)

var (
	version = "dev"
	format  string
)

var rootCmd = &cobra.Command{
	Use:   "releasekit",
	Short: "Release management toolkit for Go projects",
	Long:  "ReleaseKit provides git operations, commit analysis, and release workflow automation.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("releasekit", version)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&format, "format", "toon", "Output format: toon, json, text")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(commitsCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(modulesCmd)
	rootCmd.AddCommand(validateCmd)
}

func getFormatter() (output.Formatter, error) {
	return output.NewFormatter(output.Format(format))
}

func printFormatted(v any) error {
	f, err := getFormatter()
	if err != nil {
		return err
	}
	data, err := f.Format(v)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
