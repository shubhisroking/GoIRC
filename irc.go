package main

import (
	"crypto/tls"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	irc "github.com/fluffle/goirc/client"
)

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

			// Join channels after connection
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
