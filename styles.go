package main

import "github.com/charmbracelet/lipgloss"

var (
	// Minimal text colors
	textPrimary   = lipgloss.Color("#FFFFFF")
	textSecondary = lipgloss.Color("#A0A0A0")
	textMuted     = lipgloss.Color("#666666")

	// Clean background colors
	bgPrimary   = lipgloss.Color("#000000") // Pure black
	bgSecondary = lipgloss.Color("#111111") // Slightly lighter

	// Minimal border colors
	borderColor      = lipgloss.Color("#333333")
	borderColorFocus = lipgloss.Color("#666666")
)

var (
	headerStyle = lipgloss.NewStyle().
			Background(bgSecondary).
			Foreground(textPrimary).
			Padding(0, 2).
			Width(100)

	statusStyle = lipgloss.NewStyle().
			Background(bgPrimary).
			Foreground(textSecondary).
			Padding(0, 1).
			Width(100)

	systemMessageStyle = lipgloss.NewStyle().
				Foreground(textSecondary)

	userMessageStyle = lipgloss.NewStyle().
				Foreground(textPrimary)

	ownMessageStyle = lipgloss.NewStyle().
			Foreground(textPrimary)

	joinMessageStyle = lipgloss.NewStyle().
				Foreground(textMuted)

	partMessageStyle = lipgloss.NewStyle().
				Foreground(textMuted)

	quitMessageStyle = lipgloss.NewStyle().
				Foreground(textMuted)

	noticeMessageStyle = lipgloss.NewStyle().
				Foreground(textSecondary)

	errorMessageStyle = lipgloss.NewStyle().
				Foreground(textPrimary)

	channelSwitchStyle = lipgloss.NewStyle().
				Foreground(textSecondary)

	timestampStyle = lipgloss.NewStyle().
			Foreground(textMuted)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Width(100)

	inputBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(borderColorFocus).
				Padding(0, 1).
				Width(100)

	chatAreaStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Background(bgPrimary)

	helpStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Align(lipgloss.Center).
			Padding(0, 1).
			Margin(1, 0, 0, 0)

	sidebarStyle = lipgloss.NewStyle().
			Background(bgSecondary).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(borderColor).
			Padding(1).
			Width(30).
			Height(20)

	sidebarHeaderStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(0, 1).
				Margin(0, 0, 1, 0).
				Width(26)

	sidebarItemStyle = lipgloss.NewStyle().
				Foreground(textSecondary).
				Padding(0, 1).
				Margin(0, 0, 0, 0)

	sidebarActiveItemStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Background(bgPrimary).
				Padding(0, 1).
				Margin(0, 0, 0, 0)

	sidebarSectionStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(0, 1).
				Margin(0, 0, 0, 0)

	sidebarChannelCountStyle = lipgloss.NewStyle().
					Foreground(textSecondary).
					Background(bgPrimary).
					Padding(0, 1)

	sidebarStatusDotStyle = lipgloss.NewStyle().
				Foreground(textPrimary)

	sidebarDisconnectedDotStyle = lipgloss.NewStyle().
					Foreground(textMuted)

	// Setup wizard styles - Minimal design
	setupTitleStyle = lipgloss.NewStyle().
			Foreground(textPrimary).
			Padding(1, 0).
			Margin(1, 0, 2, 0).
			Align(lipgloss.Center).
			Width(80)

	setupSubtitleStyle = lipgloss.NewStyle().
				Foreground(textSecondary).
				Align(lipgloss.Center).
				Width(80).
				Margin(0, 0, 2, 0)

	setupStepHeaderStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(1, 0).
				Margin(1, 0, 1, 0).
				Width(70)

	setupDescStyle = lipgloss.NewStyle().
			Foreground(textSecondary).
			Margin(0, 0, 2, 0).
			Width(70)

	setupLabelStyle = lipgloss.NewStyle().
			Foreground(textPrimary).
			Margin(1, 0, 0, 0)

	setupHintStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Margin(0, 0, 1, 0)

	setupExampleBoxStyle = lipgloss.NewStyle().
				Background(bgSecondary).
				Foreground(textSecondary).
				Padding(1, 2).
				Margin(1, 0, 2, 0).
				Border(lipgloss.NormalBorder()).
				BorderForeground(borderColor).
				Width(70)

	setupInfoBoxStyle = lipgloss.NewStyle().
				Background(bgSecondary).
				Foreground(textSecondary).
				Padding(1, 2).
				Margin(1, 0, 2, 0).
				Border(lipgloss.NormalBorder()).
				BorderForeground(borderColor).
				Width(70)

	setupSummaryBoxStyle = lipgloss.NewStyle().
				Background(bgSecondary).
				Foreground(textPrimary).
				Padding(2, 3).
				Margin(1, 0, 2, 0).
				Border(lipgloss.NormalBorder()).
				BorderForeground(borderColor).
				Width(70)

	setupInputBoxStyle = lipgloss.NewStyle().
				Background(bgPrimary).
				Foreground(textPrimary).
				Padding(0, 1).
				Margin(1, 0, 2, 0).
				Border(lipgloss.NormalBorder()).
				BorderForeground(borderColor).
				Width(70)

	setupActionStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(1, 2).
				Margin(1, 0).
				Align(lipgloss.Center).
				Width(70)

	setupFooterStyle = lipgloss.NewStyle().
				Foreground(textMuted).
				Padding(1, 2).
				Margin(2, 0, 0, 0).
				Align(lipgloss.Center).
				Width(80)

	// Progress bar styles - Minimal design
	setupProgressContainerStyle = lipgloss.NewStyle().
					Align(lipgloss.Center).
					Margin(1, 0, 2, 0).
					Padding(1, 0).
					Width(78)

	setupProgressCompletedStyle = lipgloss.NewStyle().
					Foreground(textPrimary)

	setupProgressCurrentStyle = lipgloss.NewStyle().
					Foreground(textPrimary)

	setupProgressPendingStyle = lipgloss.NewStyle().
					Foreground(textMuted)

	setupProgressLabelCompletedStyle = lipgloss.NewStyle().
						Foreground(textPrimary).
						Width(12).
						Align(lipgloss.Center)

	setupProgressLabelCurrentStyle = lipgloss.NewStyle().
					Foreground(textPrimary).
					Width(12).
					Align(lipgloss.Center)

	setupProgressLabelPendingStyle = lipgloss.NewStyle().
					Foreground(textMuted).
					Width(12).
					Align(lipgloss.Center)

	// Minimal setup experience styles
	setupWelcomeBoxStyle = lipgloss.NewStyle().
				Background(bgSecondary).
				Foreground(textPrimary).
				Padding(2, 3).
				Margin(1, 0, 2, 0).
				Border(lipgloss.NormalBorder()).
				BorderForeground(borderColor).
				Width(80)

	setupValidationStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Margin(0, 0, 1, 0)

	// Command palette styles - Minimal design
	commandPaletteBorderStyle = lipgloss.NewStyle().
					Background(bgPrimary).
					Border(lipgloss.NormalBorder()).
					BorderForeground(borderColor).
					Padding(1, 2).
					Width(84).
					Height(20)

	commandPaletteHeaderStyle = lipgloss.NewStyle().
					Foreground(textPrimary).
					Padding(0, 1).
					Width(76).
					Margin(0, 0, 1, 0)

	commandPaletteQueryStyle = lipgloss.NewStyle().
					Foreground(textPrimary).
					Background(bgSecondary).
					Width(76).
					Padding(0, 1).
					Margin(0, 0, 1, 0).
					Border(lipgloss.NormalBorder()).
					BorderForeground(borderColor)

	commandPaletteItemStyle = lipgloss.NewStyle().
				Foreground(textSecondary).
				Padding(0, 1).
				Width(76).
				Margin(0, 0, 0, 0)

	commandPaletteSelectedStyle = lipgloss.NewStyle().
					Background(bgSecondary).
					Foreground(textPrimary).
					Padding(0, 1).
					Width(76)

	commandPaletteFooterStyle = lipgloss.NewStyle().
					Foreground(textMuted).
					Padding(0, 1).
					Width(76).
					Margin(1, 0, 0, 0)

	commandPaletteSeparatorStyle = lipgloss.NewStyle().
					Foreground(textSecondary).
					Width(76).
					Padding(0, 1).
					Margin(0, 0, 1, 0).
					Align(lipgloss.Center)

	commandPaletteCategoryStyle = lipgloss.NewStyle().
					Foreground(textPrimary).
					Padding(0, 1).
					Margin(0, 0, 0, 0).
					Width(74)

	commandPaletteEmptyStyle = lipgloss.NewStyle().
					Foreground(textMuted).
					Align(lipgloss.Center).
					Width(76).
					Padding(1, 1).
					Margin(1, 0, 1, 0)
)

func UpdateStyleWidths(width int) {
	sidebarWidth := 30
	if width < 120 {
		sidebarWidth = 25 // Smaller sidebar for narrow screens
	}

	chatWidth := width - sidebarWidth - 4 // Account for sidebar width and borders
	if chatWidth < 40 {
		chatWidth = 40 // Minimum chat width
	}

	headerStyle = headerStyle.Width(width)
	statusStyle = statusStyle.Width(width)
	inputBoxStyle = inputBoxStyle.Width(chatWidth)
	inputBoxFocusedStyle = inputBoxFocusedStyle.Width(chatWidth)
	chatAreaStyle = chatAreaStyle.Width(chatWidth)
	sidebarStyle = sidebarStyle.Width(sidebarWidth)
	sidebarHeaderStyle = sidebarHeaderStyle.Width(sidebarWidth - 4)

	// Update command palette styles to be responsive
	paletteWidth := 84
	if width > 0 && paletteWidth > width-6 {
		paletteWidth = width - 6
	}
	if paletteWidth < 60 {
		paletteWidth = 60
	}

	commandPaletteBorderStyle = commandPaletteBorderStyle.Width(paletteWidth)
	commandPaletteHeaderStyle = commandPaletteHeaderStyle.Width(paletteWidth - 8)
	commandPaletteQueryStyle = commandPaletteQueryStyle.Width(paletteWidth - 8)
	commandPaletteItemStyle = commandPaletteItemStyle.Width(paletteWidth - 8)
	commandPaletteSelectedStyle = commandPaletteSelectedStyle.Width(paletteWidth - 8)
	commandPaletteFooterStyle = commandPaletteFooterStyle.Width(paletteWidth - 8)
	commandPaletteSeparatorStyle = commandPaletteSeparatorStyle.Width(paletteWidth - 8)
	commandPaletteCategoryStyle = commandPaletteCategoryStyle.Width(paletteWidth - 10)
	commandPaletteEmptyStyle = commandPaletteEmptyStyle.Width(paletteWidth - 8)
}
