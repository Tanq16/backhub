package functionality

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/tanq16/backhub/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Repos []string `yaml:"repos"`
}

type Handler struct {
	token       string
	outputMgr   *utils.Manager
	concurrency int
	repos       []string
	cloneFolder string
}

func NewHandler(token string) *Handler {
	return &Handler{
		token:       token,
		outputMgr:   utils.NewManager(15),
		concurrency: 5,
		cloneFolder: ".",
	}
}

func (h *Handler) Setup() {
	h.outputMgr.Register("logistics")
	h.outputMgr.SetMessage("logistics", "Setting up BackHub")
	h.outputMgr.StartDisplay()
}

// Loads repository configuration from a file or direct repo path
func (h *Handler) LoadConfig(path string) error {
	h.outputMgr.AddStreamLine("logistics", fmt.Sprintf("Loading configuration from '%s'", path))
	repoRegex := "^github.com/[^/]+/[^/]+$"
	if regexp.MustCompile(repoRegex).MatchString(path) {
		h.repos = []string{path}
		h.outputMgr.AddStreamLine("logistics", "Direct repo specified, using it as configuration")
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		h.outputMgr.AddStreamLine("logistics", "Failed to read config file")
		return fmt.Errorf("reading config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		h.outputMgr.AddStreamLine("logistics", "Failed to parse YAML configuration")
		return fmt.Errorf("parsing config: %w", err)
	}
	h.repos = cfg.Repos
	h.outputMgr.AddStreamLine("logistics", fmt.Sprintf("Loaded %d repositories", len(h.repos)))
	return nil
}

// Validates the GitHub token
func (h *Handler) ValidateToken() error {
	if h.token == "" {
		h.outputMgr.AddStreamLine("logistics", "proceeding without GitHub token")
		h.outputMgr.SetStatus("logistics", "warning")
	} else {
		h.outputMgr.SetStatus("logistics", "pending")
		h.outputMgr.AddStreamLine("logistics", "GitHub token is set")
	}
	return nil
}

// Performs the backup operation for all repositories
func (h *Handler) ExecuteBackup() error {
	repoCount := len(h.repos)
	wg := &sync.WaitGroup{}
	toProcess := make(chan string, repoCount)

	h.outputMgr.SetMessage("logistics", fmt.Sprintf("Processing %d repositories", repoCount))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, repo := range h.repos {
			toProcess <- repo
		}
		close(toProcess)
	}()

	// Start consumer pool
	for i := range h.concurrency {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for repo := range toProcess {
				taskName := fmt.Sprintf("repo-%s", repo)
				h.outputMgr.Register(taskName)
				h.outputMgr.SetMessage(taskName, fmt.Sprintf("Processing %s", repo))
				if err := h.backupRepo(repo, taskName); err != nil {
					h.outputMgr.ReportError(taskName, err)
				} else {
					h.outputMgr.SetMessage(taskName, fmt.Sprintf("%s backed up successfully", repo))
					h.outputMgr.Complete(taskName)
				}
			}
		}(i)
	}
	wg.Wait()

	// Final summary
	h.outputMgr.SetMessage("logistics", "Backup process completed")
	h.outputMgr.Complete("logistics")
	h.outputMgr.StopDisplay()
	return nil
}

// Entry point to run the backup process
func (h *Handler) RunBackup(configPath string, unlimitedOutput bool) error {
	h.outputMgr.SetUnlimitedOutput(unlimitedOutput)
	h.Setup()
	if err := h.ValidateToken(); err != nil {
		h.outputMgr.ReportError("logistics", err)
		h.outputMgr.StopDisplay()
		return err
	}
	if err := h.LoadConfig(configPath); err != nil {
		h.outputMgr.ReportError("logistics", err)
		h.outputMgr.StopDisplay()
		return err
	}
	h.outputMgr.SetMessage("logistics", "Backup logistics completed")
	return h.ExecuteBackup()
}
