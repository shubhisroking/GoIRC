package main

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#06B6D4")
	accentColor    = lipgloss.Color("#F59E0B")
	successColor   = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")
	warningColor   = lipgloss.Color("#F59E0B")

	textPrimary   = lipgloss.Color("#F8FAFC")
	textSecondary = lipgloss.Color("#94A3B8")
	textMuted     = lipgloss.Color("#64748B")

	bgPrimary   = lipgloss.Color("#0F172A")
	bgSecondary = lipgloss.Color("#1E293B")

	borderColor      = lipgloss.Color("#475569")
	borderColorFocus = primaryColor
)

var (
	headerStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(textPrimary).
			Bold(true).
			Padding(0, 2).
			Width(100).
			Align(lipgloss.Center)

	statusStyle = lipgloss.NewStyle().
			Background(bgSecondary).
			Foreground(textSecondary).
			Padding(0, 1).
			Width(100)

	systemMessageStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	userMessageStyle = lipgloss.NewStyle().
				Foreground(textPrimary)

	ownMessageStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	joinMessageStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Italic(true)

	partMessageStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Italic(true)

	quitMessageStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Italic(true)

	noticeMessageStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Italic(true)

	errorMessageStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	channelSwitchStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				Italic(true)

	timestampStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Faint(true)

	urlTitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Width(100)

	inputBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColorFocus).
				Padding(0, 1).
				Width(100)

	chatAreaStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Background(bgPrimary)

	helpStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Italic(true).
			Align(lipgloss.Center)

	sidebarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#0F1419")).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("#2D3748")).
			Padding(1).
			Width(30).
			Height(20)

	sidebarHeaderStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#553C9A")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Align(lipgloss.Center).
				Padding(0, 1).
				Margin(0, 0, 1, 0).
				Width(26).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED"))

	sidebarItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A0AEC0")).
				Padding(0, 1).
				Margin(0, 0, 0, 0)

	sidebarActiveItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#553C9A")).
				Bold(true).
				Padding(0, 1).
				Margin(0, 0, 0, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED"))

	sidebarSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FBD38D")).
				Bold(true).
				Padding(0, 1).
				Margin(0, 0, 0, 0).
				Background(lipgloss.Color("#2D3748")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4A5568"))

	sidebarChannelCountStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#68D391")).
					Bold(true).
					Background(lipgloss.Color("#1A202C")).
					Padding(0, 1).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#2D3748"))

	sidebarStatusDotStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#68D391")).
				Bold(true)

	sidebarDisconnectedDotStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#F56565")).
					Bold(true)

	sidebarDividerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4A5568")).
				Margin(0, 0, 0, 0)
)

func UpdateStyleWidths(width int) {
	sidebarWidth := 30
	chatWidth := width - sidebarWidth - 4 // Account for sidebar width and borders

	headerStyle = headerStyle.Width(width)
	statusStyle = statusStyle.Width(width)
	inputBoxStyle = inputBoxStyle.Width(chatWidth)
	inputBoxFocusedStyle = inputBoxFocusedStyle.Width(chatWidth)
	chatAreaStyle = chatAreaStyle.Width(chatWidth)
	sidebarStyle = sidebarStyle.Width(sidebarWidth)
	sidebarHeaderStyle = sidebarHeaderStyle.Width(sidebarWidth - 4)
}
