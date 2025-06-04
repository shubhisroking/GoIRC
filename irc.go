package main

import (
	"crypto/tls"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	irc "github.com/fluffle/goirc/client"
)

func (m *model) connectToIRC() tea.Cmd {
	return func() tea.Msg {
		cfg := irc.NewConfig(m.config.IRC.Nick)
		cfg.SSL = m.config.IRC.UseSSL

		if m.config.IRC.UseSSL {
			serverHost := strings.Split(m.config.IRC.Server, ":")[0]
			cfg.SSLConfig = &tls.Config{ServerName: serverHost}
		}

		// Handle port configuration
		server := m.config.IRC.Server
		if m.config.IRC.Port != 0 && !strings.Contains(server, ":") {
			server = fmt.Sprintf("%s:%d", server, m.config.IRC.Port)
		}
		cfg.Server = server

		// Set additional IRC config fields
		if m.config.IRC.Username != "" {
			cfg.Me.Ident = m.config.IRC.Username
		}
		if m.config.IRC.RealName != "" {
			cfg.Me.Name = m.config.IRC.RealName
		}
		if m.config.IRC.Password != "" {
			cfg.Pass = m.config.IRC.Password
		}
		if m.config.IRC.QuitMsg != "" {
			cfg.QuitMessage = m.config.IRC.QuitMsg
		}

		cfg.NewNick = func(n string) string { return n + "_" }

		c := irc.Client(cfg)

		c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
			m.logger.LogIRCEvent("Connected to IRC server %s", m.config.IRC.Server)
			m.logger.Debug("Our actual nickname is: %s", conn.Me().Nick)

			// Join channels after connection
			for _, channel := range m.config.IRC.Channels {
				if channel != "" {
					m.addChannel(channel)
					conn.Join(channel)
					m.logger.LogIRCEvent("Joining channel %s", channel)
				}
			}

			if p != nil {
				p.Send(ircConnectedMsg{})
			}
		})

		c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
			m.logger.LogIRCEvent("Disconnected from IRC server")
			if p != nil {
				p.Send(ircDisconnectedMsg{})
			}
		})

		c.HandleFunc(irc.NICK, func(conn *irc.Conn, line *irc.Line) {
			oldNick := line.Nick
			newNick := line.Args[0]
			m.logger.LogIRCEvent("%s changed nick to %s", oldNick, newNick)
			if p != nil {
				p.Send(ircNickChangeMsg{oldNick: oldNick, newNick: newNick})
			}
		})

		c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
			user := line.Nick
			channel := line.Args[0]
			message := line.Args[1]
			m.logger.LogIRCMessage(channel, user, message)
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
			m.logger.LogIRCEvent("Notice from %s: %s", user, message)
			if p != nil {
				p.Send(ircMessageMsg(formatNoticeMessage(user, message)))
			}
		})

		c.HandleFunc(irc.JOIN, func(conn *irc.Conn, line *irc.Line) {
			user := line.Nick
			channel := line.Args[0]
			m.logger.LogIRCEvent("%s joined %s", user, channel)

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
			m.logger.LogIRCEvent("%s left %s (%s)", user, channel, message)
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
			m.logger.LogIRCEvent("%s quit (%s)", user, message)
			if p != nil {
				p.Send(ircMessageMsg(formatQuitMessage(user, message)))
			}
		})

		if err := c.Connect(); err != nil {
			m.logger.LogError("Error connecting to IRC: %v", err)
			return ircErrorMsg{err}
		}

		return ircClientReadyMsg{client: c}
	}
}
