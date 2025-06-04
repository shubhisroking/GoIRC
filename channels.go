package main

import (
	"log"
	"strings"
)

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
			// Deactivate current channel
			if m.currentChannel != "" {
				m.setChannelActive(m.currentChannel, false)
			}

			// Switch to new channel
			m.currentChannel = channelName
			m.setChannelActive(channelName, true)

			// Update viewport with channel messages
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
		// No current channel, switch to first
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
		// No current channel, switch to first
		m.switchToChannel(joinedChannels[0])
		return
	}

	prevIndex := (currentIndex - 1 + len(joinedChannels)) % len(joinedChannels)
	m.switchToChannel(joinedChannels[prevIndex])
}
