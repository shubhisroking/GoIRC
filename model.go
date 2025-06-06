package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func initialModel() model {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		// Fallback to default config if loading fails
		config = DefaultConfig()
	}

	// Initialize logger
	logger, err := NewLogger(config)
	if err != nil {
		// If logger creation fails, create a dummy logger
		logger = &Logger{config: config}
	}

	ta := textarea.New()
	ta.Focus()
	ta.Prompt = "▶ "
	ta.CharLimit = 500
	ta.SetWidth(minWidth)
	ta.SetHeight(1)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(minWidth, 10)

	return model{
		textarea:         ta,
		messages:         []string{},
		viewport:         vp,
		ready:            false,
		connected:        false,
		width:            minWidth,
		height:           minHeight,
		state:            stateSetup,
		setupPhase:       setupServer,
		config:           config,
		setupPrompt:      "",
		autoJoinChannels: []string{},
		channels:         make(map[string]*channelData),
		channelOrder:     []string{},
		activeChannels:   []string{},
		showSidebar:      config.UI.ShowSidebar,
		sidebarWidth:     config.UI.SidebarWidth,
		logger:           logger,
		// Initialize command palette
		commandPaletteVisible:  false,
		commandPaletteQuery:    "",
		commandPaletteSelected: 0,
		commandPaletteItems:    initCommandPaletteItems(),
		commandPaletteFiltered: []commandPaletteItem{},
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *model) handleSetupInput(input string) tea.Cmd {
	input = strings.TrimSpace(input)

	// Clear previous validation error
	m.setupValidationError = ""

	switch m.setupPhase {
	case setupServer:
		if input == "" {
			// Use default server
			m.config.IRC.Server = defaultServer
		} else {
			// Validate server format
			if !m.validateServerFormat(input) {
				m.setupValidationError = "Invalid server format. Use: hostname:port (e.g., irc.libera.chat:6697)"
				return nil
			}
			m.config.IRC.Server = input
		}

		// Determine SSL based on port
		if strings.Contains(m.config.IRC.Server, ":6697") || strings.Contains(m.config.IRC.Server, ":+6697") {
			m.config.IRC.UseSSL = true
		} else {
			m.config.IRC.UseSSL = false
		}

		m.setupPhase = setupNick
		m.textarea.SetValue("")

	case setupNick:
		if input == "" {
			// Use default nick
			m.config.IRC.Nick = defaultNick
		} else {
			// Validate nickname
			if !m.validateNickname(input) {
				m.setupValidationError = "Invalid nickname. Use 3-16 characters, letters, numbers, - and _ only"
				return nil
			}
			m.config.IRC.Nick = input
		}
		m.setupPhase = setupChannels
		m.textarea.SetValue("")

	case setupChannels:
		if input == "" {
			// Use default channel
			m.config.IRC.Channels = []string{defaultChannel}
		} else {
			// Validate and process channels
			channels := strings.Split(input, ",")
			var validChannels []string

			for _, ch := range channels {
				ch = strings.TrimSpace(ch)
				if ch == "" {
					continue
				}

				// Add # prefix if missing
				if !strings.HasPrefix(ch, "#") {
					ch = "#" + ch
				}

				// Validate channel name
				if !m.validateChannelName(ch) {
					m.setupValidationError = fmt.Sprintf("Invalid channel name: %s. Use letters, numbers, - and _ only", ch)
					return nil
				}

				validChannels = append(validChannels, ch)
			}

			if len(validChannels) == 0 {
				m.setupValidationError = "Please enter at least one valid channel"
				return nil
			}

			m.config.IRC.Channels = validChannels
		}
		m.setupPhase = setupConfirm
		m.textarea.SetValue("")

	case setupConfirm:
		if strings.ToLower(input) == "r" || strings.ToLower(input) == "restart" {
			// Restart setup
			m.setupPhase = setupServer
			m.setupValidationError = ""
			m.textarea.SetValue("")
		} else if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" || input == "" {
			m.state = stateConnecting
			m.setupValidationError = ""
			m.textarea.SetValue("")
			// Save configuration after setup is complete
			m.saveConfig()
			return m.connectToIRC()
		} else if strings.ToLower(input) == "n" || strings.ToLower(input) == "no" {
			m.setupPhase = setupServer
			m.setupValidationError = ""
			m.textarea.SetValue("")
		} else {
			m.setupValidationError = "Please type 'y' to connect, 'n' to go back, or 'r' to restart"
		}
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		UpdateStyleWidths(m.width)

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight-statusHeight)
			m.viewport.YPosition = headerHeight + statusHeight
			m.ready = true
		} else {
			viewportWidth := msg.Width
			if m.showSidebar {
				viewportWidth -= m.sidebarWidth
			}
			m.viewport.Width = viewportWidth
			m.viewport.Height = msg.Height - headerHeight - footerHeight - statusHeight
		}

		textareaWidth := msg.Width
		if m.showSidebar {
			textareaWidth -= m.sidebarWidth
		}
		m.textarea.SetWidth(textareaWidth - 4)
	}

	if m.state == stateSetup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyEnter:
				value := m.textarea.Value()
				return m, m.handleSetupInput(value)
			case tea.KeyShiftTab:
				// Go back to previous step
				if m.setupPhase > setupServer {
					m.setupPhase--
					m.setupValidationError = ""
					m.textarea.SetValue("")
				}
			case tea.KeyF1:
				// Show help for current step
				m.showSetupHelp()
			}
		}

		m.textarea, tiCmd = m.textarea.Update(msg)
		return m, tiCmd
	}

	switch msg := msg.(type) {
	case ircClientReadyMsg:
		m.ircClient = msg.client
		m.currentNick = m.config.IRC.Nick

	case ircConnectedMsg:
		m.connected = true
		m.state = stateConnected
		m.addMessage(formatSystemMessage("Connected to IRC server"))

		for _, channel := range m.config.IRC.Channels {
			m.autoJoinChannels = append(m.autoJoinChannels, channel)
			m.addChannel(channel)
			if m.ircClient != nil {
				m.ircClient.Join(channel)
			}
		}

	case ircDisconnectedMsg:
		m.connected = false
		m.state = stateSetup
		m.addMessage(formatErrorMessage("Disconnected from IRC server"))

	case ircMessageMsg:
		m.addMessage(string(msg))

	case ircPrivmsgMsg:
		message := formatUserMessageWithContext(msg.user, msg.message, m.currentNick)
		m.addMessageToChannel(msg.channel, message)

		if msg.channel == m.currentChannel {
			m.addMessage(message)
		}

	case ircErrorMsg:
		m.err = msg.err
		m.addMessage(formatErrorMessage(msg.err.Error()))

	case ircNickChangeMsg:
		if msg.oldNick == m.currentNick {
			m.currentNick = msg.newNick
		}
		message := formatSystemMessage(msg.oldNick + " is now known as " + msg.newNick)
		m.addMessage(message)

	case ircJoinMsg:
		message := formatJoinMessage(msg.user, msg.channel)

		if msg.user == m.currentNick {
			m.setChannelJoined(msg.channel, true)
			m.switchToChannel(msg.channel)
		}

		m.addMessageToChannel(msg.channel, message)

		if msg.channel == m.currentChannel {
			m.addMessage(message)
		}

	case tea.KeyMsg:
		// Handle command palette first if it's visible
		if m.commandPaletteVisible {
			switch msg.Type {
			case tea.KeyEsc:
				m.commandPaletteVisible = false
				m.commandPaletteQuery = ""
				m.commandPaletteSelected = 0
				return m, nil
			case tea.KeyCtrlP:
				// Toggle command palette off
				m.commandPaletteVisible = false
				m.commandPaletteQuery = ""
				m.commandPaletteSelected = 0
				return m, nil
			case tea.KeyEnter:
				if len(m.commandPaletteFiltered) > 0 && m.commandPaletteSelected < len(m.commandPaletteFiltered) {
					selectedItem := m.commandPaletteFiltered[m.commandPaletteSelected]
					m.commandPaletteVisible = false
					m.commandPaletteQuery = ""
					m.commandPaletteSelected = 0
					return m, m.executeCommandPaletteItem(selectedItem)
				}
			case tea.KeyUp, tea.KeyCtrlK:
				if m.commandPaletteSelected > 0 {
					m.commandPaletteSelected--
				} else {
					// Wrap to bottom
					if len(m.commandPaletteFiltered) > 0 {
						m.commandPaletteSelected = len(m.commandPaletteFiltered) - 1
					}
				}
			case tea.KeyDown, tea.KeyCtrlJ:
				if m.commandPaletteSelected < len(m.commandPaletteFiltered)-1 {
					m.commandPaletteSelected++
				} else {
					// Wrap to top
					m.commandPaletteSelected = 0
				}
			case tea.KeyCtrlU:
				// Clear search query
				m.commandPaletteQuery = ""
				m.commandPaletteSelected = 0
				m.filterCommandPalette()
			case tea.KeyBackspace:
				if len(m.commandPaletteQuery) > 0 {
					m.commandPaletteQuery = m.commandPaletteQuery[:len(m.commandPaletteQuery)-1]
					m.filterCommandPalette()
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.commandPaletteQuery += string(msg.Runes)
					m.filterCommandPalette()
				}
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlP:
			// Open command palette
			m.commandPaletteVisible = true
			m.commandPaletteQuery = ""
			m.commandPaletteSelected = 0
			m.filterCommandPalette()
			return m, nil

		case tea.KeyTab:
			m.nextChannel()

		case tea.KeyShiftTab:
			m.prevChannel()

		case tea.KeyCtrlB:
			m.showSidebar = !m.showSidebar
			m.updateDimensions()

		case tea.KeyCtrlN:
			m.nextChannel()

		case tea.KeyEnter:
			if m.textarea.Focused() {
				input := strings.TrimSpace(m.textarea.Value())
				if input == "" {
					break
				}

				if strings.HasPrefix(input, "/") {
					m.handleCommand(input)
				} else {
					if m.currentChannel != "" && m.ircClient != nil {
						m.ircClient.Privmsg(m.currentChannel, input)
						message := formatUserMessage(m.currentNick, input)
						m.addMessageToChannel(m.currentChannel, message)
						m.addMessage(message)
						// Log the sent message
						m.logger.LogIRCMessage(m.currentChannel, m.currentNick, input)
					}
				}

				m.textarea.SetValue("")
			}

		case tea.KeyCtrlU:
			m.textarea.SetValue("")
		}
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *model) addMessage(msg string) {
	m.messages = append(m.messages, msg)
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
	m.viewport.GotoBottom()
}

func (m *model) updateDimensions() {
	viewportWidth := m.width
	if m.showSidebar {
		viewportWidth -= m.sidebarWidth
	}
	m.viewport.Width = viewportWidth
	textareaWidth := viewportWidth - 4
	m.textarea.SetWidth(textareaWidth)
}

func (m *model) handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	switch command {
	case "/help", "/h":
		helpText := []string{
			"Available commands:",
			"/join <#channel> - Join a channel",
			"/part [#channel] - Leave current channel or specified channel",
			"/switch <#channel> - Switch to a channel (or /sw)",
			"/nick <nickname> - Change nickname",
			"/msg <user> <message> - Send private message",
			"/config [show|save|reload] - Manage configuration",
			"/logging [on|off|debug on|off|status] - Control logging",
			"/quit [reason] - Quit IRC",
			"/help - Show this help",
			"",
			"Key bindings:",
			"Tab - Switch to next channel",
			"Shift+Tab - Switch to previous channel",
			"Ctrl+B - Toggle sidebar",
			"Ctrl+C - Exit application",
		}
		for _, line := range helpText {
			m.addMessage(formatSystemMessage(line))
		}

	case "/join":
		if len(parts) >= 2 {
			channel := parts[1]
			if !strings.HasPrefix(channel, "#") {
				channel = "#" + channel
			}
			m.addChannel(channel)
			if m.ircClient != nil {
				m.ircClient.Join(channel)
			}
		}

	case "/part", "/leave":
		channel := m.currentChannel
		if len(parts) >= 2 {
			channel = parts[1]
			if !strings.HasPrefix(channel, "#") {
				channel = "#" + channel
			}
		}
		if channel != "" && m.ircClient != nil {
			m.ircClient.Part(channel)
			m.setChannelJoined(channel, false)
		}

	case "/nick":
		if len(parts) >= 2 && m.ircClient != nil {
			m.ircClient.Nick(parts[1])
		}

	case "/quit":
		if m.ircClient != nil {
			reason := "Leaving"
			if len(parts) >= 2 {
				reason = strings.Join(parts[1:], " ")
			}
			m.ircClient.Quit(reason)
		}

	case "/switch", "/sw":
		if len(parts) >= 2 {
			channelName := parts[1]
			if !strings.HasPrefix(channelName, "#") {
				channelName = "#" + channelName
			}
			if channel, exists := m.channels[channelName]; exists && channel.joined {
				m.switchToChannel(channelName)
			} else {
				m.addMessage(formatErrorMessage(fmt.Sprintf("Channel %s not found or not joined", channelName)))
			}
		} else {
			// Show available channels
			joinedChannels := m.getJoinedChannels()
			if len(joinedChannels) > 0 {
				channelList := strings.Join(joinedChannels, ", ")
				m.addMessage(formatSystemMessage(fmt.Sprintf("Available channels: %s", channelList)))
			} else {
				m.addMessage(formatSystemMessage("No channels joined"))
			}
		}

	case "/msg":
		if len(parts) >= 3 && m.ircClient != nil {
			target := parts[1]
			message := strings.Join(parts[2:], " ")
			m.ircClient.Privmsg(target, message)
			displayMsg := formatUserMessage(m.currentNick, message)
			m.addMessage(displayMsg)
			// Log the private message
			m.logger.LogIRCMessage(target, m.currentNick, message)
		}

	case "/config":
		if len(parts) >= 2 {
			switch parts[1] {
			case "show":
				m.addMessage(formatSystemMessage(fmt.Sprintf("Config file: %s", m.config.FilePath)))
				m.addMessage(formatSystemMessage(fmt.Sprintf("Server: %s", m.config.IRC.Server)))
				m.addMessage(formatSystemMessage(fmt.Sprintf("Nick: %s", m.config.IRC.Nick)))
				m.addMessage(formatSystemMessage(fmt.Sprintf("Channels: %s", strings.Join(m.config.IRC.Channels, ", "))))
				m.addMessage(formatSystemMessage(fmt.Sprintf("SSL: %v", m.config.IRC.UseSSL)))
				m.addMessage(formatSystemMessage(fmt.Sprintf("Logging: %v (Max: %d KB)", m.config.Logging.Enabled, m.config.Logging.MaxSizeKB)))
			case "save":
				m.saveConfig()
				m.addMessage(formatSystemMessage("Configuration saved"))
			case "reload":
				if newConfig, err := LoadConfig(); err != nil {
					m.addMessage(formatErrorMessage(fmt.Sprintf("Failed to reload config: %v", err)))
				} else {
					m.config = newConfig
					m.addMessage(formatSystemMessage("Configuration reloaded"))
				}
			default:
				m.addMessage(formatSystemMessage("Usage: /config [show|save|reload]"))
			}
		} else {
			m.addMessage(formatSystemMessage("Usage: /config [show|save|reload]"))
		}

	case "/logging", "/log":
		if len(parts) >= 2 {
			switch strings.ToLower(parts[1]) {
			case "on", "enable", "true":
				m.config.Logging.Enabled = true
				m.saveConfig()
				m.addMessage(formatSystemMessage("Logging enabled"))
			case "off", "disable", "false":
				m.config.Logging.Enabled = false
				m.saveConfig()
				m.addMessage(formatSystemMessage("Logging disabled"))
			case "debug":
				if len(parts) >= 3 {
					switch strings.ToLower(parts[2]) {
					case "on", "enable", "true":
						m.config.Logging.DebugMode = true
						m.saveConfig()
						m.addMessage(formatSystemMessage("Debug logging enabled"))
					case "off", "disable", "false":
						m.config.Logging.DebugMode = false
						m.saveConfig()
						m.addMessage(formatSystemMessage("Debug logging disabled"))
					default:
						m.addMessage(formatSystemMessage("Usage: /logging debug [on|off]"))
					}
				} else {
					m.addMessage(formatSystemMessage(fmt.Sprintf("Debug logging: %v", m.config.Logging.DebugMode)))
				}
			case "status", "show":
				m.addMessage(formatSystemMessage(fmt.Sprintf("Logging: %v", m.config.Logging.Enabled)))
				m.addMessage(formatSystemMessage(fmt.Sprintf("Debug: %v", m.config.Logging.DebugMode)))
				m.addMessage(formatSystemMessage(fmt.Sprintf("Log path: %s", m.config.Logging.LogPath)))
				m.addMessage(formatSystemMessage(fmt.Sprintf("Max size: %d KB", m.config.Logging.MaxSizeKB)))
			default:
				m.addMessage(formatSystemMessage("Usage: /logging [on|off|debug on|off|status]"))
			}
		} else {
			m.addMessage(formatSystemMessage(fmt.Sprintf("Logging: %v (use '/logging on' or '/logging off' to toggle)", m.config.Logging.Enabled)))
		}

	default:
		m.addMessage(formatErrorMessage("Unknown command: " + command))
	}
}

func (m *model) showSetupHelp() {
	// Add contextual help message based on current step
	switch m.setupPhase {
	case setupServer:
		m.setupValidationError = "💡 Enter server:port (e.g., irc.libera.chat:6697). SSL auto-detected on port 6697"
	case setupNick:
		m.setupValidationError = "💡 Nickname: 3-16 chars, letters/numbers only, must start with letter or _"
	case setupChannels:
		m.setupValidationError = "💡 Channels: comma-separated list (e.g., general,help,dev). # is added automatically"
	case setupConfirm:
		m.setupValidationError = "💡 Press Enter to connect, 'n' to go back, or 'r' to restart setup"
	}
}

// saveConfig saves the current configuration to disk
func (m *model) saveConfig() error {
	if m.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Ensure directories exist
	configDir := filepath.Dir(m.config.FilePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return m.config.Save()
}

// Validation methods for setup wizard
func (m *model) validateServerFormat(server string) bool {
	// Basic validation: must contain hostname and port
	if !strings.Contains(server, ":") {
		return false
	}

	parts := strings.Split(server, ":")
	if len(parts) != 2 {
		return false
	}

	hostname := parts[0]
	port := parts[1]

	// Check hostname is not empty
	if len(hostname) == 0 {
		return false
	}

	// Check port is numeric and within valid range
	if len(port) == 0 {
		return false
	}

	// Remove + prefix for SSL ports if present
	port = strings.TrimPrefix(port, "+")

	// Simple numeric check
	for _, char := range port {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func (m *model) validateNickname(nick string) bool {
	// IRC nickname validation: 3-16 characters, alphanumeric, - and _
	if len(nick) < 3 || len(nick) > 16 {
		return false
	}

	for i, char := range nick {
		if i == 0 {
			// First character must be letter or _
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_') {
				return false
			}
		} else {
			// Subsequent characters can be alphanumeric, - or _
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') || char == '_' || char == '-') {
				return false
			}
		}
	}

	return true
}

func (m *model) validateChannelName(channel string) bool {
	// IRC channel validation: must start with #, then alphanumeric, - and _
	if !strings.HasPrefix(channel, "#") {
		return false
	}

	if len(channel) < 2 || len(channel) > 50 {
		return false
	}

	// Check characters after #
	name := channel[1:]
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-') {
			return false
		}
	}

	return true
}

func initCommandPaletteItems() []commandPaletteItem {
	return []commandPaletteItem{
		// Channel Management
		{name: "Join Channel", description: "Join a new IRC channel", command: "/join", category: "Channels", icon: "+", shortcut: "", priority: 90},
		{name: "Part Channel", description: "Leave the current channel", command: "/part", category: "Channels", icon: "-", shortcut: "", priority: 80},
		{name: "Switch Channel", description: "Switch to a different channel", command: "/switch", category: "Channels", icon: "~", shortcut: "Tab", priority: 95},
		{name: "Next Channel", description: "Switch to the next channel", command: "next_channel", category: "Navigation", icon: ">", shortcut: "Tab", priority: 85},
		{name: "Previous Channel", description: "Switch to the previous channel", command: "prev_channel", category: "Navigation", icon: "<", shortcut: "Shift+Tab", priority: 85},
		{name: "List Channels", description: "Show all joined channels", command: "list_channels", category: "Channels", icon: "=", shortcut: "", priority: 65},

		// Communication
		{name: "Send Private Message", description: "Send a private message to a user", command: "/msg", category: "Communication", icon: "@", shortcut: "", priority: 75},
		{name: "Change Nickname", description: "Change your nickname", command: "/nick", category: "User", icon: "*", shortcut: "", priority: 70},

		// Interface
		{name: "Toggle Sidebar", description: "Show/hide the channel sidebar", command: "toggle_sidebar", category: "Interface", icon: "|", shortcut: "Ctrl+B", priority: 60},
		{name: "Command Palette", description: "Open command palette", command: "command_palette", category: "Interface", icon: ".", shortcut: "Ctrl+P", priority: 100},
		{name: "Clear Screen", description: "Clear the chat messages", command: "clear_screen", category: "Interface", icon: "x", shortcut: "", priority: 55},

		// Information
		{name: "Show Help", description: "Display available commands", command: "/help", category: "Help", icon: "?", shortcut: "", priority: 50},
		{name: "Show Configuration", description: "Display current configuration", command: "/config show", category: "Configuration", icon: "&", shortcut: "", priority: 40},
		{name: "Show Logging Status", description: "Display logging information", command: "/logging status", category: "Logging", icon: "#", shortcut: "", priority: 30},
		{name: "Connection Status", description: "Show connection information", command: "connection_status", category: "Information", icon: "!", shortcut: "", priority: 45},

		// Configuration
		{name: "Save Configuration", description: "Save current configuration to file", command: "/config save", category: "Configuration", icon: "s", shortcut: "", priority: 35},
		{name: "Reload Configuration", description: "Reload configuration from file", command: "/config reload", category: "Configuration", icon: "r", shortcut: "", priority: 25},
		{name: "Enable Logging", description: "Turn on IRC logging", command: "/logging on", category: "Logging", icon: "^", shortcut: "", priority: 20},
		{name: "Disable Logging", description: "Turn off IRC logging", command: "/logging off", category: "Logging", icon: "v", shortcut: "", priority: 15},

		// System
		{name: "Quit IRC", description: "Disconnect and exit the application", command: "/quit", category: "System", icon: "q", shortcut: "Ctrl+C", priority: 10},
	}
}

func (m *model) getDynamicCommandPaletteItems() []commandPaletteItem {
	var dynamicItems []commandPaletteItem

	// Add channel-specific commands based on joined channels
	if m.connected {
		joinedChannels := m.getJoinedChannels()

		// Add switch commands for each joined channel
		for _, channel := range joinedChannels {
			if channel != m.currentChannel {
				dynamicItems = append(dynamicItems, commandPaletteItem{
					name:        "Switch to " + channel,
					description: "Switch to channel " + channel,
					command:     "/switch " + channel,
					category:    "Quick Switch",
					icon:        "~",
					shortcut:    "",
					priority:    88,
				})
			}
		}

		// Add part command for current channel
		if m.currentChannel != "" {
			dynamicItems = append(dynamicItems, commandPaletteItem{
				name:        "Part " + m.currentChannel,
				description: "Leave channel " + m.currentChannel,
				command:     "/part " + m.currentChannel,
				category:    "Current Channel",
				icon:        "🚪",
				shortcut:    "",
				priority:    85,
			})
		}

		// Add reconnect command if disconnected
	} else {
		dynamicItems = append(dynamicItems, commandPaletteItem{
			name:        "Reconnect",
			description: "Reconnect to IRC server",
			command:     "reconnect",
			category:    "Connection",
			icon:        "🔌",
			shortcut:    "",
			priority:    95,
		})
	}

	// Add commonly used channels for quick joining
	commonChannels := []string{"#general", "#help", "#random", "#dev", "#announcements"}
	joinedChannelMap := make(map[string]bool)
	for _, ch := range m.getJoinedChannels() {
		joinedChannelMap[ch] = true
	}

	for _, channel := range commonChannels {
		if !joinedChannelMap[channel] {
			dynamicItems = append(dynamicItems, commandPaletteItem{
				name:        "Join " + channel,
				description: "Join channel " + channel,
				command:     "/join " + channel,
				category:    "Quick Join",
				icon:        "📥",
				shortcut:    "",
				priority:    75,
			})
		}
	}

	return dynamicItems
}

func (m *model) getAllCommandPaletteItems() []commandPaletteItem {
	// Combine static and dynamic items
	allItems := append([]commandPaletteItem{}, m.commandPaletteItems...)
	dynamicItems := m.getDynamicCommandPaletteItems()
	allItems = append(allItems, dynamicItems...)
	return allItems
}

func (m *model) filterCommandPalette() {
	query := strings.ToLower(m.commandPaletteQuery)
	m.commandPaletteFiltered = []commandPaletteItem{}

	// Get all available items (static + dynamic)
	allItems := m.getAllCommandPaletteItems()

	// If no query, show all items sorted by priority
	if query == "" {
		m.commandPaletteFiltered = append(m.commandPaletteFiltered, allItems...)
	} else {
		// Score-based fuzzy search
		type scoredItem struct {
			item  commandPaletteItem
			score int
		}

		var scored []scoredItem

		for _, item := range allItems {
			score := 0
			itemName := strings.ToLower(item.name)
			itemDesc := strings.ToLower(item.description)
			itemCategory := strings.ToLower(item.category)

			// Exact match in name gets highest score
			if strings.Contains(itemName, query) {
				score += 100
				if strings.HasPrefix(itemName, query) {
					score += 50 // Prefix match bonus
				}
			}

			// Match in description
			if strings.Contains(itemDesc, query) {
				score += 30
			}

			// Match in category
			if strings.Contains(itemCategory, query) {
				score += 20
			}

			// Fuzzy matching - check for character sequence
			if fuzzyMatch(itemName, query) {
				score += 15
			}

			// Add priority bonus
			score += item.priority / 10

			if score > 0 {
				scored = append(scored, scoredItem{item: item, score: score})
			}
		}

		// Sort by score (descending)
		for i := 0; i < len(scored); i++ {
			for j := i + 1; j < len(scored); j++ {
				if scored[i].score < scored[j].score {
					scored[i], scored[j] = scored[j], scored[i]
				}
			}
		}

		// Extract items
		for _, scored := range scored {
			m.commandPaletteFiltered = append(m.commandPaletteFiltered, scored.item)
		}
	}

	// Reset selection if it's out of bounds
	if m.commandPaletteSelected >= len(m.commandPaletteFiltered) {
		m.commandPaletteSelected = 0
	}
}

// Simple fuzzy matching function
func fuzzyMatch(text, pattern string) bool {
	if pattern == "" {
		return true
	}

	textRunes := []rune(text)

	textIndex := 0
	for _, p := range pattern {
		found := false
		for textIndex < len(textRunes) {
			if textRunes[textIndex] == p {
				found = true
				textIndex++
				break
			}
			textIndex++
		}
		if !found {
			return false
		}
	}
	return true
}

func (m *model) executeCommandPaletteItem(item commandPaletteItem) tea.Cmd {
	// Add to recently used commands (you could implement this later)

	switch item.command {
	case "toggle_sidebar":
		m.showSidebar = !m.showSidebar
		m.updateDimensions()
		m.addMessage(formatSystemMessage("Sidebar toggled"))

	case "next_channel":
		m.nextChannel()

	case "prev_channel":
		m.prevChannel()
	case "command_palette":
		// This shouldn't happen since we're already in the palette, but just in case
		m.commandPaletteVisible = true
		m.commandPaletteQuery = ""
		m.commandPaletteSelected = 0
		m.filterCommandPalette()

	case "reconnect":
		if !m.connected {
			m.state = stateConnecting
			return m.connectToIRC()
		} else {
			m.addMessage(formatSystemMessage("Already connected to IRC server"))
		}

	case "list_channels":
		joinedChannels := m.getJoinedChannels()
		if len(joinedChannels) > 0 {
			m.addMessage(formatSystemMessage("📋 Joined Channels:"))
			for i, channel := range joinedChannels {
				indicator := "  "
				if channel == m.currentChannel {
					indicator = "➤ "
				}
				m.addMessage(formatSystemMessage(fmt.Sprintf("%s%d. %s", indicator, i+1, channel)))
			}
		} else {
			m.addMessage(formatSystemMessage("No channels joined"))
		}

	case "clear_screen":
		m.messages = []string{}
		if m.currentChannel != "" {
			if channel, exists := m.channels[m.currentChannel]; exists {
				channel.messages = []string{}
			}
		}
		m.addMessage(formatSystemMessage("Screen cleared"))

	case "connection_status":
		if m.connected {
			uptime := time.Since(m.connectionTime).Truncate(time.Second)
			m.addMessage(formatSystemMessage("🔌 Connection Status: Connected"))
			m.addMessage(formatSystemMessage(fmt.Sprintf("Server: %s", m.config.IRC.Server)))
			m.addMessage(formatSystemMessage(fmt.Sprintf("Nickname: %s", m.currentNick)))
			m.addMessage(formatSystemMessage(fmt.Sprintf("Current Channel: %s", m.currentChannel)))
			m.addMessage(formatSystemMessage(fmt.Sprintf("Uptime: %v", uptime)))
		} else {
			m.addMessage(formatSystemMessage("🔌 Connection Status: Disconnected"))
		}

	default: // Handle IRC commands directly
		if strings.HasPrefix(item.command, "/") {
			// Execute the command directly instead of just setting it in textarea
			switch {
			case item.command == "/help":
				m.handleCommand("/help")

			case item.command == "/quit":
				if m.ircClient != nil {
					m.ircClient.Quit("Leaving via command palette")
				}
				return tea.Quit

			case strings.HasPrefix(item.command, "/config"):
				m.handleCommand(item.command)

			case strings.HasPrefix(item.command, "/logging"):
				m.handleCommand(item.command)

			case strings.HasPrefix(item.command, "/switch "):
				// Handle dynamic switch commands
				m.handleCommand(item.command)

			case strings.HasPrefix(item.command, "/join "):
				// Handle dynamic join commands
				m.handleCommand(item.command)

			case strings.HasPrefix(item.command, "/part "):
				// Handle dynamic part commands
				m.handleCommand(item.command)

			case item.command == "/join":
				// For commands that need user input, set them in textarea with a space
				m.textarea.SetValue(item.command + " ")
				m.textarea.Focus()

			case item.command == "/part":
				// Part current channel if no arguments, otherwise set in textarea
				if m.currentChannel != "" {
					m.handleCommand("/part " + m.currentChannel)
				} else {
					m.textarea.SetValue(item.command + " ")
					m.textarea.Focus()
				}

			case item.command == "/switch":
				// Show available channels or set in textarea for manual input
				joinedChannels := m.getJoinedChannels()
				if len(joinedChannels) > 1 {
					channelList := strings.Join(joinedChannels, ", ")
					m.addMessage(formatSystemMessage("Available channels: " + channelList))
					m.addMessage(formatSystemMessage("Use /switch <channel> or Tab/Shift+Tab to navigate"))
				}
				m.textarea.SetValue(item.command + " ")
				m.textarea.Focus()

			case item.command == "/nick":
				m.textarea.SetValue(item.command + " ")
				m.textarea.Focus()

			case item.command == "/msg":
				m.textarea.SetValue(item.command + " ")
				m.textarea.Focus()

			default:
				// For any other IRC command, execute it directly
				m.handleCommand(item.command)
			}
		}
	}
	return nil
}
