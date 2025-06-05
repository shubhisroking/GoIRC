package main

import "github.com/charmbracelet/lipgloss"

var (
	// Enhanced color palette with gradients and modern colors
	primaryColor   = lipgloss.Color("#8B5CF6") // Vibrant purple
	secondaryColor = lipgloss.Color("#06B6D4") // Cyan
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	successColor   = lipgloss.Color("#10B981") // Emerald
	errorColor     = lipgloss.Color("#EF4444") // Red
	warningColor   = lipgloss.Color("#F59E0B") // Amber

	// Modern text colors with better contrast
	textPrimary   = lipgloss.Color("#FFFFFF")
	textSecondary = lipgloss.Color("#CBD5E1")
	textMuted     = lipgloss.Color("#94A3B8")

	// Rich background colors
	bgPrimary   = lipgloss.Color("#0F0F23") // Deep dark blue
	bgSecondary = lipgloss.Color("#1E1E3F") // Slightly lighter

	// Enhanced border colors
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

	// Setup wizard styles - Enhanced for better aesthetics
	setupTitleStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#8B5CF6", Dark: "#8B5CF6"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Bold(true).
			Padding(2, 6).
			Margin(1, 0, 2, 0).
			Align(lipgloss.Center).
			Width(80).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#A855F7", Dark: "#A855F7"})

	setupSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#CBD5E1"}).
				Italic(true).
				Align(lipgloss.Center).
				Width(80).
				Margin(0, 0, 3, 0)

	setupStepHeaderStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#EFF6FF", Dark: "#1E3A8A"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#1E40AF", Dark: "#60A5FA"}).
				Bold(true).
				Padding(1, 3).
				Margin(1, 0, 2, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#60A5FA", Dark: "#3B82F6"}).
				Width(70)

	setupDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#CBD5E1"}).
			Margin(0, 0, 2, 0).
			Width(70).
			Align(lipgloss.Left)

	setupLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#F8FAFC"}).
			Bold(true).
			Margin(1, 0, 0, 0)

	setupHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}).
			Italic(true).
			Margin(0, 0, 2, 0)

	setupExampleBoxStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#F0F9FF", Dark: "#0C4A6E"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#0369A1", Dark: "#7DD3FC"}).
				Padding(2, 3).
				Margin(1, 0, 2, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#0284C7"}).
				Width(70)

	setupInfoBoxStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#ECFDF5", Dark: "#064E3B"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#047857", Dark: "#6EE7B7"}).
				Padding(2, 3).
				Margin(1, 0, 2, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#059669"}).
				Width(70)

	setupSummaryBoxStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#F0FDF4", Dark: "#14532D"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#BBF7D0"}).
				Padding(3, 4).
				Margin(1, 0, 3, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#22C55E", Dark: "#16A34A"}).
				Width(70)

	setupInputBoxStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0F0F23"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#F8FAFC"}).
				Padding(1, 3).
				Margin(1, 0, 2, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#8B5CF6", Dark: "#8B5CF6"}).
				Width(70)

	setupActionStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#FEF3C7", Dark: "#451A03"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#92400E", Dark: "#FCD34D"}).
				Bold(true).
				Padding(1, 3).
				Margin(2, 0).
				Align(lipgloss.Center).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#D97706"}).
				Width(70)

	setupFooterStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#F1F5F9", Dark: "#1E293B"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#475569", Dark: "#CBD5E1"}).
				Padding(2, 3).
				Margin(3, 0, 0, 0).
				Align(lipgloss.Center).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#CBD5E1", Dark: "#475569"}).
				Width(80)

	// Progress bar styles - Enhanced with better visual feedback
	setupProgressContainerStyle = lipgloss.NewStyle().
					Align(lipgloss.Center).
					Margin(2, 0, 3, 0).
					Padding(2, 0).
					Background(lipgloss.AdaptiveColor{Light: "#F8FAFC", Dark: "#1E293B"}).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#475569"}).
					Width(78)

	setupProgressCompletedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}).
					Bold(true)

	setupProgressCurrentStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#8B5CF6", Dark: "#A78BFA"}).
					Bold(true)

	setupProgressPendingStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#CBD5E1", Dark: "#64748B"})

	setupProgressLabelCompletedStyle = lipgloss.NewStyle().
						Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}).
						Bold(true).
						Width(12).
						Align(lipgloss.Center)

	setupProgressLabelCurrentStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#8B5CF6", Dark: "#A78BFA"}).
					Bold(true).
					Width(12).
					Align(lipgloss.Center)

	setupProgressLabelPendingStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#CBD5E1", Dark: "#64748B"}).
					Width(12).
					Align(lipgloss.Center)

	// New styles for enhanced setup experience
	setupWelcomeBoxStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#FEF7FF", Dark: "#581C87"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#7C2D12", Dark: "#E879F9"}).
				Padding(3, 4).
				Margin(2, 0, 3, 0).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#C084FC", Dark: "#A855F7"}).
				Width(80)

	setupValidationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#FCA5A5"}).
				Italic(true).
				Margin(0, 0, 1, 0)
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
