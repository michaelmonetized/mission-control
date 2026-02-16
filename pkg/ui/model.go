package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michaelmonetized/mission-control/pkg/discover"
	"github.com/michaelmonetized/mission-control/pkg/openclaw"
)

// =============================================================================
// TYPES
// =============================================================================

type ProjectType string

const (
	TypeVercel    ProjectType = "vercel"
	TypeSwift     ProjectType = "swift"
	TypeGo        ProjectType = "go"
	TypeC         ProjectType = "c"
	TypePython    ProjectType = "python"
	TypeRuby      ProjectType = "ruby"
	TypeRust      ProjectType = "rust"
	TypeLua       ProjectType = "lua"
	TypeHTML      ProjectType = "html"
	TypeCSS       ProjectType = "css"
	TypePHP       ProjectType = "php"
	TypeJava      ProjectType = "java"
	TypeWordPress ProjectType = "wordpress"
	TypeTerminal  ProjectType = "terminal"  // bash/zsh/dotfiles
	TypeChrome    ProjectType = "chrome"    // browser extensions
	TypeDocker    ProjectType = "docker"
	TypeMarkdown  ProjectType = "markdown"
	TypeJSON      ProjectType = "json"
	TypeGit       ProjectType = "git"       // fallback
)

// Project represents a discovered project with all stats
type Project struct {
	Name     string
	Path     string
	Type     ProjectType
	Language string // Primary language detected by tokei

	// Time-based stats
	LastBuildTime time.Time // Last Vercel/Swift build
	FirstCommit   time.Time // Project age
	LastCommit    time.Time // Time since last commit

	// Git status
	Staged    int
	Untracked int
	Modified  int

	// GitHub status
	Issues int
	PRs    int

	// Vercel status
	VercelState string // ready, building, queued, failed

	// Swift status
	SwiftClean  int
	SwiftFailed int

	// Running state
	Running bool
}

// Stats holds aggregate counts for the status bar
type Stats struct {
	// Vercel
	VercelReady    int
	VercelBuilding int
	VercelQueued   int
	VercelFailed   int

	// Swift
	SwiftClean  int
	SwiftFailed int

	// Git
	TotalStaged    int
	TotalUntracked int
	TotalModified  int

	// GitHub
	TotalIssues int
	TotalPRs    int

	TotalProjects int
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

// =============================================================================
// ASYNC MESSAGES
// =============================================================================

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

type gitTimesMsg struct {
	name        string
	firstCommit time.Time
	lastCommit  time.Time
}

type languageMsg struct {
	name     string
	language string
}

type chatResponseMsg struct {
	response string
	err      error
}

// =============================================================================
// MODEL
// =============================================================================

// ButtonAction represents a clickable action
type ButtonAction int

const (
	ActionNone ButtonAction = iota
	ActionPush
	ActionMerge
	ActionRun
	ActionDeploy
	ActionReadme
	ActionRoadmap
	ActionPlan
	ActionTodo
	ActionChat
)

// ButtonBounds tracks clickable button regions
type ButtonBounds struct {
	StartX int
	EndX   int
	Action ButtonAction
	Row    int // which project row (relative to scroll)
}

type Model struct {
	projects []Project
	filtered []Project
	stats    Stats

	selectedIdx  int
	scrollOffset int
	viewMode     ViewMode

	currentProject *Project

	searchInput textinput.Model
	chatInput   textinput.Model
	chatCwd     string // ~/Projects or selected project path

	width  int
	height int

	// Vim motion accumulator
	motionNum string

	// Loading state
	loading       bool
	statusLoading sync.Map

	// OpenClaw
	clawClient   *openclaw.Client
	chatResponse string
	chatLoading  bool
	chatError    string

	// Clickable buttons
	buttonBounds []ButtonBounds
	listStartY   int // Y offset where project list starts
}

// =============================================================================
// INITIALIZATION
// =============================================================================

func NewModel() Model {
	search := textinput.New()
	search.Placeholder = "type / to search"
	search.CharLimit = 50

	chat := textinput.New()
	chat.Placeholder = "type C to chat in ~/Projects c to chat in selected project"
	chat.CharLimit = 500

	clawClient, _ := openclaw.NewClientFromConfig()

	homeDir, _ := os.UserHomeDir()

	return Model{
		projects:    []Project{},
		filtered:    []Project{},
		searchInput: search,
		chatInput:   chat,
		chatCwd:     filepath.Join(homeDir, "Projects"),
		viewMode:    ListView,
		loading:     true,
		clawClient:  clawClient,
	}
}

func (m Model) Init() tea.Cmd {
	return loadProjectsCmd
}

// =============================================================================
// ASYNC COMMANDS
// =============================================================================

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
			pType = TypeGit
		}
		projects = append(projects, Project{
			Name: d.Name,
			Path: d.Path,
			Type: pType,
		})
	}

	return projectsLoadedMsg(projects)
}

func loadGitStatusCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		status, _ := discover.GetGitStatus(path)
		return gitStatusMsg{name: name, status: status}
	}
}

func loadGHStatusCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		status, _ := discover.GetGitHubStatus(path)
		return ghStatusMsg{name: name, status: status}
	}
}

func loadVercelStatusCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		state, _ := discover.GetVercelStatus(path)
		return vercelStatusMsg{name: name, state: state}
	}
}

func loadGitTimesCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		first, last := discover.GetGitTimes(path)
		return gitTimesMsg{name: name, firstCommit: first, lastCommit: last}
	}
}

func loadLanguageCmd(name, path string) tea.Cmd {
	return func() tea.Msg {
		lang := discover.GetPrimaryLanguage(path)
		return languageMsg{name: name, language: lang}
	}
}

func sendChatCmd(client *openclaw.Client, message, cwd string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return chatResponseMsg{err: fmt.Errorf("OpenClaw not connected")}
		}
		response, err := client.SendMessageSync(message, cwd)
		return chatResponseMsg{response: response, err: err}
	}
}

// =============================================================================
// UPDATE
// =============================================================================

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case projectsLoadedMsg:
		m.projects = []Project(msg)
		m.filtered = m.projects
		m.loading = false
		m.stats.TotalProjects = len(m.projects)

		// Start loading stats incrementally (non-blocking)
		var cmds []tea.Cmd
		for _, p := range m.projects {
			cmds = append(cmds, loadGitStatusCmd(p.Name, p.Path))
			cmds = append(cmds, loadGitTimesCmd(p.Name, p.Path))
			cmds = append(cmds, loadLanguageCmd(p.Name, p.Path))
			if p.Type == TypeVercel {
				cmds = append(cmds, loadVercelStatusCmd(p.Name, p.Path))
			}
			cmds = append(cmds, loadGHStatusCmd(p.Name, p.Path))
		}
		return m, tea.Batch(cmds...)

	case gitStatusMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name && msg.status != nil {
				m.projects[i].Staged = msg.status.Staged
				m.projects[i].Untracked = msg.status.Untracked
				m.projects[i].Modified = msg.status.Modified
				break
			}
		}
		m.updateStats()
		m.syncFiltered()
		return m, nil

	case ghStatusMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name && msg.status != nil {
				m.projects[i].Issues = msg.status.Issues
				m.projects[i].PRs = msg.status.PRs
				break
			}
		}
		m.updateStats()
		return m, nil

	case vercelStatusMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name {
				m.projects[i].VercelState = msg.state
				break
			}
		}
		m.updateStats()
		m.syncFiltered()
		return m, nil

	case gitTimesMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name {
				m.projects[i].FirstCommit = msg.firstCommit
				m.projects[i].LastCommit = msg.lastCommit
				break
			}
		}
		m.syncFiltered()
		return m, nil

	case languageMsg:
		for i := range m.projects {
			if m.projects[i].Name == msg.name {
				m.projects[i].Language = msg.language
				m.projects[i].Type = detectProjectType(m.projects[i])
				break
			}
		}
		m.syncFiltered()
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

func (m *Model) updateStats() {
	var s Stats
	s.TotalProjects = len(m.projects)

	for _, p := range m.projects {
		s.TotalStaged += p.Staged
		s.TotalUntracked += p.Untracked
		s.TotalModified += p.Modified
		s.TotalIssues += p.Issues
		s.TotalPRs += p.PRs
		s.SwiftClean += p.SwiftClean
		s.SwiftFailed += p.SwiftFailed

		switch p.VercelState {
		case "ready":
			s.VercelReady++
		case "building":
			s.VercelBuilding++
		case "queued":
			s.VercelQueued++
		case "failed":
			s.VercelFailed++
		}
	}

	m.stats = s
}

func (m *Model) syncFiltered() {
	// Re-sync filtered with updated project data
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
}

// detectProjectType determines project type from language, path, and markers
func detectProjectType(p Project) ProjectType {
	name := strings.ToLower(p.Name)
	lang := strings.ToLower(p.Language)

	// Check for specific project markers first
	expandedPath := expandPath(p.Path)

	// Vercel project
	if _, err := os.Stat(filepath.Join(expandedPath, ".vercel")); err == nil {
		return TypeVercel
	}

	// Swift project
	if _, err := os.Stat(filepath.Join(expandedPath, "Package.swift")); err == nil {
		return TypeSwift
	}

	// WordPress
	if strings.Contains(name, "wordpress") || strings.Contains(name, "wp-") {
		return TypeWordPress
	}
	if _, err := os.Stat(filepath.Join(expandedPath, "wp-config.php")); err == nil {
		return TypeWordPress
	}

	// Browser extension
	if strings.Contains(name, "extension") || strings.Contains(name, "chrome") {
		return TypeChrome
	}
	if _, err := os.Stat(filepath.Join(expandedPath, "manifest.json")); err == nil {
		// Check if it looks like a browser extension manifest
		return TypeChrome
	}

	// Dotfiles / terminal
	if name == "dotfiles" || strings.HasPrefix(name, ".") || strings.Contains(name, "zsh") || strings.Contains(name, "bash") {
		return TypeTerminal
	}

	// Docker
	if _, err := os.Stat(filepath.Join(expandedPath, "Dockerfile")); err == nil {
		return TypeDocker
	}

	// Language-based detection from tokei
	switch {
	case strings.Contains(lang, "go"):
		return TypeGo
	case strings.Contains(lang, "c") && !strings.Contains(lang, "css"):
		if lang == "c" || strings.HasPrefix(lang, "c ") {
			return TypeC
		}
	case strings.Contains(lang, "python"):
		return TypePython
	case strings.Contains(lang, "ruby"):
		return TypeRuby
	case strings.Contains(lang, "rust"):
		return TypeRust
	case strings.Contains(lang, "lua"):
		return TypeLua
	case strings.Contains(lang, "html"):
		return TypeHTML
	case strings.Contains(lang, "css"):
		return TypeCSS
	case strings.Contains(lang, "php"):
		return TypePHP
	case strings.Contains(lang, "java") && !strings.Contains(lang, "javascript"):
		return TypeJava
	case strings.Contains(lang, "markdown"):
		return TypeMarkdown
	case strings.Contains(lang, "json"):
		return TypeJSON
	case strings.Contains(lang, "tsx"), strings.Contains(lang, "typescript"), strings.Contains(lang, "javascript"):
		// TSX/TS/JS projects without .vercel are still web projects
		return TypeVercel
	}

	return TypeGit // fallback
}

// =============================================================================
// KEY HANDLING
// =============================================================================

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
			m.chatResponse = ""
			m.chatError = ""
		}
		return m, nil
	}

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

	// Vim motion number prefix
	if key >= "0" && key <= "9" && (m.motionNum != "" || key != "0") {
		m.motionNum += key
		return m, nil
	}

	count := 1
	if m.motionNum != "" {
		fmt.Sscanf(m.motionNum, "%d", &count)
		m.motionNum = ""
	}

	listHeight := m.getListHeight()

	switch key {
	case "j", "down":
		m.selectedIdx = min(m.selectedIdx+count, len(m.filtered)-1)
		m.ensureVisible(listHeight)
	case "k", "up":
		m.selectedIdx = maxInt(m.selectedIdx-count, 0)
		m.ensureVisible(listHeight)
	case "g":
		m.selectedIdx = 0
		m.scrollOffset = 0
	case "G":
		m.selectedIdx = len(m.filtered) - 1
		m.ensureVisible(listHeight)
	case "ctrl+d":
		m.selectedIdx = min(m.selectedIdx+listHeight/2, len(m.filtered)-1)
		m.ensureVisible(listHeight)
	case "ctrl+u":
		m.selectedIdx = maxInt(m.selectedIdx-listHeight/2, 0)
		m.ensureVisible(listHeight)
	case "/":
		m.viewMode = SearchMode
		m.searchInput.Focus()
		return m, textinput.Blink
	case "C":
		// Chat in ~/Projects
		homeDir, _ := os.UserHomeDir()
		m.chatCwd = filepath.Join(homeDir, "Projects")
		m.viewMode = ChatMode
		m.chatInput.Focus()
		return m, textinput.Blink
	case "c":
		// Chat in selected project
		if len(m.filtered) > 0 {
			m.chatCwd = expandPath(m.filtered[m.selectedIdx].Path)
		}
		m.viewMode = ChatMode
		m.chatInput.Focus()
		return m, textinput.Blink
	case "enter":
		if len(m.filtered) > 0 {
			m.currentProject = &m.filtered[m.selectedIdx]
			m.viewMode = DetailView
		}
	case "o":
		if len(m.filtered) > 0 {
			return m, openInEditorCmd(m.filtered[m.selectedIdx].Path, "")
		}
	case "r":
		if len(m.filtered) > 0 {
			return m, openInEditorCmd(m.filtered[m.selectedIdx].Path, "README.md")
		}
	case "R":
		if len(m.filtered) > 0 {
			return m, openInEditorCmd(m.filtered[m.selectedIdx].Path, "ROADMAP.md")
		}
	case "p":
		if len(m.filtered) > 0 {
			return m, openInEditorCmd(m.filtered[m.selectedIdx].Path, "PLAN.md")
		}
	case "t":
		if len(m.filtered) > 0 {
			return m, openInEditorCmd(m.filtered[m.selectedIdx].Path, "TODO.md")
		}
	case "l":
		if len(m.filtered) > 0 {
			return m, openLazygitCmd(m.filtered[m.selectedIdx].Path)
		}
	case "d":
		if len(m.filtered) > 0 {
			p := m.filtered[m.selectedIdx]
			if p.Type == TypeVercel {
				return m, openProductionCmd(p.Name)
			}
		}
	case "?":
		m.viewMode = HelpMode
	case "ctrl+r":
		m.loading = true
		return m, loadProjectsCmd
	}

	return m, nil
}

func (m *Model) ensureVisible(listHeight int) {
	if m.selectedIdx < m.scrollOffset {
		m.scrollOffset = m.selectedIdx
	} else if m.selectedIdx >= m.scrollOffset+listHeight {
		m.scrollOffset = m.selectedIdx - listHeight + 1
	}
}

func (m *Model) getListHeight() int {
	// Total height minus: top status (1) + search box (3) + chat box (3) + bottom status (1)
	return maxInt(m.height-8, 5)
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
	m.scrollOffset = 0

	return m, cmd
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Only handle left clicks
	if msg.Type != tea.MouseLeft {
		return m, nil
	}

	// Check if click is in project list area
	// List starts at Y=3 (after top status + search box)
	listStartY := 3
	listHeight := m.getListHeight()

	if msg.Y >= listStartY && msg.Y < listStartY+listHeight {
		// Calculate which row was clicked
		clickedRow := msg.Y - listStartY
		projectIdx := m.scrollOffset + clickedRow

		if projectIdx < len(m.filtered) {
			// Check if click is on an action button
			for _, btn := range m.buttonBounds {
				if btn.Row == clickedRow && msg.X >= btn.StartX && msg.X < btn.EndX {
					p := m.filtered[projectIdx]
					return m.executeAction(btn.Action, p)
				}
			}

			// Otherwise, select the row
			m.selectedIdx = projectIdx
		}
	}

	return m, nil
}

func (m Model) executeAction(action ButtonAction, p Project) (tea.Model, tea.Cmd) {
	expandedPath := expandPath(p.Path)
	home, _ := os.UserHomeDir()
	binDir := filepath.Join(home, "Projects", "mission-control", "bin")

	switch action {
	case ActionPush:
		return m, runScriptCmd(filepath.Join(binDir, "mc-push"), expandedPath)

	case ActionMerge:
		return m, runScriptCmd(filepath.Join(binDir, "mc-merge"), expandedPath)

	case ActionRun:
		return m, runScriptCmd(filepath.Join(binDir, "mc-run"), expandedPath)

	case ActionDeploy:
		return m, runScriptCmd(filepath.Join(binDir, "mc-deploy"), expandedPath)

	case ActionReadme:
		return m, runScriptCmd(filepath.Join(binDir, "mc-edit"), expandedPath, "README.md")

	case ActionRoadmap:
		return m, runScriptCmd(filepath.Join(binDir, "mc-edit"), expandedPath, "ROADMAP.md")

	case ActionPlan:
		return m, runScriptCmd(filepath.Join(binDir, "mc-edit"), expandedPath, "PLAN.md")

	case ActionTodo:
		return m, runScriptCmd(filepath.Join(binDir, "mc-edit"), expandedPath, "TODO.md")

	case ActionChat:
		return m, runScriptCmd(filepath.Join(binDir, "mc-chat"), expandedPath)
	}

	return m, nil
}

// runScriptCmd runs a shell script without blocking the TUI
func runScriptCmd(script string, args ...string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(script, args...)
		cmd.Start() // Don't wait
		return nil
	}
}

func (m Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		message := m.chatInput.Value()
		if message == "" {
			return m, nil
		}

		m.chatInput.SetValue("")
		m.chatLoading = true
		m.chatResponse = ""
		m.chatError = ""

		return m, sendChatCmd(m.clawClient, message, m.chatCwd)
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

// =============================================================================
// VIEW
// =============================================================================

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.loading {
		return fmt.Sprintf("\n  %s Mission Control - Discovering projects...\n", IconRocket)
	}

	var b strings.Builder

	// Top status line
	b.WriteString(m.renderTopStatus())
	b.WriteString("\n")

	// Search box (rounded)
	b.WriteString(m.renderSearchBox())
	b.WriteString("\n")

	// Project list with scrollbar
	listHeight := m.getListHeight()
	b.WriteString(m.renderProjectList(listHeight))

	// Chat box (rounded)
	b.WriteString(m.renderChatBox())
	b.WriteString("\n")

	// Bottom status line
	b.WriteString(m.renderBottomStatus())

	return b.String()
}

// =============================================================================
// TOP STATUS LINE (Powerline style)
// =============================================================================

func (m Model) renderTopStatus() string {
	// Title segment: mint
	title := fmt.Sprintf(" %s Mission Control ", IconRocket)
	titleSeg := lipgloss.NewStyle().Foreground(ColorBlack).Background(ColorMint).Render(title)
	titleCapL := lipgloss.NewStyle().Foreground(ColorMint).Render(PLLeftHalfCircle)
	titleCapR := lipgloss.NewStyle().Foreground(ColorMint).Render(PLLowerLeftTriangle)

	// Vercel segment: yellow
	vercel := fmt.Sprintf(" %s %d%s %d%s %d%s %d%s ",
		IconVercel,
		m.stats.VercelReady, IconReady,
		m.stats.VercelBuilding, IconBuilding,
		m.stats.VercelQueued, IconQueued,
		m.stats.VercelFailed, IconX)
	vercelSeg := lipgloss.NewStyle().Foreground(ColorBlack).Background(ColorVercel).Render(vercel)
	vercelCapL := lipgloss.NewStyle().Foreground(ColorVercel).Render(PLUpperRightTriangle)
	vercelCapR := lipgloss.NewStyle().Foreground(ColorVercel).Render(PLLowerLeftTriangle)

	// Swift segment: magenta
	swift := fmt.Sprintf(" %s %d%s %d%s ",
		IconSwift,
		m.stats.SwiftClean, IconCheck,
		m.stats.SwiftFailed, IconX)
	swiftSeg := lipgloss.NewStyle().Foreground(ColorBlack).Background(ColorSwift).Render(swift)
	swiftCapL := lipgloss.NewStyle().Foreground(ColorSwift).Render(PLUpperRightTriangle)
	swiftCapR := lipgloss.NewStyle().Foreground(ColorSwift).Render(PLFlameThick)

	// Calculate elastic gap
	leftPart := titleCapL + titleSeg + titleCapR + vercelCapL + vercelSeg + vercelCapR + swiftCapL + swiftSeg + swiftCapR
	leftLen := lipgloss.Width(leftPart)

	// Git segment: cyan
	git := fmt.Sprintf(" %s %s%d %s%d %s%d ",
		IconGit,
		IconStaged, m.stats.TotalStaged,
		IconUntracked, m.stats.TotalUntracked,
		IconModified, m.stats.TotalModified)
	gitSeg := lipgloss.NewStyle().Foreground(ColorBlack).Background(ColorGit).Render(git)
	gitCapL := lipgloss.NewStyle().Foreground(ColorGit).Render(PLFlameThickMirrored)
	gitCapR := lipgloss.NewStyle().Foreground(ColorGit).Render(PLRightHardDivider)

	// GitHub segment: green
	gh := fmt.Sprintf(" %s %s%d %s%d ",
		IconGitHub,
		IconIssue, m.stats.TotalIssues,
		IconPR, m.stats.TotalPRs)
	ghSeg := lipgloss.NewStyle().Foreground(ColorBlack).Background(ColorGH).Render(gh)
	ghCapL := lipgloss.NewStyle().Foreground(ColorGH).Render(PLLeftHardDivider)
	ghCapR := lipgloss.NewStyle().Foreground(ColorGH).Render(PLRightHalfCircle)

	rightPart := gitCapL + gitSeg + gitCapR + ghCapL + ghSeg + ghCapR
	rightLen := lipgloss.Width(rightPart)

	// Elastic gap
	gap := m.width - leftLen - rightLen
	if gap < 0 {
		gap = 0
	}

	return leftPart + strings.Repeat(" ", gap) + rightPart
}

// =============================================================================
// SEARCH BOX (Rounded)
// =============================================================================

func (m Model) renderSearchBox() string {
	content := fmt.Sprintf("%s %s", IconSearch, m.searchInput.View())
	if m.viewMode != SearchMode {
		content = fmt.Sprintf("%s %s", IconSearch, m.searchInput.Placeholder)
	}

	box := SearchBoxStyle.Width(m.width - 4).Render(content)
	return box
}

// =============================================================================
// PROJECT LIST (Striped with scrollbar)
// =============================================================================

func (m *Model) renderProjectList(height int) string {
	if m.viewMode == HelpMode {
		return m.renderHelp(height)
	}
	if m.viewMode == DetailView {
		return m.renderDetailView(height)
	}

	var rows []string
	listWidth := m.width - 3 // Leave room for scrollbar

	// Clear button bounds for fresh calculation
	m.buttonBounds = nil

	for i := m.scrollOffset; i < len(m.filtered) && i < m.scrollOffset+height; i++ {
		p := m.filtered[i]
		isSelected := i == m.selectedIdx
		isOdd := (i-m.scrollOffset)%2 == 1
		rowNum := i - m.scrollOffset

		row := m.renderProjectRow(p, i, listWidth, isOdd, isSelected, rowNum)
		rows = append(rows, row)
	}

	// Pad remaining height
	for i := len(rows); i < height; i++ {
		rows = append(rows, strings.Repeat(" ", listWidth))
	}

	// Add scrollbar
	scrollbar := RenderScrollbar(m.scrollOffset, len(m.filtered), height)
	scrollLines := strings.Split(scrollbar, "\n")

	var result strings.Builder
	for i, row := range rows {
		sb := " "
		if i < len(scrollLines) {
			sb = scrollLines[i]
		}
		result.WriteString(row + " " + sb + "\n")
	}

	return result.String()
}

func (m *Model) renderProjectRow(p Project, idx int, width int, isOdd bool, isSelected bool, rowNum int) string {
	// Type icon based on detected language/type
	typeIcon := getTypeIcon(p.Type)

	// Time formatting with icons
	projectAge := formatTimeSince(p.FirstCommit)
	lastCommit := formatTimeSince(p.LastCommit)

	// Build content
	seg1 := fmt.Sprintf("%s %-18s", typeIcon, truncate(p.Name, 18))
	seg2 := fmt.Sprintf(" %s%4s %s%4s ", IconCommitStart, projectAge, IconCommitEnd, lastCommit)
	seg3 := fmt.Sprintf(" %s%-2d %s%-2d %s%-2d ", IconStaged, p.Staged, IconUntracked, p.Untracked, IconModified, p.Modified)
	seg4 := fmt.Sprintf(" %s%-2d %s%-2d", IconIssue, p.Issues, IconPR, p.PRs)
	
	// Action buttons - track positions for click handling
	buttonIcons := []struct {
		icon   string
		action ButtonAction
	}{
		{IconPush, ActionPush},
		{IconMerge, ActionMerge},
		{IconPlayPause, ActionRun},
		{IconDeploy, ActionDeploy},
		{IconReadme, ActionReadme},
		{IconRoadmap, ActionRoadmap},
		{IconPlan, ActionPlan},
		{IconTodo, ActionTodo},
		{IconChat, ActionChat},
	}

	// Build actions string
	var actionsBuilder strings.Builder
	actionsBuilder.WriteString(" ")
	for i, btn := range buttonIcons {
		actionsBuilder.WriteString(btn.icon)
		if i < len(buttonIcons)-1 {
			actionsBuilder.WriteString(" ")
		}
	}
	actions := actionsBuilder.String()

	// Combine content
	content := seg1 + seg2 + seg3 + seg4
	contentWidth := lipgloss.Width(content)
	actionsWidth := lipgloss.Width(actions)
	
	// Calculate gap for elastic spacing
	gap := width - contentWidth - actionsWidth
	if gap < 0 {
		gap = 0
	}

	// Calculate button X positions (after gap)
	buttonsStartX := contentWidth + gap + 1 // +1 for leading space
	currentX := buttonsStartX
	
	for _, btn := range buttonIcons {
		iconWidth := lipgloss.Width(btn.icon)
		m.buttonBounds = append(m.buttonBounds, ButtonBounds{
			StartX: currentX,
			EndX:   currentX + iconWidth,
			Action: btn.action,
			Row:    rowNum,
		})
		currentX += iconWidth + 1 // +1 for space between icons
	}

	// Build full row with padding to exact width
	fullRow := content + strings.Repeat(" ", gap) + actions
	currentWidth := lipgloss.Width(fullRow)
	if currentWidth < width {
		fullRow += strings.Repeat(" ", width-currentWidth)
	}

	// Apply ANSI background color directly (bypassing lipgloss to avoid icon issues)
	// Very subtle striping: no bg (even) vs 233 (odd) - barely visible
	if isSelected {
		return fmt.Sprintf("\033[30;48;5;6m%s\033[0m", fullRow) // black on cyan
	} else if isOdd {
		return fmt.Sprintf("\033[48;5;233m%s\033[0m", fullRow) // very dark gray
	}
	// Even rows: no background (terminal default)
	return fullRow
}

// getTypeIcon returns the appropriate icon for a project type
func getTypeIcon(t ProjectType) string {
	switch t {
	case TypeVercel:
		return IconVercel
	case TypeSwift:
		return IconSwift
	case TypeGo:
		return IconTypeGo
	case TypeC:
		return IconTypeC
	case TypePython:
		return IconTypePython
	case TypeRuby:
		return IconTypeRuby
	case TypeRust:
		return IconTypeRust
	case TypeLua:
		return IconTypeLua
	case TypeHTML:
		return IconTypeHTML
	case TypeCSS:
		return IconTypeCss
	case TypePHP:
		return IconTypePhp
	case TypeJava:
		return IconTypeJava
	case TypeWordPress:
		return IconTypeWordPress
	case TypeTerminal:
		return IconTypeTerminal
	case TypeChrome:
		return IconTypeChrome
	case TypeDocker:
		return IconTypeDocker
	case TypeMarkdown:
		return IconTypeMarkdown
	case TypeJSON:
		return IconTypeJson
	default:
		return IconTypeDefault
	}
}

func formatTimeSince(t time.Time) string {
	if t.IsZero() {
		return "  - "
	}

	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%2ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%2dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%2dh", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%2dd", int(d.Hours()/24))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%2dw", int(d.Hours()/(24*7)))
	}
	if d < 365*24*time.Hour {
		return fmt.Sprintf("%2dM", int(d.Hours()/(24*30)))
	}
	return fmt.Sprintf("%2dy", int(d.Hours()/(24*365)))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// =============================================================================
// CHAT BOX (Rounded)
// =============================================================================

func (m Model) renderChatBox() string {
	var content string

	if m.chatLoading {
		content = fmt.Sprintf("%s Thinking...", IconBrain)
	} else if m.chatError != "" {
		content = fmt.Sprintf("%s %s", IconX, m.chatError)
	} else if m.chatResponse != "" {
		resp := strings.ReplaceAll(m.chatResponse, "\n", " ")
		if len(resp) > m.width-10 {
			resp = resp[:m.width-13] + "..."
		}
		content = fmt.Sprintf("%s %s", IconChat, resp)
	} else if m.viewMode == ChatMode {
		content = fmt.Sprintf("%s %s", IconChat, m.chatInput.View())
	} else {
		cwdDisplay := "~/Projects"
		if m.chatCwd != "" && !strings.HasSuffix(m.chatCwd, "/Projects") {
			cwdDisplay = filepath.Base(m.chatCwd)
		}
		content = fmt.Sprintf("%s type C to chat in ~/Projects c to chat in %s", IconChat, cwdDisplay)
	}

	box := ChatBoxStyle.Width(m.width - 4).Render(content)
	return box
}

// =============================================================================
// BOTTOM STATUS LINE
// =============================================================================

func (m Model) renderBottomStatus() string {
	// Left side: project count + add
	left := fmt.Sprintf("%s %d  %s",
		IconProjects, m.stats.TotalProjects, IconPlus)

	// Right side: OpenClaw status + model + thinking + tokens
	connected := IconConnected
	if m.clawClient == nil {
		connected = IconX
	}

	// TODO: Get real values from OpenClaw client
	agent := "main:main"
	model := "anthropic/claude-sonnet-4"
	thinking := "high"
	tokens := "35k/200k (18%)"

	right := fmt.Sprintf("%s %s  %s  %s %s  %s %s",
		connected, agent, model,
		IconBrain, thinking, IconCoins, tokens)

	// Elastic gap
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 1
	}

	return BottomStatusStyle.Render(left) + strings.Repeat(" ", gap) + BottomStatusStyle.Render(right)
}

// =============================================================================
// HELP VIEW
// =============================================================================

func (m Model) renderHelp(height int) string {
	help := `
  Mission Control - Keyboard Shortcuts

  Navigation
    j/k        Move down/up
    g/G        Go to top/bottom
    Ctrl+d/u   Page down/up
    /          Search projects
    Enter      Select project

  Actions
    o          Open project in nvim
    l          Open lazygit
    d          Open production URL (Vercel)

  Files
    r          Edit README.md
    R          Edit ROADMAP.md
    p          Edit PLAN.md
    t          Edit TODO.md

  Chat
    C          Chat in ~/Projects
    c          Chat in selected project

  Other
    Ctrl+r     Refresh all
    ?          Show this help
    q/Esc      Back/Quit
`
	return help
}

// =============================================================================
// DETAIL VIEW
// =============================================================================

func (m Model) renderDetailView(height int) string {
	if m.currentProject == nil {
		return "No project selected\n\nPress 'q' or 'esc' to go back"
	}

	p := m.currentProject
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\n  Project: %s\n", p.Name))
	b.WriteString(fmt.Sprintf("  Path: %s\n", p.Path))
	b.WriteString(fmt.Sprintf("  Type: %s\n", p.Type))
	b.WriteString(fmt.Sprintf("  State: %s\n", p.VercelState))
	b.WriteString(fmt.Sprintf("\n  Git: %d staged, %d untracked, %d modified\n", p.Staged, p.Untracked, p.Modified))
	b.WriteString(fmt.Sprintf("  GitHub: %d issues, %d PRs\n", p.Issues, p.PRs))
	b.WriteString("\n  Press 'q' or 'esc' to go back\n")

	return b.String()
}

// =============================================================================
// EXTERNAL COMMANDS
// =============================================================================

func openInEditorCmd(projectPath, file string) tea.Cmd {
	return tea.ExecProcess(
		func() *exec.Cmd {
			expanded := expandPath(projectPath)
			if file != "" {
				return exec.Command("nvim", filepath.Join(expanded, file))
			}
			cmd := exec.Command("nvim", ".")
			cmd.Dir = expanded
			return cmd
		}(),
		nil,
	)
}

func openLazygitCmd(projectPath string) tea.Cmd {
	return tea.ExecProcess(
		func() *exec.Cmd {
			expanded := expandPath(projectPath)
			cmd := exec.Command("lazygit")
			cmd.Dir = expanded
			return cmd
		}(),
		nil,
	)
}

func openProductionCmd(projectName string) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("open", fmt.Sprintf("https://%s", projectName)),
		nil,
	)
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
