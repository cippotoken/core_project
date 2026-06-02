package githelper

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RunCommand executes a system command in the specified directory and returns stdout/err.
func RunCommand(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed running '%s %s': %w\nstderr: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// SetupGitRepo initializes git in the working directory and configures the origin remote if specified.
func SetupGitRepo(dir string, remoteURL string) error {
	// 1. Check if .git directory exists
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		fmt.Println("[GitHelper] Initializing new git repository...")
		_, err = RunCommand(dir, "git", "init", "-b", "main")
		if err != nil {
			// Fallback if git version is old and doesn't support -b
			_, err = RunCommand(dir, "git", "init")
			if err != nil {
				return err
			}
			_, _ = RunCommand(dir, "git", "checkout", "-b", "main")
		}
	}

	if remoteURL == "" {
		return nil
	}

	// 2. Set or add origin remote
	remotes, err := RunCommand(dir, "git", "remote")
	if err != nil {
		return err
	}

	hasOrigin := false
	for _, r := range strings.Split(remotes, "\n") {
		if strings.TrimSpace(r) == "origin" {
			hasOrigin = true
			break
		}
	}

	if hasOrigin {
		fmt.Printf("[GitHelper] Setting remote 'origin' URL to %s...\n", remoteURL)
		_, err = RunCommand(dir, "git", "remote", "set-url", "origin", remoteURL)
	} else {
		fmt.Printf("[GitHelper] Adding remote 'origin' URL %s...\n", remoteURL)
		_, err = RunCommand(dir, "git", "remote", "add", "origin", remoteURL)
	}
	return err
}

// CommitAndPush stages the .env file, commits it with a timestamp, and pushes it to GitHub.
func CommitAndPush(dir string) error {
	fmt.Println("[GitHelper] Staging file .env...")
	_, err := RunCommand(dir, "git", "add", ".env")
	if err != nil {
		return err
	}

	// Check if there is at least one commit in the repository
	hasCommits := false
	_, err = RunCommand(dir, "git", "rev-parse", "HEAD")
	if err == nil {
		hasCommits = true
	}

	commitMsg := fmt.Sprintf("Update 100k generated wallets: %s", time.Now().Format(time.RFC3339))
	if hasCommits {
		fmt.Printf("[GitHelper] Amending existing commit to prevent history bloat: '%s'...\n", commitMsg)
		_, err = RunCommand(dir, "git", "commit", "--amend", "-m", commitMsg, "--allow-empty")
	} else {
		fmt.Printf("[GitHelper] Creating initial commit: '%s'...\n", commitMsg)
		_, err = RunCommand(dir, "git", "commit", "-m", commitMsg, "--allow-empty")
	}
	if err != nil {
		return err
	}

	fmt.Println("[GitHelper] Pushing changes to remote repository...")
	// Force push to overwrite old commits and keep history completely flat
	output, err := RunCommand(dir, "git", "push", "-u", "origin", "main", "--force")
	if err != nil {
		return fmt.Errorf("push failed: %w. Make sure your GitHub SSH/token credentials are set up correctly on your machine.", err)
	}

	fmt.Printf("[GitHelper] Git push successful!\n%s\n", output)

	// Prune loose objects from the amended commit to keep the .git folder small and efficient
	fmt.Println("[GitHelper] Pruning unreferenced loose objects to keep disk usage and CPU low...")
	_, _ = RunCommand(dir, "git", "reflog", "expire", "--expire=now", "--all")
	_, _ = RunCommand(dir, "git", "prune", "--expire", "now")

	return nil
}
