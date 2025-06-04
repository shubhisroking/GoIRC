package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	irc "github.com/fluffle/goirc/client"
)

const (
	defaultServer  = "irc.libera.chat:6697"
	defaultChannel = "#bubbletea-test"
	defaultNick    = "bubbletea-user"
	headerHeight   = 4
	footerHeight   = 4
	statusHeight   = 1
	minWidth       = 80
	minHeight      = 24
)

type appState int

const (
	stateSetup appState = iota
	stateConnecting
	stateConnected
)

type setupPhase int

const (
	setupServer setupPhase = iota
	setupNick
	setupChannels
	setupConfirm
)

var p *tea.Program

type channelData struct {
	name     string
	messages []string
	active   bool
	joined   bool
}

type model struct {
	ircClient      *irc.Conn
	viewport       viewport.Model
	messages       []string
	textarea       textarea.Model
	ready          bool
	err            error
	connectionTime time.Time
	connected      bool
	currentChannel string
	currentNick    string
	width          int
	height         int

	
	channels       map[string]*channelData
	channelOrder   []string
	activeChannels []string

	
	showSidebar  bool
	sidebarWidth int

	state            appState
	setupPhase       setupPhase
	config           ircConfig
	setupPrompt      string
	autoJoinChannels []string
}

type ircConfig struct {
	Server   string
	Nick     string
	Channels []string
	UseSSL   bool
}

type (
	ircMessageMsg      string
	ircPrivmsgMsg      struct{ user, message, channel string }
	ircErrorMsg        struct{ err error }
	ircConnectedMsg    struct{}
	ircDisconnectedMsg struct{}
	ircNickChangeMsg   struct{ oldNick, newNick string }
	ircJoinMsg         struct{ user, channel string }
	ircClientReadyMsg  struct{ client *irc.Conn }
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

func (m *model) connectToIRC() tea.Cmd {
	return func() tea.Msg {
		cfg := irc.NewConfig(m.config.Nick)
		cfg.SSL = m.config.UseSSL

		if m.config.UseSSL {
			serverHost := strings.Split(m.config.Server, ":")[0]
			cfg.SSLConfig = &tls.Config{ServerName: serverHost}
		}

		cfg.Server = m.config.Server
		cfg.NewNick = func(n string) string { return n + "_" }

		c := irc.Client(cfg)

		c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
			log.Println("Connected to IRC")
			log.Printf("Our actual nickname is: %s", conn.Me().Nick)

			
			for _, channel := range m.config.Channels {
				if channel != "" {
					m.addChannel(channel)
					conn.Join(channel)
				}
			}

			if p != nil {
				p.Send(ircConnectedMsg{})
			}
		})

		c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
			log.Println("Disconnected from IRC")
			if p != nil {
				p.Send(ircDisconnectedMsg{})
			}
		})

		c.HandleFunc(irc.NICK, func(conn *irc.Conn, line *irc.Line) {
			oldNick := line.Nick
			newNick := line.Args[0]
			log.Printf("%s changed nick to %s", oldNick, newNick)
			if p != nil {
				p.Send(ircNickChangeMsg{oldNick: oldNick, newNick: newNick})
			}
		})

		c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
			user := line.Nick
			channel := line.Args[0]
			message := line.Args[1]
			log.Printf("Received PRIVMSG: %s from %s in %s", message, user, channel)
			if p != nil {
				p.Send(ircPrivmsgMsg{user: user, message: message, channel: channel})
			}
		})

		c.HandleFunc(irc.NOTICE, func(conn *irc.Conn, line *irc.Line) {
			user := line.Nick
			if user == "" {
				user = line.Host
			}
			message := line.Args[1]
			log.Printf("Received notice: %s from %s", message, user)
			if p != nil {
				p.Send(ircMessageMsg(formatNoticeMessage(user, message)))
			}
		})

		c.HandleFunc(irc.JOIN, func(conn *irc.Conn, line *irc.Line) {
			user := line.Nick
			channel := line.Args[0]
			log.Printf("%s joined %s", user, channel)

			if p != nil {
				p.Send(ircJoinMsg{user: user, channel: channel})
			}
		})

		c.HandleFunc(irc.PART, func(conn *irc.Conn, line *irc.Line) {
			user := line.Nick
			channel := line.Args[0]
			message := ""
			if len(line.Args) > 1 {
				message = line.Args[1]
			}
			log.Printf("%s left %s (%s)", user, channel, message)
			if p != nil {
				p.Send(ircMessageMsg(formatPartMessage(user, channel, message)))
			}
		})

		c.HandleFunc(irc.QUIT, func(conn *irc.Conn, line *irc.Line) {
			user := line.Nick
			message := ""
			if len(line.Args) > 0 {
				message = line.Args[0]
			}
			log.Printf("%s quit (%s)", user, message)
			if p != nil {
				p.Send(ircMessageMsg(formatQuitMessage(user, message)))
			}
		})

		if err := c.Connect(); err != nil {
			log.Printf("Error connecting to IRC: %v", err)
			return ircErrorMsg{err}
		}

		return ircClientReadyMsg{client: c}
	}
}

func (m *model) handleSetupInput(input string) tea.Cmd {
	input = strings.TrimSpace(input)

	switch m.setupPhase {
	case setupServer:
		if input != "" {
			m.config.Server = input
		}

		if strings.Contains(m.config.Server, ":6697") || strings.Contains(m.config.Server, ":7000") {
			m.config.UseSSL = true
		} else if strings.Contains(m.config.Server, ":6667") {
			m.config.UseSSL = false
		}
		m.setupPhase = setupNick
		m.textarea.Placeholder = "Enter your nickname or press Enter for default..."
		m.textarea.Reset()

	case setupNick:
		if input != "" {
			m.config.Nick = input
		}
		m.setupPhase = setupChannels
		m.textarea.Placeholder = "Enter channels to join (comma-separated, e.g., #channel1,#channel2) or press Enter for default..."
		m.textarea.Reset()

	case setupChannels:
		if input != "" {
			channels := strings.Split(input, ",")
			m.config.Channels = []string{}
			for _, ch := range channels {
				ch = strings.TrimSpace(ch)
				if ch != "" {
					if !strings.HasPrefix(ch, "#") {
						ch = "#" + ch
					}
					m.config.Channels = append(m.config.Channels, ch)
				}
			}
		}
		m.setupPhase = setupConfirm
		m.textarea.Placeholder = "Press Enter to connect or 'r' to restart setup..."
		m.textarea.Reset()

	case setupConfirm:
		if strings.ToLower(input) == "r" || strings.ToLower(input) == "restart" {
			m.setupPhase = setupServer
			m.config = ircConfig{
				Server:   defaultServer,
				Nick:     defaultNick,
				Channels: []string{defaultChannel},
				UseSSL:   true,
			}
			m.textarea.Placeholder = "Enter IRC server (e.g., irc.libera.chat:6697) or press Enter for default..."
			m.textarea.Reset()
		} else {
			m.state = stateConnecting
			m.currentChannel = m.config.Channels[0]
			m.currentNick = m.config.Nick
			m.textarea.Placeholder = "💬 Type your message here... (Press Enter to send, Ctrl+C to quit)"
			return m.connectToIRC()
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

		sidebarWidth := 0
		if m.showSidebar {
			sidebarWidth = m.sidebarWidth + 2 
		}

		chatWidth := max(msg.Width-sidebarWidth-4, minWidth-4)
		m.viewport.Width = chatWidth
		m.viewport.Height = max(msg.Height-headerHeight-footerHeight-statusHeight-2, 10)

		m.textarea.SetWidth(chatWidth)

		UpdateStyleWidths(msg.Width)

		m.ready = true
		if m.state == stateConnected || m.state == stateConnecting {
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		}

		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, tea.Batch(tiCmd, vpCmd)
	}

	if m.state == stateSetup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyEnter:
				inputValue := m.textarea.Value()
				cmd := m.handleSetupInput(inputValue)
				return m, cmd
			}
		}

		m.textarea, tiCmd = m.textarea.Update(msg)
		return m, tiCmd
	}

	switch msg := msg.(type) {
	case ircClientReadyMsg:
		m.ircClient = msg.client
		log.Println("IRC client initialized and stored in model")

	case ircConnectedMsg:
		m.connected = true
		m.connectionTime = time.Now()
		m.state = stateConnected

		if m.ircClient != nil && m.ircClient.Connected() {
			actualNick := m.ircClient.Me().Nick
			m.currentNick = actualNick
			log.Printf("Connected with nickname: %s", actualNick)
		}

		channelList := strings.Join(m.config.Channels, ", ")
		systemMsg1 := formatSystemMessage(fmt.Sprintf("✅ Connected to %s", m.config.Server))
		systemMsg2 := formatSystemMessage(fmt.Sprintf("📋 Joining channels: %s", channelList))

		
		m.messages = append(m.messages, systemMsg1, systemMsg2)
		if m.currentChannel != "" {
			m.addMessageToChannel(m.currentChannel, systemMsg1)
			m.addMessageToChannel(m.currentChannel, systemMsg2)
		}

		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case ircDisconnectedMsg:
		m.connected = false
		m.messages = append(m.messages, formatErrorMessage("Disconnected from IRC"))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, tea.Quit

	case ircMessageMsg:
		m.messages = append(m.messages, string(msg))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case ircPrivmsgMsg:
		formattedMsg := formatUserMessageWithContext(msg.user, msg.message, m.currentNick)

		
		if msg.channel != "" {
			m.addMessageToChannel(msg.channel, formattedMsg)

			
			if msg.channel == m.currentChannel {
				m.messages = append(m.messages, formattedMsg)
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.viewport.GotoBottom()
			}
		} else {
			
			m.messages = append(m.messages, formattedMsg)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		}

	case ircErrorMsg:
		m.err = msg.err
		m.messages = append(m.messages, formatErrorMessage(fmt.Sprintf("IRC Error: %v", msg.err)))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, tea.Quit

	case ircNickChangeMsg:
		if msg.oldNick == m.currentNick {
			m.currentNick = msg.newNick
			log.Printf("Our nick changed from %s to %s", msg.oldNick, msg.newNick)
		}

		m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("👤 %s is now known as %s", msg.oldNick, msg.newNick)))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case ircJoinMsg:
		joinMsg := formatJoinMessage(msg.user, msg.channel)

		
		m.addMessageToChannel(msg.channel, joinMsg)

		
		ourNick := m.currentNick
		if m.ircClient != nil && m.ircClient.Connected() {
			ourNick = m.ircClient.Me().Nick
			if ourNick != m.currentNick {
				m.currentNick = ourNick
				log.Printf("Updated our nick to: %s", ourNick)
			}
		}

		
		log.Printf("Processing join message: %s joined %s (our nick: %s)", msg.user, msg.channel, ourNick)
		if strings.EqualFold(msg.user, ourNick) {
			log.Printf("We joined channel: %s, marking as joined", msg.channel)
			m.addChannel(msg.channel)
			m.setChannelJoined(msg.channel, true)

			
			joinedChannels := m.getJoinedChannels()
			if m.currentChannel == "" || len(joinedChannels) == 1 {
				log.Printf("Switching to channel: %s", msg.channel)
				m.switchToChannel(msg.channel)
			}
		}

		
		if msg.channel == m.currentChannel {
			m.messages = append(m.messages, joinMsg)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.ircClient != nil && m.ircClient.Connected() {
				m.ircClient.Quit("Bubble Tea client closing")
			}
			return m, tea.Quit
		case tea.KeyTab:
			
			m.nextChannel()
			return m, nil
		case tea.KeyShiftTab:
			
			m.prevChannel()
			return m, nil
		case tea.KeyF1:
			
			m.showSidebar = !m.showSidebar

			
			sidebarWidth := 0
			if m.showSidebar {
				sidebarWidth = m.sidebarWidth + 2
			}
			chatWidth := max(m.width-sidebarWidth-4, minWidth-4)
			m.viewport.Width = chatWidth
			m.textarea.SetWidth(chatWidth)
			UpdateStyleWidths(m.width)

			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			return m, nil
		case tea.KeyEnter:
			if m.ircClient == nil {
				m.messages = append(m.messages, formatErrorMessage("IRC client not initialized"))
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.viewport.GotoBottom()
				m.textarea.Reset()
				return m, nil
			}

			if !m.ircClient.Connected() {
				m.messages = append(m.messages, formatErrorMessage("Not connected to IRC server"))
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.viewport.GotoBottom()
				m.textarea.Reset()
				return m, nil
			}

			inputValue := strings.TrimSpace(m.textarea.Value())
			if inputValue == "" {
				m.textarea.Reset()
				return m, nil
			}

			log.Printf("Processing input: '%s' (connected: %v, channel: '%s', nick: '%s')",
				inputValue, m.ircClient.Connected(), m.currentChannel, m.currentNick)

			if strings.HasPrefix(inputValue, "/") {
				parts := strings.SplitN(inputValue, " ", 2)
				command := strings.ToUpper(strings.TrimPrefix(parts[0], "/"))
				args := ""
				if len(parts) > 1 {
					args = parts[1]
				}

				log.Printf("Processing command: /%s with args: '%s'", command, args)

				switch command {
				case "JOIN":
					if args != "" {
						log.Printf("Joining channel: %s", args)
						m.addChannel(args)
						m.ircClient.Join(args)
						systemMsg := formatSystemMessage(fmt.Sprintf("🚪 Joining %s...", args))
						m.messages = append(m.messages, systemMsg)
						if m.currentChannel != "" {
							m.addMessageToChannel(m.currentChannel, systemMsg)
						}
					} else {
						m.messages = append(m.messages, formatErrorMessage("Usage: /JOIN #channel"))
					}
				case "PART":
					channelToPart := m.currentChannel
					if args != "" {
						channelToPart = args
					}
					log.Printf("Leaving channel: %s", channelToPart)
					m.ircClient.Part(channelToPart)
					systemMsg := formatSystemMessage(fmt.Sprintf("🚪 Leaving %s...", channelToPart))
					m.messages = append(m.messages, systemMsg)

					
					m.setChannelJoined(channelToPart, false)

					
					joinedChannels := m.getJoinedChannels()
					if channelToPart == m.currentChannel && len(joinedChannels) > 0 {
						m.nextChannel()
					} else if len(joinedChannels) == 0 {
						m.currentChannel = ""
						m.messages = []string{}
						m.viewport.SetContent("")
					}
				case "NICK":
					if args != "" {
						log.Printf("Changing nick to: %s", args)
						m.ircClient.Nick(args)
						m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("👤 Changing nick to %s...", args)))
					} else {
						m.messages = append(m.messages, formatErrorMessage("Usage: /NICK <newnick>"))
					}
				case "QUIT":
					quitMsg := "Leaving"
					if args != "" {
						quitMsg = args
					}
					log.Printf("Quitting with message: %s", quitMsg)
					m.ircClient.Quit(quitMsg)

				case "MSG", "QUERY":
					if strings.Count(args, " ") >= 1 {
						targetAndMsg := strings.SplitN(args, " ", 2)
						target, message := targetAndMsg[0], targetAndMsg[1]
						log.Printf("Sending private message to %s: %s", target, message)
						m.ircClient.Privmsg(target, message)
						m.messages = append(m.messages, formatUserMessageWithContext(m.currentNick, fmt.Sprintf("(to %s) %s", target, message), m.currentNick))
					} else {
						m.messages = append(m.messages, formatErrorMessage("Usage: /MSG <nick> <message>"))
					}
				case "LIST", "LS":

					if args == "" {
						m.ircClient.Raw("LIST")
						m.messages = append(m.messages, formatSystemMessage("📋 Listing all channels"))
					} else {
						m.ircClient.Raw("LIST " + args)
						m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("📋 Listing channels matching: %s", args)))
					}
				case "SWITCH", "SW":
					if args != "" {
						
						channelName := args
						if !strings.HasPrefix(channelName, "#") {
							channelName = "#" + channelName
						}

						
						var foundChannel string
						var foundChannelData *channelData
						for chName, chData := range m.channels {
							if strings.EqualFold(chName, channelName) && chData.joined {
								foundChannel = chName
								foundChannelData = chData
								break
							}
						}

						if foundChannelData != nil {
							m.switchToChannel(foundChannel)
							m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("🔄 Switched to %s", foundChannel)))
						} else {
							
							log.Printf("Switch failed. Available channels:")
							for chName, chData := range m.channels {
								log.Printf("  %s (joined: %v)", chName, chData.joined)
							}
							m.messages = append(m.messages, formatErrorMessage(fmt.Sprintf("Channel %s not found in joined channels", channelName)))
						}
					} else {
						channelsList := m.getChannelsList()
						if len(channelsList) > 0 {
							m.messages = append(m.messages, formatSystemMessage("📋 Available channels:"))
							for _, ch := range channelsList {
								m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("  %s", ch)))
							}
						} else {
							m.messages = append(m.messages, formatSystemMessage("📋 No channels joined"))
						}
					}
				case "HELP":
					helpText := []string{
						"🆘 Available Commands:",
						"  /join #channel  - Join a channel",
						"  /part [#channel] - Leave current or specified channel",
						"  /nick <newnick> - Change nickname",
						"  /msg <nick> <message> - Send private message",
						"  /switch <channel> - Switch to a joined channel",
						"  /list [pattern] - List channels",
						"  /quit [message] - Quit IRC",
						"  /help - Show this help",
						"",
						"🎮 Keyboard Shortcuts:",
						"  F1 - Toggle sidebar",
						"  Tab - Switch to next channel",
						"  Shift+Tab - Switch to previous channel",
						"  Alt+1-9 - Jump to channel number (shown in sidebar)",
						"  Ctrl+C - Quit IRC client",
					}
					for _, line := range helpText {
						m.messages = append(m.messages, formatSystemMessage(line))
					}
				default:
					log.Printf("Sending raw IRC command: %s %s", command, args)
					if args != "" {
						m.ircClient.Raw(command + " " + args)
					} else {
						m.ircClient.Raw(command)
					}
					m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("📡 Sent RAW: /%s %s", command, args)))
				}
			} else {
				if m.currentChannel == "" && len(m.activeChannels) > 0 {
					m.currentChannel = m.activeChannels[0]
				}
				log.Printf("Sending PRIVMSG to %s: '%s' (from nick: %s)", m.currentChannel, inputValue, m.currentNick)

				m.ircClient.Privmsg(m.currentChannel, inputValue)

				
				formattedMsg := formatUserMessageWithContext(m.currentNick, inputValue, m.currentNick)
				m.messages = append(m.messages, formattedMsg)
				m.addMessageToChannel(m.currentChannel, formattedMsg)
				log.Printf("Message added to local display: <%s> %s", m.currentNick, inputValue)
			}

			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			m.textarea.Reset()

			return m, nil
		}

		
		if msg.Alt && len(msg.Runes) > 0 {
			switch msg.Runes[0] {
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				channelNum := int(msg.Runes[0] - '1') 
				joinedChannels := m.getJoinedChannels()
				if channelNum < len(joinedChannels) {
					m.switchToChannel(joinedChannels[channelNum])
				}
				return m, nil
			}
		}
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

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

	
	if m.showSidebar {
		sidebar := m.renderSidebar()
		mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input, help)
		contentArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
		return lipgloss.JoinVertical(lipgloss.Left, header, status, contentArea)
	}

	
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
	
	
	content = append(content, sidebarHeaderStyle.Render("IRC CLIENT"))
	content = append(content, "")
	
	
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
	
	
	if m.connected && m.currentNick != "" {
		content = append(content, sidebarItemStyle.Render(fmt.Sprintf("User: %s", m.currentNick)))
		serverName := strings.Split(m.config.Server, ":")[0]
		if len(serverName) > 18 {
			serverName = serverName[:15] + "..."
		}
		content = append(content, sidebarItemStyle.Render(fmt.Sprintf("Server: %s", serverName)))
		content = append(content, "")
	}
	
	
	joinedChannels := m.getJoinedChannels()
	channelCountBadge := sidebarChannelCountStyle.Render(fmt.Sprintf(" %d ", len(joinedChannels)))
	channelsHeader := fmt.Sprintf("CHANNELS %s", channelCountBadge)
	content = append(content, sidebarSectionStyle.Render(channelsHeader))
	
	if len(joinedChannels) > 0 {
		for i, channelName := range joinedChannels {
			
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
	
	
	maxLines := sidebarHeight - 2
	if len(content) < maxLines {
		for len(content) < maxLines {
			content = append(content, "")
		}
	} else if len(content) > maxLines {
		
		content = content[:maxLines-1]
		content = append(content, sidebarItemStyle.Render("..."))
	}
	
	sidebarContent := strings.Join(content, "\n")
	return sidebarStyle.Height(sidebarHeight).Render(sidebarContent)
}


func (m *model) addChannel(channelName string) {
	log.Printf("Adding channel: %s", channelName)
	if _, exists := m.channels[channelName]; !exists {
		m.channels[channelName] = &channelData{
			name:     channelName,
			messages: []string{},
			active:   false,
			joined:   false,
		}
		m.channelOrder = append(m.channelOrder, channelName)
		log.Printf("Channel %s added successfully", channelName)
	} else {
		log.Printf("Channel %s already exists", channelName)
	}
}

func (m *model) setChannelActive(channelName string, active bool) {
	if channel, exists := m.channels[channelName]; exists {
		channel.active = active
		if active && !m.isChannelInActiveList(channelName) {
			m.activeChannels = append(m.activeChannels, channelName)
		} else if !active {
			for i, ch := range m.activeChannels {
				if ch == channelName {
					m.activeChannels = append(m.activeChannels[:i], m.activeChannels[i+1:]...)
					break
				}
			}
		}
	}
}

func (m *model) setChannelJoined(channelName string, joined bool) {
	log.Printf("Setting channel %s joined status to: %v", channelName, joined)
	if channel, exists := m.channels[channelName]; exists {
		channel.joined = joined
		log.Printf("Channel %s joined status updated successfully", channelName)
	} else {
		log.Printf("Channel %s not found when trying to set joined status", channelName)
	}
}

func (m *model) isChannelInActiveList(channelName string) bool {
	for _, ch := range m.activeChannels {
		if ch == channelName {
			return true
		}
	}
	return false
}

func (m *model) addMessageToChannel(channelName, message string) {
	if channel, exists := m.channels[channelName]; exists {
		channel.messages = append(channel.messages, message)
	}
}

func (m *model) switchToChannel(channelName string) {
	log.Printf("Attempting to switch to channel: %s", channelName)
	if channel, exists := m.channels[channelName]; exists {
		log.Printf("Channel %s exists, joined: %v", channelName, channel.joined)
		if channel.joined {
			
			if m.currentChannel != "" {
				m.setChannelActive(m.currentChannel, false)
			}

			
			m.currentChannel = channelName
			m.setChannelActive(channelName, true)

			
			m.messages = channel.messages
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			log.Printf("Successfully switched to channel: %s", channelName)
		} else {
			log.Printf("Channel %s not joined yet", channelName)
		}
	} else {
		log.Printf("Channel %s does not exist in channels map", channelName)
	}
}

func (m *model) getChannelsList() []string {
	var result []string
	for _, channelName := range m.channelOrder {
		if channel, exists := m.channels[channelName]; exists && channel.joined {
			status := ""
			if channel.active {
				status = " [ACTIVE]"
			}
			result = append(result, channelName+status)
		}
	}
	return result
}

func (m *model) getJoinedChannels() []string {
	var joined []string
	for _, channelName := range m.channelOrder {
		if channel, exists := m.channels[channelName]; exists && channel.joined {
			joined = append(joined, channelName)
		}
	}
	return joined
}

func (m *model) nextChannel() {
	joinedChannels := m.getJoinedChannels()
	if len(joinedChannels) <= 1 {
		return
	}

	currentIndex := -1
	for i, ch := range joinedChannels {
		if ch == m.currentChannel {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 && len(joinedChannels) > 0 {
		
		m.switchToChannel(joinedChannels[0])
		return
	}

	nextIndex := (currentIndex + 1) % len(joinedChannels)
	m.switchToChannel(joinedChannels[nextIndex])
}

func (m *model) prevChannel() {
	joinedChannels := m.getJoinedChannels()
	if len(joinedChannels) <= 1 {
		return
	}

	currentIndex := -1
	for i, ch := range joinedChannels {
		if ch == m.currentChannel {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 && len(joinedChannels) > 0 {
		
		m.switchToChannel(joinedChannels[0])
		return
	}

	prevIndex := (currentIndex - 1 + len(joinedChannels)) % len(joinedChannels)
	m.switchToChannel(joinedChannels[prevIndex])
}

func main() {
	logFile, err := tea.LogToFile("irc_debug.log", "debug")
	if err != nil {
		fmt.Println("could not create log file:", err)
	}
	defer logFile.Close()

	log.Println("Starting Bubble Tea IRC Client...")

	prog := tea.NewProgram(initialModel(), tea.WithAltScreen())
	p = prog

	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
	log.Println("Bubble Tea IRC Client stopped.")
}
