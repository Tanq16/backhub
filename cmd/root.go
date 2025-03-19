package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/backhub/functionality"
)

var (
	configPath string
)

var BackHubVersion = "dev"

var rootCmd = &cobra.Command{
	Use:     "backhub",
	Short:   "GitHub repository backup tool using local mirrors",
	Version: BackHubVersion,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

		configPath = args[0]
		cfg, err := functionality.LoadConfig(configPath)
		if configPath != "" {
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}
		}
		token := os.Getenv("GH_TOKEN")
		if token == "" {
			log.Fatal().Msg("GH_TOKEN not set - intended for use with GitHub API")
		}
		m := functionality.NewManager(token)
		err = m.BackupAll(cfg.Repos)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to backup repositories")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
