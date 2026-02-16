package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michaelmonetized/mission-control/pkg/discover"
	"github.com/michaelmonetized/mission-control/pkg/openclaw"
)

// Project types
type ProjectType string

const (
	TypeVercel ProjectType = "vercel"
	TypeSwift  ProjectType = "swift"
	TypeCLI    ProjectType = "cli"
)

// Project represents a discovered project
type Project struct {
	Name     string
	Path     string
	Type     ProjectType
	Running  bool
	// Git status
	Untracked int
	Modified  int
	Files     int
	// GitHub status
	Issues int
	PRs    int
	// Deploy status
	State string // ready, building, queued, failed
	URL   string
}

// Stats holds aggregate counts
type Stats struct {
	VercelReady    int
	VercelBuilding int
	VercelQueued   int
	VercelFailed   int
	SwiftSuccess   int
	SwiftFailed    int
	TotalFiles     int
	TotalUntracked int
	TotalModified  int
	TotalIssues    int
	TotalPRs       int
	TotalProjects  int
}

// Messages for async loading
type projectsLoadedMsg []Project
type gitStatusMsg struct {
	name   string
	status *discover.GitStatus
}
type ghStatusMsg struct {
	name   string
	status *discover.GitHubStatus
}
type vercelStatusMsg struct {
	name  string
	state string
}

// ViewMode determines current view
type ViewMode int

const (
	ListView ViewMode = iota
	DetailView
	SearchMode
	ChatMode
	HelpMode
)

// Model is the main application state
type Model struct {
	projects      []Project
	filtered      []Project
	stats         Stats
	selectedIdx   int
	viewMode      ViewMode
	currentProject *Project
	
	searchInput   textinput.Model
	chatInput     textinput.Model
	
	width         int
	height        int
	
	// Vim motion accumulator
	motionNum     string
	
	// Loading state
	loading       bool
	statusLoading sync.Map
	
	// OpenClaw chat
	clawClient    *openclaw.Client
	chatResponse  string
	chatLoading   bool
	chatError     string
}

// Chat messages for async response
type chatResponseMsg struct {
	response string
	err      error
}

// NewModel creates a new application model
func NewModel() Model {
	search := textinput.New()
	search.Placeholder = "Search projects..."
	search.CharLimit = 50
	
	chat := textinput.New()
	chat.Placeholder = "Ask OpenClaw..."
	chat.CharLimit = 500
	
	// Try to connect to OpenClaw gateway
	clawClient, _ := openclaw.NewClientFromConfig()
	
	return Model{
		projects:    []Project{},
		filtered:    []Project{},
		searchInput: search,
		chatInput:   chat,
		viewMode:    ListView,
		loading:     true,
		clawClient:  clawClient,
	}
}

// loadProjectsCmd loads projects asynchronously
func loadProjectsCmd() tea.Msg {
	discovered, err := discover.LoadProjects()
	if err != nil {
		return projectsLoadedMsg{}
	}
	
	projects := make([]Project, 0, len(discovered))
	for _, d := range discovered {
		var pType ProjectType
		switch d.Type {
		case "vercel":
			pType = TypeVercel
		case "swift":
			pType = TypeSwift
		default:
			pType = TypeCLI
		}
		projects = append(projects, Project{
			Name: d.Name,
			Path: d.Path,
			Type: pType,
		})
	}
	
	return projectsLoadedMsg(projects)
}

// loadGitStatusCmd loads git status for a project
func loadGitStatusCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		status, _ := discover.GetGitStatus(path)
		return gitStatusMsg{name: name, status: status}
	}
}

// loadGHStatusCmd loads GitHub status for a project
func loadGHStatusCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		status, _ := discover.GetGitHubStatus(path)
		return ghStatusMsg{name: name, status: status}
	}
}

// loadVercelStatusCmd loads Vercel status for a project
func loadVercelStatusCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		state, _ := discover.GetVercelStatus(path)
		return vercelStatusMsg{name: name, state: state}
	}
}

// openInEditorCmd opens a project or file in nvim
func openInEditorCmd(projectPath, file string) tea.Cmd {
	return tea.ExecProcess(
		func() *exec.Cmd {
			expandedPath := expandPath(projectPath)
			if file != "" {
				return exec.Command("nvim", filepath.Join(expandedPath, file))
			}
			cmd := exec.Command("nvim", ".")
			cmd.Dir = expandedPath
			return cmd
		}(),
		nil,
	)
}

// openClawCmd opens OpenClaw TUI in a project
func openClawCmd(projectPath string) tea.Cmd {
	return tea.ExecProcess(
		func() *exec.Cmd {
			expandedPath := expandPath(projectPath)
			cmd := exec.Command("openclaw")
			cmd.Dir = expandedPath
			return cmd
		}(),
		nil,
	)
}

// openProductionCmd opens the production URL
func openProductionCmd(projectName string) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("open", fmt.Sprintf("https://%s", projectName)),
		nil,
	)
}

// openLazygitCmd opens lazygit in a project
func openLazygitCmd(projectPath string) tea.Cmd {
	return tea.ExecProcess(
		func() *exec.Cmd {
			expandedPath := expandPath(projectPath)
			cmd := exec.Command("lazygit")
			cmd.Dir = expandedPath
			return cmd
		}(),
		nil,
	)
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func calculateStats(projects []Project) Stats {
	var s Stats
	for _, p := range projects {
		s.TotalUntracked += p.Untracked
		s.TotalModified += p.Modified
		s.TotalFiles += p.Files
		s.TotalIssues += p.Issues
		s.TotalPRs += p.PRs
		
		switch p.State {
		case "ready":
			s.VercelReady++
		case "building":
			s.VercelBuilding++
		case "queued":
			s.VercelQueued++
		case "failed":
			s.VercelFailed++
		case "success":
			s.SwiftSuccess++
		}
	}
	return s
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return loadProjectsCmd
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case projectsLoadedMsg:
		m.projects = []Project(msg)
		m.filtered = m.projects
		m.loading = false
		m.stats = calculateStats(m.projects)
		m.stats.TotalProjects = len(m.projects)
		
		// Start loading statuses for visible projects
		var cmds []tea.Cmd
		for i, p := range m.projects {
			if i >= 20 { // Limit initial batch
				break
			}
			cmds = append(cmds, loadGitStatusCmd(p.Name, p.Path))
			if p.Type == TypeVercel {
				cmds = append(cmds, loadVercelStatusCmd(p.Name, p.Path))
			}
		}
		return m, tea.Batch(cmds...)
		
	case gitStatusMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name && msg.status != nil {
				m.projects[i].Untracked = msg.status.Untracked
				m.projects[i].Modified = msg.status.Modified
				break
			}
		}
		m.stats = calculateStats(m.projects)
		m.stats.TotalProjects = len(m.projects)
		// Update filtered too
		for i := range m.filtered {
			if m.filtered[i].Name == msg.name && msg.status != nil {
				m.filtered[i].Untracked = msg.status.Untracked
				m.filtered[i].Modified = msg.status.Modified
				break
			}
		}
		return m, nil
		
	case ghStatusMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name && msg.status != nil {
				m.projects[i].Issues = msg.status.Issues
				m.projects[i].PRs = msg.status.PRs
				break
			}
		}
		m.stats = calculateStats(m.projects)
		m.stats.TotalProjects = len(m.projects)
		return m, nil
		
	case vercelStatusMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name {
				m.projects[i].State = msg.state
				break
			}
		}
		m.stats = calculateStats(m.projects)
		m.stats.TotalProjects = len(m.projects)
		for i := range m.filtered {
			if m.filtered[i].Name == msg.name {
				m.filtered[i].State = msg.state
				break
			}
		}
		return m, nil
		
	case chatResponseMsg:
		m.chatLoading = false
		if msg.err != nil {
			m.chatError = msg.err.Error()
		} else {
			m.chatResponse = msg.response
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	
	// Global keys
	switch key {
	case "q", "ctrl+c":
		if m.viewMode == ListView {
			return m, tea.Quit
		}
		m.viewMode = ListView
		return m, nil
	case "esc":
		if m.viewMode != ListView {
			m.viewMode = ListView
			m.searchInput.SetValue("")
			m.chatInput.SetValue("")
			m.filtered = m.projects
		}
		return m, nil
	}
	
	// Mode-specific handling
	switch m.viewMode {
	case SearchMode:
		return m.handleSearchKey(msg)
	case ChatMode:
		return m.handleChatKey(msg)
	default:
		return m.handleListKey(msg)
	}
}

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	
	// Check for number prefix (vim motion)
	if key >= "0" && key <= "9" && (m.motionNum != "" || key != "0") {
		m.motionNum += key
		return m, nil
	}
	
	count := 1
	if m.motionNum != "" {
		fmt.Sscanf(m.motionNum, "%d", &count)
		m.motionNum = ""
	}
	
	switch key {
	case "j", "down":
		m.selectedIdx = min(m.selectedIdx+count, len(m.filtered)-1)
	case "k", "up":
		m.selectedIdx = max(m.selectedIdx-count, 0)
	case "g":
		m.selectedIdx = 0
	case "G":
		m.selectedIdx = len(m.filtered) - 1
	case "ctrl+d":
		m.selectedIdx = min(m.selectedIdx+10, len(m.filtered)-1)
	case "ctrl+u":
		m.selectedIdx = max(m.selectedIdx-10, 0)
	case "/":
		m.viewMode = SearchMode
		m.searchInput.Focus()
		return m, textinput.Blink
	case ">":
		m.viewMode = ChatMode
		m.chatInput.Focus()
		return m, textinput.Blink
	case "enter":
		if len(m.filtered) > 0 {
			m.currentProject = &m.filtered[m.selectedIdx]
			m.viewMode = DetailView
		}
	case "o":
		// Open project in nvim
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			return m, openInEditorCmd(p.Path, "")
		}
	case "r":
		// Edit README
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			return m, openInEditorCmd(p.Path, "README.md")
		}
	case "R":
		// Edit ROADMAP
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			return m, openInEditorCmd(p.Path, "ROADMAP.md")
		}
	case "p":
		// Edit PLAN
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			return m, openInEditorCmd(p.Path, "PLAN.md")
		}
	case "t":
		// Edit TODO
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			return m, openInEditorCmd(p.Path, "TODO.md")
		}
	case "c":
		// Launch OpenClaw TUI in project
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			return m, openClawCmd(p.Path)
		}
	case "d":
		// Open production URL
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			if p.Type == TypeVercel {
				return m, openProductionCmd(p.Name)
			}
		}
	case "l":
		// Lazygit
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			return m, openLazygitCmd(p.Path)
		}
	case "?":
		m.viewMode = HelpMode
	case "ctrl+r":
		// Refresh all
		m.loading = true
		return m, loadProjectsCmd
	}
	
	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.viewMode = ListView
		return m, nil
	case "esc":
		m.viewMode = ListView
		m.searchInput.SetValue("")
		m.filtered = m.projects
		return m, nil
	}
	
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	
	// Filter projects
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filtered = m.projects
	} else {
		m.filtered = nil
		for _, p := range m.projects {
			if strings.Contains(strings.ToLower(p.Name), query) {
				m.filtered = append(m.filtered, p)
			}
		}
	}
	m.selectedIdx = 0
	
	return m, cmd
}

func (m Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		message := m.chatInput.Value()
		if message == "" {
			return m, nil
		}
		
		// Get project context if one is selected
		projectContext := ""
		if len(m.filtered) > 0 && m.selectedIdx < len(m.filtered) {
			p := m.filtered[m.selectedIdx]
			projectContext = fmt.Sprintf("Project: %s (%s) at %s", p.Name, p.Type, p.Path)
		}
		
		m.chatInput.SetValue("")
		m.chatLoading = true
		m.chatResponse = ""
		m.chatError = ""
		
		return m, sendChatCmd(m.clawClient, message, projectContext)
	case "esc":
		m.viewMode = ListView
		m.chatResponse = ""
		m.chatError = ""
		return m, nil
	}
	
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)
	return m, cmd
}

// sendChatCmd sends a message to OpenClaw
func sendChatCmd(client *openclaw.Client, message, projectContext string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return chatResponseMsg{err: fmt.Errorf("OpenClaw not connected")}
		}
		
		response, err := client.SendMessageSync(message, projectContext)
		return chatResponseMsg{response: response, err: err}
	}
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	
	if m.loading {
		return "🚀 Discovering projects..."
	}
	
	var b strings.Builder
	
	// Top status line
	b.WriteString(m.renderTopStatus())
	b.WriteString("\n")
	
	// Search bar
	b.WriteString(m.renderSearchBar())
	b.WriteString("\n")
	
	// Main content
	contentHeight := m.height - 6 // top(1) + search(1) + chat(1) + bottom(1) + padding
	b.WriteString(m.renderContent(contentHeight))
	
	// Chat bar
	b.WriteString(m.renderChatBar())
	b.WriteString("\n")
	
	// Bottom status line
	b.WriteString(m.renderBottomStatus())
	
	return b.String()
}

func (m Model) renderTopStatus() string {
	// Title
	title := "🚀Mission Control"
	if m.viewMode == DetailView && m.currentProject != nil {
		title = fmt.Sprintf("🚀 mc:%s", m.currentProject.Name)
	}
	
	// Vercel segment
	vercel := fmt.Sprintf(" %d%s %d%s %d%s %d%s",
		m.stats.VercelReady, ReadyIcon,
		m.stats.VercelBuilding, BuildingIcon,
		m.stats.VercelQueued, QueuedIcon,
		m.stats.VercelFailed, FailedIcon)
	
	// Swift segment
	swift := fmt.Sprintf(" %d%s %d%s",
		m.stats.SwiftSuccess, SuccessIcon,
		m.stats.SwiftFailed, FailIcon)
	
	// Git segment
	git := fmt.Sprintf(" %d%s %d%s %d%s %d%s",
		m.stats.TotalUntracked, UntrackedIcon,
		m.stats.TotalModified, ModifiedIcon,
		m.stats.TotalIssues, IssuesIcon,
		m.stats.TotalPRs, PRsIcon)
	
	left := TitleStyle.Render(title) + " " +
		VercelStyle.Render(vercel) + " " +
		SwiftStyle.Render(swift) + " " +
		GitStyle.Render(git)
	
	return left
}

func (m Model) renderSearchBar() string {
	prefix := SearchStyle.Render("/")
	if m.viewMode == SearchMode {
		return prefix + " " + m.searchInput.View()
	}
	return prefix + " " + SearchInputStyle.Render(m.searchInput.Placeholder)
}

func (m Model) renderContent(height int) string {
	switch m.viewMode {
	case DetailView:
		return m.renderDetailView(height)
	case HelpMode:
		return m.renderHelp(height)
	default:
		return m.renderProjectList(height)
	}
}

func (m Model) renderHelp(height int) string {
	help := `
🚀 Mission Control - Keyboard Shortcuts

Navigation
  j/k      Move down/up
  g/G      Go to top/bottom
  Ctrl+d/u Page down/up
  /        Search projects
  Enter    Select project

Actions
  o        Open project in nvim
  l        Open lazygit
  d        Open production URL (Vercel)
  c        Launch OpenClaw TUI
  
Files
  r        Edit README.md
  R        Edit ROADMAP.md
  p        Edit PLAN.md
  t        Edit TODO.md

Other
  >        OpenClaw chat
  Ctrl+r   Refresh all
  ?        Show this help
  q/Esc    Back/Quit
`
	return help
}

func (m Model) renderProjectList(height int) string {
	var b strings.Builder
	
	for i, p := range m.filtered {
		if i >= height {
			break
		}
		
		// Play/pause icon
		playIcon := PauseIcon
		if p.Running {
			playIcon = PlayIcon
		}
		
		// Type icon
		var typeIcon string
		switch p.Type {
		case TypeVercel:
			typeIcon = VercelIcon
		case TypeSwift:
			typeIcon = SwiftIcon
		case TypeCLI:
			typeIcon = CLIIcon
		}
		
		// State indicator
		var stateStyle lipgloss.Style
		switch p.State {
		case "ready", "success":
			stateStyle = ReadyStyle
		case "building":
			stateStyle = BuildingStyle
		case "queued":
			stateStyle = QueuedStyle
		case "failed":
			stateStyle = FailedStyle
		default:
			stateStyle = QueuedStyle
		}
		
		// Build row
		row := fmt.Sprintf("[%s] %s %s  %d%s %d%s %d%s %d%s",
			playIcon,
			stateStyle.Render(typeIcon),
			p.Name,
			p.Untracked, UntrackedIcon,
			p.Modified, ModifiedIcon,
			p.Issues, IssuesIcon,
			p.PRs, PRsIcon)
		
		// Add action buttons
		row += fmt.Sprintf("  %s %s %s", ProdIcon, EditorIcon, OpenClawIcon)
		
		// Apply selection style
		if i == m.selectedIdx {
			row = SelectedRowStyle.Render(row)
		} else {
			row = ProjectRowStyle.Render(row)
		}
		
		b.WriteString(row)
		b.WriteString("\n")
	}
	
	// Pad remaining height
	for i := len(m.filtered); i < height; i++ {
		b.WriteString("\n")
	}
	
	return b.String()
}

func (m Model) renderDetailView(height int) string {
	if m.currentProject == nil {
		return "No project selected"
	}
	
	p := m.currentProject
	var b strings.Builder
	
	b.WriteString(fmt.Sprintf("Project: %s\n", p.Name))
	b.WriteString(fmt.Sprintf("Path: %s\n", p.Path))
	b.WriteString(fmt.Sprintf("Type: %s\n", p.Type))
	b.WriteString(fmt.Sprintf("State: %s\n", p.State))
	b.WriteString("\n")
	b.WriteString("Press 'q' or 'esc' to go back\n")
	
	return b.String()
}

func (m Model) renderChatBar() string {
	prefix := ChatPromptStyle.Render(">")
	
	// Show loading indicator
	if m.chatLoading {
		return prefix + " 󰔟 Thinking..."
	}
	
	// Show error
	if m.chatError != "" {
		return prefix + " " + FailedStyle.Render("✗ "+m.chatError)
	}
	
	// Show response (truncate to fit)
	if m.chatResponse != "" {
		response := m.chatResponse
		// Clean up response (remove newlines, truncate)
		response = strings.ReplaceAll(response, "\n", " ")
		maxLen := m.width - 5
		if maxLen > 0 && len(response) > maxLen {
			response = response[:maxLen-3] + "..."
		}
		return prefix + " " + ReadyStyle.Render("󱐏 ") + response
	}
	
	if m.viewMode == ChatMode {
		return prefix + " " + m.chatInput.View()
	}
	return prefix + " " + SearchInputStyle.Render(m.chatInput.Placeholder)
}

func (m Model) renderBottomStatus() string {
	status := fmt.Sprintf(" %d projects  %d%s  %d%s  %d%s  %d%s  %d%s",
		m.stats.TotalProjects,
		m.stats.TotalFiles, FilesIcon,
		m.stats.TotalUntracked, UntrackedIcon,
		m.stats.TotalModified, ModifiedIcon,
		m.stats.TotalIssues, IssuesIcon,
		m.stats.TotalPRs, PRsIcon)
	
	return BottomStatusStyle.Render(status)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
