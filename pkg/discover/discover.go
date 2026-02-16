package discover

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

// ProjectCache holds cached status for a project
type ProjectCache struct {
	UpdatedAt   time.Time   `json:"updated_at"`
	Language    string      `json:"language,omitempty"`
	GitStatus   *GitStatus  `json:"git_status,omitempty"`
	GHStatus    *GitHubStatus `json:"gh_status,omitempty"`
	VercelState string      `json:"vercel_state,omitempty"`
	FirstCommit int64       `json:"first_commit,omitempty"` // Unix timestamp
	LastCommit  int64       `json:"last_commit,omitempty"`  // Unix timestamp
}

const CacheTTL = 5 * time.Minute // Cache validity duration

// CacheDir returns the global cache directory path
func CacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hustlemc")
}

// ProjectCacheDir returns the cache directory for a specific project
func ProjectCacheDir(projectPath string) string {
	return filepath.Join(expandPath(projectPath), ".hustlemc")
}

// LoadProjectCache loads cached status for a project
func LoadProjectCache(projectPath string) (*ProjectCache, error) {
	cacheFile := filepath.Join(ProjectCacheDir(projectPath), "status.json")
	
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	
	var cache ProjectCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	
	// Check if cache is still valid
	if time.Since(cache.UpdatedAt) > CacheTTL {
		return nil, fmt.Errorf("cache expired")
	}
	
	return &cache, nil
}

// SaveProjectCache saves status cache for a project
func SaveProjectCache(projectPath string, cache *ProjectCache) error {
	cacheDir := ProjectCacheDir(projectPath)
	
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	
	cache.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(cacheDir, "status.json"), data, 0644)
}

// UpdateProjectCache updates specific fields in the project cache
func UpdateProjectCache(projectPath string, updates func(*ProjectCache)) error {
	cache, _ := LoadProjectCache(projectPath)
	if cache == nil {
		cache = &ProjectCache{}
	}
	
	updates(cache)
	
	return SaveProjectCache(projectPath, cache)
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

// GetGitStatus returns git status for a project using mc-git-status script
func GetGitStatus(projectPath string) (*GitStatus, error) {
	expandedPath := expandPath(projectPath)
	
	// Check if it's a git repo
	gitDir := filepath.Join(expandedPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, nil
	}
	
	// Git status changes frequently, so we always fetch fresh
	// But we still save to cache for reference
	
	// Use mc-git-status script
	home, _ := os.UserHomeDir()
	binPath := filepath.Join(home, "Projects", "mission-control", "bin", "mc-git-status")
	
	cmd := exec.Command(binPath, expandedPath, "--json")
	output, err := cmd.Output()
	
	var status *GitStatus
	if err != nil {
		// Fallback to direct git
		status, _ = getGitStatusDirect(expandedPath)
	} else {
		var result struct {
			Branch    string `json:"branch"`
			Untracked int    `json:"untracked"`
			Modified  int    `json:"modified"`
			Staged    int    `json:"staged"`
			Ahead     int    `json:"ahead"`
			Behind    int    `json:"behind"`
		}
		if err := json.Unmarshal(output, &result); err != nil {
			status, _ = getGitStatusDirect(expandedPath)
		} else {
			status = &GitStatus{
				Branch:    result.Branch,
				Untracked: result.Untracked,
				Modified:  result.Modified,
				Staged:    result.Staged,
				Ahead:     result.Ahead,
				Behind:    result.Behind,
			}
		}
	}
	
	// Update cache
	if status != nil {
		UpdateProjectCache(projectPath, func(c *ProjectCache) {
			c.GitStatus = status
		})
	}
	
	return status, nil
}

// getGitStatusDirect is a fallback using git directly
func getGitStatusDirect(expandedPath string) (*GitStatus, error) {
	status := &GitStatus{}
	
	cmd := exec.Command("git", "-C", expandedPath, "status", "--porcelain", "-b")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, "## ") {
			parts := strings.Split(line[3:], "...")
			status.Branch = parts[0]
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

// GetGitHubStatus returns GitHub status (issues/PRs) for a project using mc-gh-status script
func GetGitHubStatus(projectPath string) (*GitHubStatus, error) {
	expandedPath := expandPath(projectPath)
	
	// Use mc-gh-status script
	home, _ := os.UserHomeDir()
	binPath := filepath.Join(home, "Projects", "mission-control", "bin", "mc-gh-status")
	
	cmd := exec.Command(binPath, expandedPath, "--json")
	output, err := cmd.Output()
	if err != nil {
		return getGitHubStatusDirect(expandedPath)
	}
	
	var result struct {
		Issues int `json:"issues"`
		PRs    int `json:"prs"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return getGitHubStatusDirect(expandedPath)
	}
	
	return &GitHubStatus{
		Issues: result.Issues,
		PRs:    result.PRs,
	}, nil
}

// getGitHubStatusDirect is a fallback using gh directly
func getGitHubStatusDirect(expandedPath string) (*GitHubStatus, error) {
	status := &GitHubStatus{}
	
	cmd := exec.Command("gh", "issue", "list", "--state", "open", "--json", "number", "-q", "length")
	cmd.Dir = expandedPath
	output, err := cmd.Output()
	if err == nil {
		var count int
		json.Unmarshal(output, &count)
		status.Issues = count
	}
	
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

// GetVercelStatus returns the latest deployment status using mc-vl-status script
func GetVercelStatus(projectPath string) (string, error) {
	expandedPath := expandPath(projectPath)
	
	// Check if it's a Vercel project
	vercelDir := filepath.Join(expandedPath, ".vercel")
	if _, err := os.Stat(vercelDir); os.IsNotExist(err) {
		return "", nil
	}
	
	// Use mc-vl-status script
	home, _ := os.UserHomeDir()
	binPath := filepath.Join(home, "Projects", "mission-control", "bin", "mc-vl-status")
	
	cmd := exec.Command(binPath, expandedPath, "--json")
	output, err := cmd.Output()
	if err != nil {
		return getVercelStatusDirect(expandedPath)
	}
	
	var result struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return getVercelStatusDirect(expandedPath)
	}
	
	return result.State, nil
}

// getVercelStatusDirect is a fallback using vercel directly
func getVercelStatusDirect(expandedPath string) (string, error) {
	cmd := exec.Command("vercel", "ls", "--json", "-n", "1")
	cmd.Dir = expandedPath
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}
	
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

// GetPrimaryLanguage uses mc-tokei-lang-perc to detect the primary language
func GetPrimaryLanguage(projectPath string) string {
	expandedPath := expandPath(projectPath)

	// Check cache first (language doesn't change often)
	if cache, err := LoadProjectCache(projectPath); err == nil && cache.Language != "" {
		return cache.Language
	}

	home, _ := os.UserHomeDir()
	binPath := filepath.Join(home, "Projects", "mission-control", "bin", "mc-tokei-lang-perc")

	cmd := exec.Command(binPath, expandedPath)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Output format: "Language: NN%"
	result := strings.TrimSpace(string(output))
	if result == "" || result == "null: null%" {
		return ""
	}

	// Extract just the language name
	var language string
	parts := strings.Split(result, ":")
	if len(parts) > 0 {
		language = strings.TrimSpace(parts[0])
	}

	// Update cache
	if language != "" {
		UpdateProjectCache(projectPath, func(c *ProjectCache) {
			c.Language = language
		})
	}

	return language
}

// GetGitTimes returns the first commit time (project age) and last commit time
func GetGitTimes(projectPath string) (firstCommit, lastCommit time.Time) {
	expandedPath := expandPath(projectPath)

	// Check if it's a git repo
	gitDir := filepath.Join(expandedPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return time.Time{}, time.Time{}
	}

	// Check cache for first commit (doesn't change)
	if cache, err := LoadProjectCache(projectPath); err == nil {
		if cache.FirstCommit > 0 {
			firstCommit = time.Unix(cache.FirstCommit, 0)
		}
		if cache.LastCommit > 0 {
			lastCommit = time.Unix(cache.LastCommit, 0)
		}
		// If we have first commit cached, we might still need fresh last commit
		if cache.FirstCommit > 0 && time.Since(cache.UpdatedAt) < CacheTTL {
			return firstCommit, lastCommit
		}
	}

	// Get first commit time (oldest) - only if not cached
	if firstCommit.IsZero() {
		cmd := exec.Command("git", "-C", expandedPath, "log", "--reverse", "--format=%ct", "-1")
		output, err := cmd.Output()
		if err == nil {
			var ts int64
			if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &ts); err == nil {
				firstCommit = time.Unix(ts, 0)
			}
		}
	}

	// Get last commit time (newest) - always fetch fresh
	cmd := exec.Command("git", "-C", expandedPath, "log", "-1", "--format=%ct")
	output, err := cmd.Output()
	if err == nil {
		var ts int64
		if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &ts); err == nil {
			lastCommit = time.Unix(ts, 0)
		}
	}

	// Update cache
	UpdateProjectCache(projectPath, func(c *ProjectCache) {
		if !firstCommit.IsZero() {
			c.FirstCommit = firstCommit.Unix()
		}
		if !lastCommit.IsZero() {
			c.LastCommit = lastCommit.Unix()
		}
	})

	return firstCommit, lastCommit
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
