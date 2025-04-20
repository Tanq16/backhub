package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tanq16/backhub/functionality"
)

var BackHubVersion = "dev"
var unlimitedOutput bool

var rootCmd = &cobra.Command{
	Use:     "backhub",
	Short:   "GitHub repository backup tool using local mirrors",
	Version: BackHubVersion,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]
		token := os.Getenv("GH_TOKEN")
		handler := functionality.NewHandler(token)
		err := handler.RunBackup(configPath, unlimitedOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVar(&unlimitedOutput, "debug", false, "Disable output size limit")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
