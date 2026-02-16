package discover

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Project represents a discovered project
type Project struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // vercel, swift, cli
}

// GitStatus holds git repository status
type GitStatus struct {
	Untracked int
	Modified  int
	Staged    int
	Branch    string
	Ahead     int
	Behind    int
}

// GitHubStatus holds GitHub repo status
type GitHubStatus struct {
	Issues int
	PRs    int
}

// CacheDir returns the cache directory path
func CacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hustlemc")
}

// LoadProjects loads projects from cache or runs discovery
func LoadProjects() ([]Project, error) {
	cacheFile := filepath.Join(CacheDir(), "projects.json")
	
	// Check if cache exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		// Run discovery
		if err := RunDiscovery(); err != nil {
			return nil, err
		}
	}
	
	// Read cache
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}
	
	return projects, nil
}

// RunDiscovery runs the mc-discover script
func RunDiscovery() error {
	home, _ := os.UserHomeDir()
	binPath := filepath.Join(home, "Projects", "mission-control", "bin", "mc-discover")
	
	cmd := exec.Command(binPath, filepath.Join(home, "Projects"), "--json")
	return cmd.Run()
}

// GetGitStatus returns git status for a project
func GetGitStatus(projectPath string) (*GitStatus, error) {
	expandedPath := expandPath(projectPath)
	
	// Check if it's a git repo
	gitDir := filepath.Join(expandedPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, nil
	}
	
	status := &GitStatus{}
	
	// Get porcelain status
	cmd := exec.Command("git", "-C", expandedPath, "status", "--porcelain", "-b")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, "## ") {
			// Parse branch line: ## main...origin/main [ahead 2, behind 1]
			parts := strings.Split(line[3:], "...")
			status.Branch = parts[0]
			if len(parts) > 1 {
				if strings.Contains(parts[1], "[ahead") {
					// Parse ahead/behind
				}
			}
			continue
		}
		
		if len(line) < 2 {
			continue
		}
		
		xy := line[:2]
		switch {
		case xy == "??":
			status.Untracked++
		case xy[0] != ' ' && xy[0] != '?':
			status.Staged++
			if xy[1] != ' ' {
				status.Modified++
			}
		case xy[1] != ' ' && xy[1] != '?':
			status.Modified++
		}
	}
	
	return status, nil
}

// GetGitHubStatus returns GitHub status (issues/PRs) for a project
func GetGitHubStatus(projectPath string) (*GitHubStatus, error) {
	expandedPath := expandPath(projectPath)
	
	status := &GitHubStatus{}
	
	// Get open issues count
	cmd := exec.Command("gh", "issue", "list", "--state", "open", "--json", "number", "-q", "length")
	cmd.Dir = expandedPath
	output, err := cmd.Output()
	if err == nil {
		var count int
		json.Unmarshal(output, &count)
		status.Issues = count
	}
	
	// Get open PRs count
	cmd = exec.Command("gh", "pr", "list", "--state", "open", "--json", "number", "-q", "length")
	cmd.Dir = expandedPath
	output, err = cmd.Output()
	if err == nil {
		var count int
		json.Unmarshal(output, &count)
		status.PRs = count
	}
	
	return status, nil
}

// GetVercelStatus returns the latest deployment status
func GetVercelStatus(projectPath string) (string, error) {
	expandedPath := expandPath(projectPath)
	
	// Check if it's a Vercel project
	vercelDir := filepath.Join(expandedPath, ".vercel")
	if _, err := os.Stat(vercelDir); os.IsNotExist(err) {
		return "", nil
	}
	
	// Get latest deployment
	cmd := exec.Command("vercel", "ls", "--json", "-n", "1")
	cmd.Dir = expandedPath
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}
	
	// Parse JSON output
	var deployments []struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(output, &deployments); err != nil {
		return "unknown", nil
	}
	
	if len(deployments) > 0 {
		state := strings.ToLower(deployments[0].State)
		switch state {
		case "ready":
			return "ready", nil
		case "building":
			return "building", nil
		case "queued":
			return "queued", nil
		case "error":
			return "failed", nil
		default:
			return state, nil
		}
	}
	
	return "ready", nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
