package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func initialModel() model {
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
		textarea:   ta,
		messages:   []string{},
		viewport:   vp,
		ready:      false,
		connected:  false,
		width:      minWidth,
		height:     minHeight,
		state:      stateSetup,
		setupPhase: setupServer,
		config: ircConfig{
			Server:   defaultServer,
			Nick:     defaultNick,
			Channels: []string{defaultChannel},
			UseSSL:   true,
		},
		setupPrompt:      "",
		autoJoinChannels: []string{},
		channels:         make(map[string]*channelData),
		channelOrder:     []string{},
		activeChannels:   []string{},
		showSidebar:      true,
		sidebarWidth:     30,
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
			m.config.Server = input
		}
		m.setupPhase = setupNick
		m.textarea.SetValue("")

	case setupNick:
		if input != "" {
			m.config.Nick = input
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
			m.config.Channels = channels
		}
		m.setupPhase = setupConfirm
		m.textarea.SetValue("")

	case setupConfirm:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" || input == "" {
			m.state = stateConnecting
			m.textarea.SetValue("")
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
		m.currentNick = m.config.Nick

	case ircConnectedMsg:
		m.connected = true
		m.state = stateConnected
		m.addMessage(formatSystemMessage("Connected to IRC server"))

		for _, channel := range m.config.Channels {
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

	case "/msg":
		if len(parts) >= 3 && m.ircClient != nil {
			target := parts[1]
			message := strings.Join(parts[2:], " ")
			m.ircClient.Privmsg(target, message)
			displayMsg := formatUserMessage(m.currentNick, message)
			m.addMessage(displayMsg)
		}

	default:
		m.addMessage(formatErrorMessage("Unknown command: " + command))
	}
}
