package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/backhub/functionality"
)

var (
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "backhub",
	Short: "GitHub repository backup tool using local mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

		cfg, err := functionality.LoadConfig(configPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load configuration")
		}

		token := os.Getenv("GH_TOKEN")
		if token == "" {
			log.Warn().Msg("GH_TOKEN not set - intended for use with GitHub API")
		}

		m := functionality.NewManager(token)
		return m.BackupAll(cfg.Repos)
	},
}

func main() {
	rootCmd.Flags().StringVarP(&configPath, "config", "c", ".backhub.yaml", "path to config file")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
