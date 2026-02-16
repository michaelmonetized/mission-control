package ui

import "github.com/charmbracelet/lipgloss"

// p10k style: black text on colored backgrounds
var (
	// Status segment styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).  // black
			Background(lipgloss.Color("2")).  // green
			Padding(0, 1)

	VercelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).  // black
			Background(lipgloss.Color("3")).  // yellow
			Padding(0, 1)

	SwiftStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).  // black
			Background(lipgloss.Color("5")).  // magenta
			Padding(0, 1)

	GitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).  // black
			Background(lipgloss.Color("6")).  // cyan
			Padding(0, 1)

	// Search bar
	SearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("3")).
			Padding(0, 1)

	SearchInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7"))

	// Project list
	ProjectRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	SelectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("6")).
				Bold(true)

	// State colors
	ReadyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	BuildingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))  // blue
	QueuedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray
	FailedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // red

	// Icons (Nerd Fonts)
	PlayIcon    = "▶"
	PauseIcon   = "󰏤"
	VercelIcon  = "󰐎"
	SwiftIcon   = "󰣪"
	CLIIcon     = ""
	FilesIcon   = ""
	UntrackedIcon = ""
	ModifiedIcon  = ""
	IssuesIcon    = ""
	PRsIcon       = ""
	ReadyIcon     = "◬"
	BuildingIcon  = "󱫟"
	QueuedIcon    = "⨻"
	FailedIcon    = ""
	SuccessIcon   = "󰸞"
	FailIcon      = "✘"
	ProdIcon      = "󰑢"
	EditorIcon    = ""
	RoadmapIcon   = "󱔘"
	OpenClawIcon  = "󱐏"

	// Chat bar
	ChatPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("2")).
			Padding(0, 1)

	// Bottom status
	BottomStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))
)

// Helper to render segment with separator
func Segment(style lipgloss.Style, content string) string {
	return style.Render(content)
}
