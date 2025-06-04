package main

import (
	"fmt"
	"strings"

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
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *model) handleSetupInput(input string) tea.Cmd {
	input = strings.TrimSpace(input)

	switch m.setupPhase {
	case setupServer:
		if input != "" {
			m.config.IRC.Server = input
		}
		m.setupPhase = setupNick
		m.textarea.SetValue("")

	case setupNick:
		if input != "" {
			m.config.IRC.Nick = input
		}
		m.setupPhase = setupChannels
		m.textarea.SetValue("")

	case setupChannels:
		if input != "" {
			channels := strings.Split(input, ",")
			for i, ch := range channels {
				channels[i] = strings.TrimSpace(ch)
				if !strings.HasPrefix(channels[i], "#") {
					channels[i] = "#" + channels[i]
				}
			}
			m.config.IRC.Channels = channels
		}
		m.setupPhase = setupConfirm
		m.textarea.SetValue("")

	case setupConfirm:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" || input == "" {
			m.state = stateConnecting
			m.textarea.SetValue("")
			// Save configuration after setup is complete
			m.saveConfig()
			return m.connectToIRC()
		} else if strings.ToLower(input) == "n" || strings.ToLower(input) == "no" {
			m.setupPhase = setupServer
			m.textarea.SetValue("")
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
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyTab:
			m.nextChannel()

		case tea.KeyShiftTab:
			m.prevChannel()

		case tea.KeyCtrlB:
			m.showSidebar = !m.showSidebar
			m.updateDimensions()

		case tea.KeyCtrlN:
			m.nextChannel()

		case tea.KeyCtrlP:
			m.prevChannel()

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

// saveConfig saves the current configuration to disk
func (m *model) saveConfig() {
	if err := m.config.Save(); err != nil {
		m.logger.LogError("Failed to save configuration: %v", err)
	} else {
		m.logger.Log("Configuration saved successfully")
	}
}
