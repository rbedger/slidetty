package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type commandBlock struct {
	commands []string
}

func parseCommandBlocks(content string) []string {
	re := regexp.MustCompile("(?s)```commands\\s*\\n(.*?)\\n```")
	matches := re.FindAllStringSubmatch(content, -1)
	var commands []string

	for _, match := range matches {
		if len(match) > 1 {
			commandText := strings.TrimSpace(match[1])
			lines := strings.Split(commandText, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					commands = append(commands, line)
				}
			}
		}
	}

	return commands
}

func stripCommandBlocks(content string) string {
	re := regexp.MustCompile("(?s)```commands\\s*\\n.*?\\n```")
	return re.ReplaceAllString(content, "")
}

func renderCommandHotkeys(commands []string, width int) []string {
	if len(commands) == 0 {
		return []string{}
	}

	keyLabels := []string{"d", "f", "g", "t", "y", "u", "i", "o", "p", "z"}

	var hotkeyLines []string
	for i, cmd := range commands {
		if i >= len(keyLabels) { // Only show first 10 commands
			break
		}
		// Truncate long commands to fit width
		displayCmd := safeTruncate(cmd, width-10)

		// Style the key with darker background (Dracula Green)
		keyStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#50fa7b")).
			Foreground(lipgloss.Color("#282a36")).
			Bold(true).
			Padding(0, 1).
			Render(keyLabels[i])

		hotkey := fmt.Sprintf("%s %s", keyStyle, displayCmd)

		// Style each hotkey line (Dracula Selection color background)
		hotkeyLine := lipgloss.NewStyle().
			Background(lipgloss.Color("#44475a")).
			Foreground(lipgloss.Color("#f8f8f2")).
			Width(width).
			Padding(0, 1).
			Render(hotkey)

		hotkeyLines = append(hotkeyLines, hotkeyLine)
	}

	return hotkeyLines
}

func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
