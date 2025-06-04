package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if !m.ready {
		return systemMessageStyle.Render("🚀 Initializing IRC client...")
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
			m.currentNick, m.currentChannel, m.config.Server, uptime)
	} else if m.state == stateConnecting {
		headerText = fmt.Sprintf("IRC Client - Connecting to %s...", m.config.Server)
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
		help = helpStyle.Render("Commands: /help, /join #channel, /switch <channel> | F1: toggle sidebar | Tab/Shift+Tab/Alt+1-9: switch channels | Ctrl+C: exit")
	} else {
		help = helpStyle.Render("Commands: /help, /join #channel, /switch <channel>, /nick <name>, /quit | F1: show sidebar | Tab/Shift+Tab/Alt+1-9: switch channels | Ctrl+C: exit")
	}

	// Layout with sidebar
	if m.showSidebar {
		sidebar := m.renderSidebar()
		mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input, help)
		contentArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
		return lipgloss.JoinVertical(lipgloss.Left, header, status, contentArea)
	}

	// Layout without sidebar
	mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input, help)
	return lipgloss.JoinVertical(lipgloss.Left, header, status, mainContent)
}

func (m model) renderSetupView() string {
	var content []string

	content = append(content, headerStyle.Render("IRC Client Setup"))
	content = append(content, "")

	switch m.setupPhase {
	case setupServer:
		content = append(content, systemMessageStyle.Render("Step 1/4: Server Configuration"))
		content = append(content, "")
		content = append(content, fmt.Sprintf("Enter IRC server address (default: %s)", defaultServer))
		content = append(content, helpStyle.Render("Format: server.com:port (e.g., irc.libera.chat:6697 for SSL, irc.libera.chat:6667 for non-SSL)"))

	case setupNick:
		content = append(content, systemMessageStyle.Render("Step 2/4: Nickname"))
		content = append(content, "")
		content = append(content, fmt.Sprintf("Server: %s (SSL: %v)", m.config.Server, m.config.UseSSL))
		content = append(content, "")
		content = append(content, fmt.Sprintf("Enter your nickname (default: %s)", defaultNick))

	case setupChannels:
		content = append(content, systemMessageStyle.Render("Step 3/4: Channels"))
		content = append(content, "")
		content = append(content, fmt.Sprintf("Server: %s (SSL: %v)", m.config.Server, m.config.UseSSL))
		content = append(content, fmt.Sprintf("Nickname: %s", m.config.Nick))
		content = append(content, "")
		content = append(content, fmt.Sprintf("Enter channels to join (default: %s)", defaultChannel))
		content = append(content, helpStyle.Render("Separate multiple channels with commas (e.g., #general,#random,#help)"))

	case setupConfirm:
		content = append(content, systemMessageStyle.Render("Step 4/4: Confirmation"))
		content = append(content, "")
		content = append(content, "Configuration Summary:")
		content = append(content, fmt.Sprintf("  Server: %s (SSL: %v)", m.config.Server, m.config.UseSSL))
		content = append(content, fmt.Sprintf("  Nickname: %s", m.config.Nick))
		content = append(content, fmt.Sprintf("  Channels: %s", strings.Join(m.config.Channels, ", ")))
		content = append(content, "")
		content = append(content, helpStyle.Render("Press Enter to connect, or type 'r' to restart setup"))
	}

	content = append(content, "")
	content = append(content, inputBoxFocusedStyle.Render(m.textarea.View()))
	content = append(content, "")
	content = append(content, helpStyle.Render("Ctrl+C to exit"))

	return strings.Join(content, "\n")
}

func (m model) renderSidebar() string {
	if !m.showSidebar {
		return ""
	}

	var content []string
	sidebarHeight := m.height - headerHeight - footerHeight - statusHeight - 2

	// Sidebar header
	content = append(content, sidebarHeaderStyle.Render("IRC CLIENT"))
	content = append(content, "")

	// Connection status
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

	// User and server info
	if m.connected && m.currentNick != "" {
		content = append(content, sidebarItemStyle.Render(fmt.Sprintf("User: %s", m.currentNick)))
		serverName := strings.Split(m.config.Server, ":")[0]
		if len(serverName) > 18 {
			serverName = serverName[:15] + "..."
		}
		content = append(content, sidebarItemStyle.Render(fmt.Sprintf("Server: %s", serverName)))
		content = append(content, "")
	}

	// Channels section
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

	// Fill remaining space or truncate if too long
	maxLines := sidebarHeight - 2
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
