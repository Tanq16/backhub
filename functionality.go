package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type config struct {
	Repos []string `mapstructure:"repos"`
}

func loadConfig(path string) (*config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		log.Error().Err(err).Str("path", path).Msg("failed to read config file")
		return nil, err
	}
	var cfg config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type manager struct {
	token string
}

func newManager(token string) *manager {
	return &manager{token: token}
}

func (m *manager) backupAll(repos []string) error {
	toProcess := make(chan string)
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(toProcess)
		for _, repo := range repos {
			toProcess <- repo
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for repo := range toProcess {
			wg.Add(1)
			go func(repo string) {
				defer wg.Done()
				if err := m.backupRepo(repo); err != nil {
					log.Error().Err(err).Str("repo", repo).Msg("failed to backup repository")
				} else {
					log.Info().Str("repo", repo).Msg("repository backed up")
				}
			}(repo)
		}
	}()

	wg.Wait()
	return nil
}

func (m *manager) backupRepo(repo string) error {
	folderName := getLocalFolderName(repo)
	repoURL := m.buildRepoURL(repo)
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		return m.cloneRepo(repoURL, folderName)
	}
	return m.updateRepo(repoURL, folderName)
}

func (m *manager) cloneRepo(repoURL, folderName string) error {
	log.Info().Str("repo", repoURL).Msg("cloning repository")
	cmd := exec.Command("git", "clone", "--mirror", repoURL, folderName)
	return m.runGitCommand(cmd)
}

func (m *manager) updateRepo(repoURL, folderName string) error {
	log.Info().Str("repo", repoURL).Msg("updating repository")
	setURLCmd := exec.Command("git", "remote", "set-url", "origin", repoURL)
	setURLCmd.Dir = folderName
	if err := m.runGitCommand(setURLCmd); err != nil {
		return err
	}
	fetchCmd := exec.Command("git", "fetch", "--all")
	fetchCmd.Dir = folderName
	return m.runGitCommand(fetchCmd)
}

func (m *manager) runGitCommand(cmd *exec.Cmd) error {
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %s: %w", string(output), err)
	}
	return nil
}

func (m *manager) buildRepoURL(repo string) string {
	if m.token == "" {
		return fmt.Sprintf("https://%s", repo)
	}
	return fmt.Sprintf("https://%s@%s", m.token, repo)
}

func getLocalFolderName(repo string) string {
	base := filepath.Base(repo)
	return base + ".git"
}
