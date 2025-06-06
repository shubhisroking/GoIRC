package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if !m.ready {
		return systemMessageStyle.Render("Initializing IRC client...")
	}

	if m.err != nil {
		return fmt.Sprintf("%s\n\n%s",
			formatErrorMessage(fmt.Sprintf("Error: %v", m.err)),
			helpStyle.Render("Press any key to quit."))
	}

	if m.state == stateSetup {
		return m.renderSetupView()
	}

	var headerText string
	if m.connected {
		uptime := time.Since(m.connectionTime).Truncate(time.Second)
		headerText = fmt.Sprintf("IRC Client - %s @ %s (%s) - Connected for %v",
			m.currentNick, m.currentChannel, m.config.IRC.Server, uptime)
	} else if m.state == stateConnecting {
		headerText = fmt.Sprintf("IRC Client - Connecting to %s...", m.config.IRC.Server)
	} else {
		headerText = "IRC Client - Disconnected"
	}
	header := headerStyle.Render(headerText)

	var statusText string
	if m.connected {
		joinedChannels := m.getJoinedChannels()
		joinedChannelsList := strings.Join(joinedChannels, ", ")
		if joinedChannelsList == "" {
			joinedChannelsList = "none"
		}
		statusText = fmt.Sprintf("Connected | Current: %s | Joined: [%s] | Tab/Shift+Tab/Alt+1-9 to switch", m.currentChannel, joinedChannelsList)
	} else if m.state == stateConnecting {
		statusText = "Connecting to server..."
	} else {
		statusText = "Disconnected"
	}
	status := statusStyle.Render(statusText)

	chatContent := m.viewport.View()
	chat := chatAreaStyle.Render(chatContent)

	textareaView := m.textarea.View()
	input := inputBoxFocusedStyle.Render(textareaView)

	var help string
	if m.showSidebar {
		help = helpStyle.Render("Commands: /help, /join #channel | Tab: next channel | Shift+Tab: prev channel | Ctrl+B: toggle sidebar | Ctrl+P: command palette | Ctrl+C: exit")
	} else {
		help = helpStyle.Render("Commands: /help, /join #channel, /nick <name>, /quit | Tab: next channel | Shift+Tab: prev channel | Ctrl+B: show sidebar | Ctrl+P: command palette | Ctrl+C: exit")
	}

	// Render command palette overlay if visible
	baseView := ""
	if m.showSidebar {
		sidebar := m.renderSidebar()
		mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input, help)
		contentArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
		baseView = lipgloss.JoinVertical(lipgloss.Left, header, status, contentArea)
	} else {
		mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input, help)
		baseView = lipgloss.JoinVertical(lipgloss.Left, header, status, mainContent)
	}

	if m.commandPaletteVisible {
		// Create an enhanced command palette overlay with better centering
		overlay := m.renderCommandPalette()

		// Place the command palette in the center with improved positioning
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
	}

	return baseView
}

func (m model) renderSetupView() string {
	var content []string

	// Welcome section with improved spacing
	if m.setupPhase == setupServer {
		welcomeBox := setupWelcomeBoxStyle.Render(
			"Welcome to GoIRC!\n\n" +
				"Let's get you connected to your favorite IRC networks.\n" +
				"This setup wizard will guide you through the configuration process.\n\n" +
				"Features:\n" +
				"  - Modern terminal-based UI\n" +
				"  - Multiple channel support\n" +
				"  - SSL/TLS encryption\n" +
				"  - Customizable interface")
		content = append(content, welcomeBox)
		content = append(content, "")
	}

	// Main header with better spacing
	title := setupTitleStyle.Render("GoIRC Setup Wizard")
	content = append(content, title)
	content = append(content, setupSubtitleStyle.Render("Modern IRC Client Configuration"))
	content = append(content, "") // Extra spacing after header

	// Progress bar
	progressBar := m.renderProgressBar()
	content = append(content, progressBar)
	content = append(content, "") // Extra spacing after progress

	switch m.setupPhase {
	case setupServer:
		content = append(content, setupStepHeaderStyle.Render("Step 1/4: Server Configuration"))
		content = append(content, setupDescStyle.Render("Connect to your favorite IRC server. We support both standard and SSL connections."))

		// Server input section with validation hints
		serverLabel := setupLabelStyle.Render("IRC Server Address:")
		content = append(content, serverLabel)

		// Show validation error if any
		if m.setupValidationError != "" {
			content = append(content, setupValidationStyle.Render("Warning: "+m.setupValidationError))
		}

		content = append(content, setupHintStyle.Render(fmt.Sprintf("Default: %s (press Enter to use default)", defaultServer)))

		// Examples box with more servers
		exampleBox := setupExampleBoxStyle.Render(
			"Popular IRC Networks:\n\n" +
				"   Libera.Chat:     irc.libera.chat:6697 (SSL)\n" +
				"   OFTC:           irc.oftc.net:6697 (SSL)\n" +
				"   Rizon:          irc.rizon.net:6697 (SSL)\n" +
				"   EFnet:          irc.efnet.org:6697 (SSL)\n" +
				"   Freenode:       chat.freenode.net:6667\n\n" +
				"Tip: Use port 6697 for SSL, 6667 for standard")
		content = append(content, exampleBox)

	case setupNick:
		content = append(content, setupStepHeaderStyle.Render("Step 2/4: Your Identity"))
		content = append(content, setupDescStyle.Render("Choose a unique nickname that represents you on IRC. Make it memorable!"))

		// Server confirmation with SSL indicator
		sslText := "Standard"
		if m.config.IRC.UseSSL {
			sslText = "SSL/TLS"
		}
		serverInfo := setupInfoBoxStyle.Render(fmt.Sprintf("Server Configuration Complete\n\nServer: %s\nConnection: %s",
			m.config.IRC.Server, sslText))
		content = append(content, serverInfo)

		nickLabel := setupLabelStyle.Render("Your Nickname:")
		content = append(content, nickLabel)

		if m.setupValidationError != "" {
			content = append(content, setupValidationStyle.Render("Warning: "+m.setupValidationError))
		}

		content = append(content, setupHintStyle.Render(fmt.Sprintf("Default: %s (press Enter to use default)", defaultNick)))
		content = append(content, setupHintStyle.Render("Tip: Choose 3-16 characters, letters and numbers only"))

	case setupChannels:
		content = append(content, setupStepHeaderStyle.Render("Step 3/4: Join Channels"))
		content = append(content, setupDescStyle.Render("Channels are where conversations happen. Join some to get started!"))

		// Configuration summary
		sslText := "Standard"
		if m.config.IRC.UseSSL {
			sslText = "SSL/TLS"
		}
		configInfo := setupInfoBoxStyle.Render(fmt.Sprintf(
			"Configuration Progress\n\nServer: %s\nConnection: %s\nNickname: %s",
			m.config.IRC.Server, sslText, m.config.IRC.Nick))
		content = append(content, configInfo)

		channelLabel := setupLabelStyle.Render("Channels to Join:")
		content = append(content, channelLabel)

		if m.setupValidationError != "" {
			content = append(content, setupValidationStyle.Render("Warning: "+m.setupValidationError))
		}

		content = append(content, setupHintStyle.Render(fmt.Sprintf("Default: %s (press Enter to use default)", defaultChannel)))
		content = append(content, setupHintStyle.Render("Tip: Separate multiple channels with commas (e.g., #general, #help, #dev)"))

		// Popular channels example
		exampleBox := setupExampleBoxStyle.Render(
			"Popular Channels by Network:\n\n" +
				"   Libera.Chat:  #archlinux, #ubuntu, #python, #javascript\n" +
				"   OFTC:        #debian, #tor, #spi\n" +
				"   Rizon:       #news, #anime, #programming\n\n" +
				"Tip: Channel names start with # (automatically added)")
		content = append(content, exampleBox)

	case setupConfirm:
		content = append(content, setupStepHeaderStyle.Render("Step 4/4: Ready to Connect"))
		content = append(content, setupDescStyle.Render("Review your configuration and let's get you connected to IRC!"))

		// Final configuration summary
		sslText := "Standard Connection"
		if m.config.IRC.UseSSL {
			sslText = "Secure SSL/TLS Connection"
		}

		channelList := strings.Join(m.config.IRC.Channels, ", ")
		if len(channelList) > 50 {
			channelList = channelList[:47] + "..."
		}

		summaryBox := setupSummaryBoxStyle.Render(
			"Configuration Complete!\n\n" +
				fmt.Sprintf("Server:      %s\n", m.config.IRC.Server) +
				fmt.Sprintf("Connection:  %s\n", sslText) +
				fmt.Sprintf("Nickname:    %s\n", m.config.IRC.Nick) +
				fmt.Sprintf("Channels:    %s\n\n", channelList) +
				"Ready to connect and start chatting!")
		content = append(content, summaryBox)

		actionHint := setupActionStyle.Render("Press Enter to connect • Type 'r' to restart setup • Ctrl+C to exit")
		content = append(content, actionHint)
	}

	// Enhanced input box with better prompts
	if m.setupPhase != setupConfirm {
		inputBox := setupInputBoxStyle.Render(m.textarea.View())
		content = append(content, inputBox)
	}

	// Footer with helpful controls
	var footerText string
	switch m.setupPhase {
	case setupServer:
		footerText = "Tip: Press Tab for autocomplete • Enter to continue • Ctrl+C to exit"
	case setupNick:
		footerText = "Your nickname is your identity on IRC • Enter to continue • Ctrl+C to exit"
	case setupChannels:
		footerText = "You can join more channels later with /join • Enter to continue • Ctrl+C to exit"
	default:
		footerText = "Almost there! • Enter to connect • Ctrl+C to exit"
	}

	footer := setupFooterStyle.Render(footerText)
	content = append(content, footer)

	return lipgloss.JoinVertical(lipgloss.Center, content...)
}

func (m model) renderProgressBar() string {
	steps := []string{"Server", "Nickname", "Channels", "Confirm"}
	current := int(m.setupPhase)

	var segments []string
	for i := range steps {
		if i < current {
			// Completed step
			segments = append(segments, setupProgressCompletedStyle.Render("●"))
		} else if i == current {
			// Current step
			segments = append(segments, setupProgressCurrentStyle.Render("●"))
		} else {
			// Pending step
			segments = append(segments, setupProgressPendingStyle.Render("○"))
		}

		if i < len(steps)-1 {
			// Add connection line between steps
			if i < current {
				segments = append(segments, setupProgressCompletedStyle.Render("───"))
			} else if i == current-1 {
				// Gradient effect: completed to current
				segments = append(segments, setupProgressCurrentStyle.Render("───"))
			} else {
				segments = append(segments, setupProgressPendingStyle.Render("───"))
			}
		}
	}

	progressLine := lipgloss.JoinHorizontal(lipgloss.Left, segments...)

	// Add step labels
	var labels []string
	for i, step := range steps {
		var style lipgloss.Style
		var label string

		if i < current {
			style = setupProgressLabelCompletedStyle
			label = fmt.Sprintf("[%s]", step)
		} else if i == current {
			style = setupProgressLabelCurrentStyle
			label = fmt.Sprintf("> %s", step)
		} else {
			style = setupProgressLabelPendingStyle
			label = step
		}
		labels = append(labels, style.Render(label))
	}

	// Proper spacing for labels to align with progress dots
	// Calculate spacing based on label width to improve alignment
	spacing := "   " // Base spacing

	// Calculate proper spacing for visual balance
	labelLine := lipgloss.JoinHorizontal(lipgloss.Left,
		labels[0], spacing, labels[1], spacing, labels[2], spacing, labels[3])

	// Progress percentage
	progressPercent := fmt.Sprintf("%d%% Complete", (current*100)/len(steps))
	percentStyle := lipgloss.NewStyle().
		Foreground(textMuted).
		Align(lipgloss.Center)

	return setupProgressContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			progressLine,
			"",
			labelLine,
			"",
			percentStyle.Render(progressPercent)))
}

func (m model) renderSidebar() string {
	if !m.showSidebar {
		return ""
	}

	var content []string
	sidebarHeight := m.height - headerHeight - footerHeight - statusHeight - 4 // Improved calculation

	// Sidebar header
	content = append(content, sidebarHeaderStyle.Render("IRC CLIENT"))
	content = append(content, "")

	// Connection status with better spacing
	statusText := ""
	if m.connected {
		statusIcon := sidebarStatusDotStyle.Render("●")
		statusText = fmt.Sprintf("%s CONNECTED", statusIcon)
	} else {
		statusIcon := sidebarDisconnectedDotStyle.Render("●")
		statusText = fmt.Sprintf("%s DISCONNECTED", statusIcon)
	}
	content = append(content, sidebarSectionStyle.Render(statusText))
	content = append(content, "")

	// User and server info with better formatting
	if m.connected && m.currentNick != "" {
		content = append(content, sidebarItemStyle.Render(fmt.Sprintf("User: %s", m.currentNick)))
		serverName := strings.Split(m.config.IRC.Server, ":")[0]
		if len(serverName) > 18 {
			serverName = serverName[:15] + "..."
		}
		content = append(content, sidebarItemStyle.Render(fmt.Sprintf("Server: %s", serverName)))
		content = append(content, "")
	}

	// Channels section with improved layout
	joinedChannels := m.getJoinedChannels()
	channelCountBadge := sidebarChannelCountStyle.Render(fmt.Sprintf(" %d ", len(joinedChannels)))
	channelsHeader := fmt.Sprintf("CHANNELS %s", channelCountBadge)
	content = append(content, sidebarSectionStyle.Render(channelsHeader))

	if len(joinedChannels) > 0 {
		for i, channelName := range joinedChannels {
			// Truncate long channel names
			displayName := channelName
			if len(displayName) > 20 {
				displayName = displayName[:17] + "..."
			}

			channelLine := fmt.Sprintf("%d %s", i+1, displayName)

			if channelName == m.currentChannel {
				content = append(content, sidebarActiveItemStyle.Render(fmt.Sprintf("> %s", channelLine)))
			} else {
				content = append(content, sidebarItemStyle.Render(fmt.Sprintf("  %s", channelLine)))
			}
		}
	} else {
		content = append(content, sidebarItemStyle.Render("  No channels joined"))
	}

	// Improved spacing and height management
	maxLines := sidebarHeight - 2
	if maxLines < 10 {
		maxLines = 10 // Minimum height to prevent UI breaking
	}

	if len(content) < maxLines {
		for len(content) < maxLines {
			content = append(content, "")
		}
	} else if len(content) > maxLines {
		// Truncate and add ellipsis
		content = content[:maxLines-1]
		content = append(content, sidebarItemStyle.Render("..."))
	}

	sidebarContent := strings.Join(content, "\n")
	return sidebarStyle.Height(sidebarHeight).Render(sidebarContent)
}

func (m model) renderCommandPalette() string {
	const maxItems = 12

	// Make palette width responsive to screen size
	paletteWidth := 84
	if m.width > 0 && paletteWidth > m.width-6 {
		paletteWidth = m.width - 6
	}
	if paletteWidth < 60 {
		paletteWidth = 60 // Minimum width
	}

	var content []string

	// Enhanced header with gradient-like effect and better styling
	totalCommands := len(m.commandPaletteItems)
	filteredCommands := len(m.commandPaletteFiltered)

	var headerText string
	if m.commandPaletteQuery == "" {
		headerText = fmt.Sprintf("Command Palette - %d Commands", totalCommands)
	} else {
		headerText = fmt.Sprintf("Search Results - %d of %d", filteredCommands, totalCommands)
	}
	content = append(content, commandPaletteHeaderStyle.Render(headerText))

	// Search query with improved display
	query := m.commandPaletteQuery
	queryPrompt := "> "
	cursorChar := "|"

	if query == "" {
		queryDisplay := "Type to search commands..."
		content = append(content, commandPaletteQueryStyle.Render(queryPrompt+queryDisplay))
	} else {
		queryDisplay := query + cursorChar
		content = append(content, commandPaletteQueryStyle.Render(queryPrompt+queryDisplay))
	}

	// Category separator with better spacing
	if m.commandPaletteQuery != "" && len(m.commandPaletteFiltered) > 0 {
		separatorText := fmt.Sprintf("--- Results (%d found) ---", len(m.commandPaletteFiltered))
		content = append(content, commandPaletteSeparatorStyle.Render(separatorText))
	} else if m.commandPaletteQuery == "" {
		content = append(content, commandPaletteSeparatorStyle.Render("--- Available Commands ---"))
	}

	// Items with enhanced display and better visual hierarchy
	displayItems := m.commandPaletteFiltered
	displaySelectionIndex := m.commandPaletteSelected

	if len(displayItems) > maxItems {
		// Show items around the selected one with scroll indicators
		start := m.commandPaletteSelected - maxItems/2
		if start < 0 {
			start = 0
		}
		end := start + maxItems
		if end > len(displayItems) {
			end = len(displayItems)
			start = end - maxItems
			if start < 0 {
				start = 0
			}
		}

		// Add scroll indicators
		if start > 0 {
			content = append(content, commandPaletteItemStyle.Render("    ^  More commands above..."))
		}

		displayItems = displayItems[start:end]
		displaySelectionIndex = m.commandPaletteSelected - start

		if end < len(m.commandPaletteFiltered) {
			// We'll add the bottom indicator after the items
		}
	}

	// Track current category for grouping
	var lastCategory string
	itemCount := 0

	for i, item := range displayItems {
		// Add category header (only when not searching) with better spacing
		if m.commandPaletteQuery == "" && item.category != lastCategory {
			if lastCategory != "" {
				content = append(content, "") // Add spacing between categories
			}

			categoryText := fmt.Sprintf("-- %s", item.category)
			content = append(content, commandPaletteCategoryStyle.Render(categoryText))
			lastCategory = item.category
		}

		// Item styling
		var style lipgloss.Style
		selectionIndicator := "  "

		if i == displaySelectionIndex {
			style = commandPaletteSelectedStyle
			selectionIndicator = "> "
		} else {
			style = commandPaletteItemStyle
		}

		// Item display with responsive layout
		iconAndName := fmt.Sprintf("%s%s %s", selectionIndicator, item.icon, item.name)

		// Calculate responsive widths based on palette width
		nameWidth := paletteWidth / 3
		if nameWidth < 20 {
			nameWidth = 20
		}
		shortcutWidth := 15
		descWidth := paletteWidth - nameWidth - shortcutWidth - 8

		// Smart truncation with ellipsis
		if len(iconAndName) > nameWidth {
			iconAndName = iconAndName[:nameWidth-3] + "..."
		}

		description := item.description
		if len(description) > descWidth {
			description = description[:descWidth-3] + "..."
		}

		shortcut := ""
		if item.shortcut != "" {
			shortcut = item.shortcut
		}

		// Create visually appealing item line with proper spacing
		itemLine := fmt.Sprintf("%-*s  %-*s  %s",
			nameWidth, iconAndName,
			descWidth, description,
			shortcut)

		content = append(content, style.Render(itemLine))
		itemCount++
	}

	// Add scroll indicator for bottom if needed
	if len(displayItems) == maxItems && m.commandPaletteSelected+1 < len(m.commandPaletteFiltered) {
		content = append(content, commandPaletteItemStyle.Render("    v  More commands below..."))
	}

	// Show "no results" message with better spacing
	if len(displayItems) == 0 && m.commandPaletteQuery != "" {
		content = append(content, "")
		content = append(content, commandPaletteEmptyStyle.Render("No commands found"))
		content = append(content, commandPaletteEmptyStyle.Render(fmt.Sprintf("No matches for \"%s\"", m.commandPaletteQuery)))
		content = append(content, commandPaletteEmptyStyle.Render("Try different keywords or clear search"))
	}

	// Fill empty space to maintain consistent height
	minHeight := maxItems + 8
	for len(content) < minHeight {
		content = append(content, "")
	}

	// Footer with improved text
	var footerText string
	if len(m.commandPaletteFiltered) > 0 {
		selectedNum := m.commandPaletteSelected + 1
		totalNum := len(m.commandPaletteFiltered)
		footerText = fmt.Sprintf("↑↓ Navigate (%d/%d) • Enter Execute • Esc Close • Ctrl+P Toggle", selectedNum, totalNum)
	} else {
		footerText = "Start typing to search commands • Esc Close"
	}
	content = append(content, commandPaletteFooterStyle.Render(footerText))

	return commandPaletteBorderStyle.Width(paletteWidth).Render(strings.Join(content, "\n"))
}
