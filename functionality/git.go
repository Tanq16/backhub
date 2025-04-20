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
	auth := &http.BasicAuth{
		Username: "backhub",
		Password: h.token,
	}
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		return h.cloneRepo(repoURL, folderName, auth, taskName)
	}
	return h.updateRepo(folderName, auth, taskName)
}

// Clones a repository as a mirror
func (h *Handler) cloneRepo(repoURL, folderName string, auth *http.BasicAuth, taskName string) error {
	h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Cloning %s to %s", repoURL, folderName))
	if auth.Password == "" {
		auth = nil
	}
	_, err := git.PlainClone(folderName, true, &git.CloneOptions{
		URL:      repoURL,
		Auth:     auth,
		Mirror:   true,
		Progress: nil,
	})
	if err != nil {
		h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Failed to clone repository: %s", err))
		return err
	}
	h.outputMgr.AddStreamLine(taskName, "Clone completed successfully")
	return nil
}

// Updates an existing repository
func (h *Handler) updateRepo(folderName string, auth *http.BasicAuth, taskName string) error {
	h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Updating existing repository at %s", folderName))
	repo, err := git.PlainOpen(folderName)
	if err != nil {
		h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Failed to open repository: %s", err))
		return fmt.Errorf("failed to open repository: %w", err)
	}
	err = repo.Fetch(&git.FetchOptions{
		Auth:     auth,
		Force:    true,
		Progress: nil,
		Tags:     git.AllTags,
	})
	if err == git.NoErrAlreadyUpToDate {
		h.outputMgr.AddStreamLine(taskName, "Repository already up to date")
		return nil
	}
	if err != nil {
		h.outputMgr.AddStreamLine(taskName, fmt.Sprintf("Failed to fetch updates: %s", err))
		return err
	}
	h.outputMgr.AddStreamLine(taskName, "Repository updated successfully")
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
