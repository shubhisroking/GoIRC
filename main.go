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
	ircServer    = "irc.libera.chat:6697"
	ircChannel   = "#bubbletea-test"
	ircNick      = "bubbletea-user"
	headerHeight = 4
	footerHeight = 4
	statusHeight = 1
	minWidth     = 80
	minHeight    = 24
)

var p *tea.Program

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
}

type (
	ircMessageMsg      string
	ircPrivmsgMsg      struct{ user, message string }
	ircErrorMsg        struct{ err error }
	ircConnectedMsg    struct{}
	ircDisconnectedMsg struct{}
	ircNickChangeMsg   struct{ oldNick, newNick string }
	ircJoinMsg         struct{ user, channel string }
	ircClientReadyMsg  struct{ client *irc.Conn }
)

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "💬 Type your message here... (Press Enter to send, Ctrl+C to quit)"
	ta.Focus()

	ta.Prompt = "▶ "
	ta.CharLimit = 500

	ta.SetWidth(minWidth)
	ta.SetHeight(1)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()

	vp := viewport.New(minWidth, 10)
	vp.SetContent(systemMessageStyle.Render("🔌 Connecting to IRC server..."))

	return model{
		textarea:       ta,
		messages:       []string{},
		viewport:       vp,
		ready:          false,
		connected:      false,
		currentChannel: ircChannel,
		currentNick:    ircNick,
		width:          minWidth,
		height:         minHeight,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.connectToIRC())
}

func (m *model) connectToIRC() tea.Cmd {
	return func() tea.Msg {
		cfg := irc.NewConfig(ircNick)
		cfg.SSL = true
		cfg.SSLConfig = &tls.Config{ServerName: strings.Split(ircServer, ":")[0]}
		cfg.Server = ircServer
		cfg.NewNick = func(n string) string { return n + "_" }

		c := irc.Client(cfg)

		c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
			log.Println("Connected to IRC")
			log.Printf("Our actual nickname is: %s", conn.Me().Nick)
			conn.Join(ircChannel)
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
				p.Send(ircPrivmsgMsg{user: user, message: message})
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.viewport.Width = max(msg.Width-4, minWidth-4)
		m.viewport.Height = max(msg.Height-headerHeight-footerHeight-statusHeight-2, 10)

		m.textarea.SetWidth(max(msg.Width-4, minWidth-4))

		UpdateStyleWidths(msg.Width)

		m.ready = true
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, tea.Batch(tiCmd, vpCmd)
	}

	switch msg := msg.(type) {
	case ircClientReadyMsg:

		m.ircClient = msg.client
		log.Println("IRC client initialized and stored in model")

	case ircConnectedMsg:
		m.connected = true
		m.connectionTime = time.Now()

		if m.ircClient != nil && m.ircClient.Connected() {
			actualNick := m.ircClient.Me().Nick
			m.currentNick = actualNick
			log.Printf("Connected with nickname: %s", actualNick)
		}
		m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("✅ Connected to %s and joined %s", ircServer, ircChannel)))
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
		m.messages = append(m.messages, formattedMsg)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

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

		if msg.user == m.currentNick || msg.user == ircNick || strings.HasPrefix(msg.user, ircNick) {
			m.currentChannel = msg.channel
			m.currentNick = msg.user
			log.Printf("Updated current channel to: %s and nick to: %s", msg.channel, msg.user)
		}

		m.messages = append(m.messages, formatJoinMessage(msg.user, msg.channel))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.ircClient != nil && m.ircClient.Connected() {
				m.ircClient.Quit("Bubble Tea client closing")
			}
			return m, tea.Quit
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
						m.ircClient.Join(args)
						m.currentChannel = args
						m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("🚪 Joining %s...", args)))
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
					m.messages = append(m.messages, formatSystemMessage(fmt.Sprintf("🚪 Leaving %s...", channelToPart)))
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
				case "HELP":
					helpText := []string{
						"🆘 Available Commands:",
						"  /join #channel  - Join a channel",
						"  /part [#channel] - Leave current or specified channel",
						"  /nick <newnick> - Change nickname",
						"  /msg <nick> <message> - Send private message",
						"  /quit [message] - Quit IRC",
						"  /help - Show this help",
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

				if m.currentChannel == "" {
					m.currentChannel = ircChannel
				}
				log.Printf("Sending PRIVMSG to %s: '%s' (from nick: %s)", m.currentChannel, inputValue, m.currentNick)

				m.ircClient.Privmsg(m.currentChannel, inputValue)

				m.messages = append(m.messages, formatUserMessageWithContext(m.currentNick, inputValue, m.currentNick))
				log.Printf("Message added to local display: <%s> %s", m.currentNick, inputValue)
			}

			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			m.textarea.Reset()

			return m, nil
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

	var headerText string
	if m.connected {
		uptime := time.Since(m.connectionTime).Truncate(time.Second)
		headerText = fmt.Sprintf("📡 IRC Client - %s @ %s (%s) - Connected for %v",
			m.currentNick, m.currentChannel, ircServer, uptime)
	} else {
		headerText = "📡 IRC Client - Connecting..."
	}
	header := headerStyle.Render(headerText)

	var statusText string
	if m.connected {
		statusText = fmt.Sprintf("✅ Connected | Channel: %s | Users: Online", m.currentChannel)
	} else {
		statusText = "🔄 Connecting to server..."
	}
	status := statusStyle.Render(statusText)

	chatContent := m.viewport.View()
	chat := chatAreaStyle.Render(chatContent)

	textareaView := m.textarea.View()
	input := inputBoxFocusedStyle.Render(textareaView)

	help := helpStyle.Render("Commands: /help, /join #channel, /nick <name>, /quit | Ctrl+C to exit")

	return lipgloss.JoinVertical(lipgloss.Left, header, status, chat, input, help)
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
