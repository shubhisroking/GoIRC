package main

import (
	"fmt"
	"time"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatTimestamp() string {
	return timestampStyle.Render(time.Now().Format("15:04"))
}

func formatMessage(timestamp, content string) string {
	return fmt.Sprintf("%s %s", timestamp, content)
}

func formatUserMessage(user, message string) string {
	timestamp := formatTimestamp()

	return formatMessage(timestamp, userMessageStyle.Render(fmt.Sprintf("<%s> %s", user, message)))
}

func formatUserMessageWithContext(user, message, currentNick string) string {
	timestamp := formatTimestamp()
	if user == currentNick {
		return formatMessage(timestamp, ownMessageStyle.Render(fmt.Sprintf("<%s> %s", user, message)))
	}
	return formatMessage(timestamp, userMessageStyle.Render(fmt.Sprintf("<%s> %s", user, message)))
}

func formatSystemMessage(message string) string {
	timestamp := formatTimestamp()
	return formatMessage(timestamp, systemMessageStyle.Render(message))
}

func formatJoinMessage(user, channel string) string {
	timestamp := formatTimestamp()
	return formatMessage(timestamp, joinMessageStyle.Render(fmt.Sprintf("→ %s joined %s", user, channel)))
}

func formatPartMessage(user, channel, reason string) string {
	timestamp := formatTimestamp()
	if reason != "" {
		return formatMessage(timestamp, partMessageStyle.Render(fmt.Sprintf("← %s left %s (%s)", user, channel, reason)))
	}
	return formatMessage(timestamp, partMessageStyle.Render(fmt.Sprintf("← %s left %s", user, channel)))
}

func formatQuitMessage(user, reason string) string {
	timestamp := formatTimestamp()
	if reason != "" {
		return formatMessage(timestamp, quitMessageStyle.Render(fmt.Sprintf("⇐ %s quit (%s)", user, reason)))
	}
	return formatMessage(timestamp, quitMessageStyle.Render(fmt.Sprintf("⇐ %s quit", user)))
}

func formatNoticeMessage(from, message string) string {
	timestamp := formatTimestamp()
	return formatMessage(timestamp, noticeMessageStyle.Render(fmt.Sprintf("[%s] %s", from, message)))
}

func formatErrorMessage(message string) string {
	timestamp := formatTimestamp()
	return formatMessage(timestamp, errorMessageStyle.Render(fmt.Sprintf("⚠ %s", message)))
}
