package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type tickMsg struct{}

type editorFinishedMsg struct {
	err error
}

func extractH1(content string) (string, string) {
	lines := strings.Split(content, "\n")
	var h1 string
	var remainingLines []string
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !found && strings.HasPrefix(trimmed, "# ") {
			h1 = strings.TrimPrefix(trimmed, "# ")
			h1 = strings.TrimSpace(h1)
			found = true
			continue
		}
		remainingLines = append(remainingLines, line)
	}

	if !found {
		return "", content
	}
	return h1, strings.Join(remainingLines, "\n")
}

func getRelevantEmoji(text string) string {
	text = strings.ToLower(text)

	keywordMap := map[string][]string{
		"welcome":    {"👋", "✨", "🚀", "🎉"},
		"intro":      {"👋", "✨", "🚀"},
		"navigat":    {"🧭", "🗺️", "🔀", "➡️"},
		"key":        {"🔑", "⌨️"},
		"control":    {"🎮", "⚙️"},
		"markdown":   {"📝", "✍️", "📖"},
		"table":      {"📊", "📈", "📋"},
		"link":       {"🌐", "🔗"},
		"quote":      {"💬", "💭", "✍️"},
		"emph":       {"💡", "🔥", "✨"},
		"code":       {"💻", "🖥️", "🛠️", "⚙️"},
		"python":     {"🐍", "💻"},
		"javascript": {"🟨", "💻"},
		"js":         {"🟨", "💻"},
		"rust":       {"🦀", "⚙️"},
		"go":         {"🐹", "🚀"},
		"list":       {"📋", "📑", "📝"},
		"task":       {"✅", "☑️", "📋"},
		"git":        {"🐙", "🌲", "🌿"},
		"branch":     {"🌿", "🌲", "🔀"},
		"commit":     {"💾", "📌"},
		"restore":    {"⏪", "🔄"},
		"status":     {"🚦", "🔍", "📊"},
		"future":     {"🔮", "⏳", "🚀"},
		"thanks":     {"🙌", "💖", "🙏"},
		"qa":         {"❓", "💬", "🙋"},
		"q&a":        {"❓", "💬", "🙋"},
		"help":       {"🙋", "🆘", "❓"},
		"install":    {"📥", "⚙️", "🔧"},
	}

	var matches []string
	for keyword, emojis := range keywordMap {
		if strings.Contains(text, keyword) {
			matches = append(matches, emojis...)
		}
	}

	if len(matches) == 0 {
		matches = []string{"💡", "✨", "🚀", "📢", "🔮", "🔥", "⚙️"}
	}

	hash := 0
	for _, char := range text {
		hash += int(char)
	}

	return matches[hash%len(matches)]
}

func getWordWrapWidth(width, height int) int {
	if width == 0 {
		return 74
	}
	contentHeight := height - 2
	if width >= 50 && contentHeight >= 6 {
		return width - 6
	}
	return width - 2
}

func safeTruncate(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
