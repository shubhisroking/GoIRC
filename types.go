package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	stateCommandPalette
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

type commandPaletteItem struct {
	name        string
	description string
	command     string
	category    string
	icon        string
	shortcut    string
	priority    int // Higher priority items appear first
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

	// Channel management
	channels       map[string]*channelData
	channelOrder   []string
	activeChannels []string

	// UI state
	showSidebar  bool
	sidebarWidth int

	// Command palette
	commandPaletteVisible  bool
	commandPaletteQuery    string
	commandPaletteItems    []commandPaletteItem
	commandPaletteSelected int
	commandPaletteFiltered []commandPaletteItem

	state                appState
	setupPhase           setupPhase
	config               *Config // Updated to use the new Config struct
	setupPrompt          string
	setupValidationError string // For showing validation errors in setup
	autoJoinChannels     []string
	logger               *Logger // Add logger instance
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
