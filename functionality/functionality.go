package functionality

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type config struct {
	Repos []string `yaml:"repos"`
}

func LoadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

type manager struct {
	token string
}

func NewManager(token string) *manager {
	return &manager{token: token}
}

func (m *manager) BackupAll(repos []string) error {
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
	auth := &http.BasicAuth{
		Username: "backhub",
		Password: m.token,
	}
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		return m.cloneRepo(repoURL, folderName, auth)
	}
	return m.updateRepo(folderName, auth)
}

func (m *manager) cloneRepo(repoURL, folderName string, auth *http.BasicAuth) error {
	log.Info().Str("repo", repoURL).Msg("cloning repository")
	_, err := git.PlainClone(folderName, true, &git.CloneOptions{
		URL:      repoURL,
		Auth:     auth,
		Mirror:   true,
		Progress: nil,
	})
	return err
}

func (m *manager) updateRepo(folderName string, auth *http.BasicAuth) error {
	log.Info().Str("folder", folderName).Msg("updating repository")
	repo, err := git.PlainOpen(folderName)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	err = repo.Fetch(&git.FetchOptions{
		Auth:     auth,
		Force:    true,
		Progress: nil,
		Tags:     git.AllTags,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (m *manager) buildRepoURL(repo string) string {
	return fmt.Sprintf("https://%s", repo)
}

func getLocalFolderName(repo string) string {
	base := filepath.Base(repo)
	return base + ".git"
}
