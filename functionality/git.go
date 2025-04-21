package functionality

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Handles the cloning or updating of a single repository
func (h *Handler) backupRepo(repo, taskName string) error {
	folderName := getLocalFolderName(repo)
	repoURL := h.buildRepoURL(repo)
	h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Preparing to backup repository from %s", repoURL))
	h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Target directory: %s", folderName))
	// Set up authentication if token exists
	var auth *http.BasicAuth
	if h.token != "" {
		auth = &http.BasicAuth{
			Username: "backhub",
			Password: h.token,
		}
	}
	// Check if repository exists locally
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		return h.cloneRepo(repoURL, folderName, auth, taskName)
	}
	h.outputMgr.AddStreamLine(taskName, "Repository exists locally, will update")
	return h.updateRepo(folderName, auth, taskName)
}

// Clones a repository as a mirror
func (h *Handler) cloneRepo(repoURL, folderName string, auth *http.BasicAuth, taskName string) error {
	h.outputMgr.SetMessage(taskName, fmt.Sprintf("Cloning %s", repoURL))
	h.outputMgr.AddStreamLine(taskName, "Starting clone operation")
	_, err := git.PlainClone(folderName, true, &git.CloneOptions{
		URL:      repoURL,
		Auth:     auth,
		Mirror:   true,
		Progress: nil,
	})
	if err != nil {
		h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Clone failed: %s", err))
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	h.outputMgr.AddStreamLine(taskName, "Clone completed successfully")
	h.outputMgr.SetMessage(taskName, fmt.Sprintf("Successfully cloned %s", repoURL))
	return nil
}

// Updates an existing repository
func (h *Handler) updateRepo(folderName string, auth *http.BasicAuth, taskName string) error {
	h.outputMgr.SetMessage(taskName, fmt.Sprintf("Updating %s", folderName))
	h.outputMgr.AddStreamLine(taskName, "Opening local repository")
	repo, err := git.PlainOpen(folderName)
	if err != nil {
		h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Failed to open repository: %s", err))
		return fmt.Errorf("failed to open repository: %w", err)
	}
	h.outputMgr.AddStreamLine(taskName, "Fetching updates from remote")
	err = repo.Fetch(&git.FetchOptions{
		Auth:     auth,
		Force:    true,
		Progress: nil,
		Tags:     git.AllTags,
	})
	if err == git.NoErrAlreadyUpToDate {
		h.outputMgr.AddStreamLine(taskName, "Repository already up to date")
		h.outputMgr.SetMessage(taskName, fmt.Sprintf("Repository %s is already up to date", folderName))
		return nil
	}
	if err != nil {
		h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Fetch failed: %s", err))
		return fmt.Errorf("failed to fetch updates: %w", err)
	}
	h.outputMgr.AddStreamLine(taskName, "Repository updated successfully")
	h.outputMgr.SetMessage(taskName, fmt.Sprintf("Successfully updated %s", folderName))
	return nil
}

// Constructs the HTTPS URL for a repository
func (h *Handler) buildRepoURL(repo string) string {
	return fmt.Sprintf("https://%s", repo)
}

// Generates the local folder name for a repository
func getLocalFolderName(repo string) string {
	base := filepath.Base(repo)
	return base + ".git"
}
