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
)

func UpdateStyleWidths(width int) {
	headerStyle = headerStyle.Width(width)
	statusStyle = statusStyle.Width(width)
	inputBoxStyle = inputBoxStyle.Width(width - 4)
	inputBoxFocusedStyle = inputBoxFocusedStyle.Width(width - 4)
	chatAreaStyle = chatAreaStyle.Width(width - 4)
}
