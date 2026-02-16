package ui

import "github.com/charmbracelet/lipgloss"

// =============================================================================
// COLOR PALETTE (Catppuccin-inspired)
// =============================================================================

var (
	ColorBlack   = lipgloss.Color("0")
	ColorRed     = lipgloss.Color("1")
	ColorGreen   = lipgloss.Color("2")
	ColorYellow  = lipgloss.Color("3")
	ColorBlue    = lipgloss.Color("4")
	ColorMagenta = lipgloss.Color("5")
	ColorCyan    = lipgloss.Color("6")
	ColorWhite   = lipgloss.Color("7")
	ColorGray    = lipgloss.Color("8")

	// Semantic colors
	ColorMint   = lipgloss.Color("#98c379") // Title
	ColorVercel = lipgloss.Color("#e5c07b") // Yellow for Vercel
	ColorSwift  = lipgloss.Color("#c678dd") // Magenta for Swift
	ColorGit    = lipgloss.Color("#56b6c2") // Cyan for Git
	ColorGH     = lipgloss.Color("#98c379") // Green for GitHub
)

// =============================================================================
// POWERLINE SEPARATORS (Nerd Fonts)
// =============================================================================

const (
	// Rounded caps
	PLLeftHalfCircle  = "\ue0b6" // U+E0B6 - left half circle thick
	PLRightHalfCircle = "\ue0b4" // U+E0B4 - right half circle thick

	// Triangular separators
	PLLowerLeftTriangle  = "\ue0b8" // U+E0B8 - lower left triangle
	PLUpperRightTriangle = "\ue0be" // U+E0BE - upper right triangle

	// Flame separators
	PLFlameThick         = "\ue0c0" // U+E0C0 - flame thick
	PLFlameThickMirrored = "\ue0c2" // U+E0C2 - flame thick mirrored

	// Hard dividers
	PLLeftHardDivider  = "\ue0b2" // U+E0B2 - left hard divider
	PLRightHardDivider = "\ue0d6" // U+E0D6 - right hard divider
)

// =============================================================================
// ICONS (Nerd Fonts with U+ addresses from spec)
// =============================================================================

const (
	// Title
	IconRocket = "\uf427" // U+F427 oct-rocket

	// Vercel build status
	IconVercel       = "\ue8d3"  // U+E8D3 dev-vercel
	IconReady        = "\uf0063" // U+F0063 md-arrow_up_drop_circle_outline
	IconBuilding     = "\uf1adf" // U+F1ADF md-timer_pause_outline
	IconQueued       = "\uead8"  // U+EAD8 cod-debug
	IconFailed       = "\uead8"  // U+EAD8 cod-debug (same, red color distinguishes)

	// Swift build status
	IconSwift   = "\ue699" // U+E699 seti-swift
	IconCheck   = "\u2714" // U+2714 heavy check mark
	IconX       = "\u2718" // U+2718 heavy ballot x

	// Git status
	IconGit       = "\ue702"  // U+E702 dev-git
	IconStaged    = "\uf1a9e" // U+F1A9E md-file_document_plus_outline
	IconUntracked = "\uf262"  // U+F262 fa-firstdraft
	IconModified  = "\uf459"  // U+F459 oct-diff-modified

	// GitHub status
	IconGitHub = "\ueb00" // U+EB00 cod-github_alt
	IconIssue  = "\uf41b" // U+F41B oct-issue_opened
	IconPR     = "\uf407" // U+F407 oct-git_pull_request

	// Project row action buttons
	IconPush     = "\uf403" // U+F403 oct-repo_push
	IconMerge    = "\ueafe" // U+EAFE cod-git_merge
	IconPlayPause = "\uf04b" // U+F04B fa-play (toggle with F04C pause)
	IconDeploy   = "\uebaa" // U+EBAA cod-cloud
	IconReadme   = "\ueaf0" // U+EAF0 cod-files (readme)
	IconRoadmap  = "\uf018" // U+F018 fa-road
	IconPlan     = "\ueaf0" // U+EAF0 cod-files
	IconTodo     = "\uf0ae" // U+F0AE fa-tasks
	IconChat     = "\uf27a" // U+F27A fa-message

	// Bottom status
	IconProjects  = "\uf502" // U+F502 oct-project
	IconPlus      = "\uea60" // U+EA60 cod-add
	IconConnected = "\ueb99" // U+EB99 cod-account (connected indicator)
	IconBrain     = "\uee9c" // U+EE9C fa-brain
	IconCoins     = "\uede8" // U+EDE8 fa-coins

	// Misc
	IconSearch = "\uf422" // U+F422 oct-search
	IconTime   = "\uf43a" // U+F43A oct-clock
)

// =============================================================================
// SEGMENT STYLES (for top status line)
// =============================================================================

var (
	// Title segment: mint bg, black fg
	TitleSegmentStyle = lipgloss.NewStyle().
		Foreground(ColorBlack).
		Background(ColorMint)

	// Vercel segment: yellow bg, black fg
	VercelSegmentStyle = lipgloss.NewStyle().
		Foreground(ColorBlack).
		Background(ColorVercel)

	// Swift segment: magenta bg, black fg
	SwiftSegmentStyle = lipgloss.NewStyle().
		Foreground(ColorBlack).
		Background(ColorSwift)

	// Git segment: cyan bg, black fg
	GitSegmentStyle = lipgloss.NewStyle().
		Foreground(ColorBlack).
		Background(ColorGit)

	// GitHub segment: green bg, black fg
	GHSegmentStyle = lipgloss.NewStyle().
		Foreground(ColorBlack).
		Background(ColorGH)
)

// =============================================================================
// BOX STYLES (rounded corners for search/chat)
// =============================================================================

var (
	RoundedBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGray).
		Padding(0, 1)

	SearchBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGray).
		Padding(0, 1)

	ChatBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGreen).
		Padding(0, 1)
)

// =============================================================================
// PROJECT LIST STYLES
// =============================================================================

var (
	// Alternating row colors (striped)
	RowEvenStyle = lipgloss.NewStyle().
		Foreground(ColorWhite)

	RowOddStyle = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Background(lipgloss.Color("235")) // Slightly lighter bg

	// Selected row
	SelectedRowStyle = lipgloss.NewStyle().
		Foreground(ColorBlack).
		Background(ColorCyan).
		Bold(true)

	// Column styles
	ProjectNameStyle = lipgloss.NewStyle().
		Width(20).
		MaxWidth(20)

	StatColumnStyle = lipgloss.NewStyle().
		Width(4).
		Align(lipgloss.Right)

	TimeColumnStyle = lipgloss.NewStyle().
		Width(4).
		Align(lipgloss.Right).
		Foreground(ColorGray)

	ActionButtonStyle = lipgloss.NewStyle().
		Foreground(ColorGray).
		PaddingLeft(1)

	ActionButtonActiveStyle = lipgloss.NewStyle().
		Foreground(ColorGreen).
		PaddingLeft(1)
)

// =============================================================================
// BOTTOM STATUS STYLES
// =============================================================================

var (
	BottomStatusStyle = lipgloss.NewStyle().
		Foreground(ColorGray)

	BottomStatusActiveStyle = lipgloss.NewStyle().
		Foreground(ColorGreen)
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// RenderPLSegment renders a powerline segment with proper separators
func RenderPLSegment(content string, style lipgloss.Style, leftCap, rightCap string, fgColor lipgloss.Color) string {
	// Left cap: fg=segment color, bg=none (terminal)
	leftCapStyle := lipgloss.NewStyle().Foreground(style.GetBackground())
	// Right cap: fg=segment color, bg=none
	rightCapStyle := lipgloss.NewStyle().Foreground(style.GetBackground())

	return leftCapStyle.Render(leftCap) + style.Render(content) + rightCapStyle.Render(rightCap)
}

// RenderScrollbar renders an OS9-style scrollbar
func RenderScrollbar(current, total, height int) string {
	if total <= height {
		return ""
	}

	thumbSize := max(1, height*height/total)
	thumbPos := current * (height - thumbSize) / (total - height)

	var sb string
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			sb += "█"
		} else {
			sb += "░"
		}
		if i < height-1 {
			sb += "\n"
		}
	}
	return sb
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
