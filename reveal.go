package main

import (
	"strings"
)

type revealConfig struct {
	directiveLines []int
	items          [][]int
}

func (rc revealConfig) totalItems() int {
	return len(rc.items)
}

func adjustReveal(m *model, slideIndex, delta int) bool {
	if slideIndex < 0 || slideIndex >= len(m.revealConfigs) {
		return false
	}
	cfg := m.revealConfigs[slideIndex]
	total := cfg.totalItems()
	if m.revealProgress == nil {
		m.revealProgress = make(map[int]int)
	}
	current, ok := m.revealProgress[slideIndex]
	minVisible := 0
	if total > 0 {
		minVisible = 1
	}
	if !ok {
		current = minVisible
		if total > 0 {
			m.revealProgress[slideIndex] = current
		}
	}
	next := current + delta
	if next < minVisible {
		next = minVisible
	}
	if next > total {
		next = total
	}
	if next == current {
		return false
	}
	if total == 0 {
		return false
	}
	m.revealProgress[slideIndex] = next
	return true
}

func clampRevealProgress(current, total int) int {
	if total <= 0 {
		return 0
	}
	if current < 1 {
		return 1
	}
	if current > total {
		return total
	}
	return current
}

func applyReveal(content string, cfg revealConfig, count int) string {
	total := cfg.totalItems()
	if total == 0 && len(cfg.directiveLines) == 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	hide := make(map[int]struct{}, len(cfg.directiveLines))
	for _, idx := range cfg.directiveLines {
		hide[idx] = struct{}{}
	}
	visible := count
	if visible < 0 {
		visible = 0
	}
	if total > 0 && visible == 0 {
		visible = 1
	}
	if visible > len(cfg.items) {
		visible = len(cfg.items)
	}
	for _, item := range cfg.items[visible:] {
		for _, idx := range item {
			hide[idx] = struct{}{}
		}
	}
	filtered := make([]string, 0, len(lines)+1)
	for i, line := range lines {
		if _, hidden := hide[i]; hidden {
			continue
		}
		filtered = append(filtered, line)
	}
	if total > 0 && visible < len(cfg.items) {
		nextIndices := cfg.items[visible]
		if len(nextIndices) > 0 && nextIndices[0] < len(lines) {
			filtered = append(filtered, ellipsisLine(lines[nextIndices[0]]))
		} else {
			filtered = append(filtered, "...")
		}
	}
	return strings.Join(filtered, "\n")
}

func ellipsisLine(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]
	switch {
	case strings.HasPrefix(trimmed, "- "), strings.HasPrefix(trimmed, "* "), strings.HasPrefix(trimmed, "+ "):
		return indent + trimmed[:2] + "..."
	default:
		digits := 0
		for digits < len(trimmed) && trimmed[digits] >= '0' && trimmed[digits] <= '9' {
			digits++
		}
		if digits > 0 && digits < len(trimmed) && trimmed[digits] == '.' {
			if digits+1 < len(trimmed) && trimmed[digits+1] == ' ' {
				return indent + trimmed[:digits+2] + "..."
			}
		}
	}
	return indent + "..."
}

func analyzeReveal(content string) revealConfig {
	lines := strings.Split(content, "\n")
	var directive []int
	var items [][]int

	for i := 0; i < len(lines); {
		if strings.TrimSpace(lines[i]) != ":reveal:" {
			i++
			continue
		}
		directive = append(directive, i)
		i++
		for i < len(lines) {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed == "" {
				i++
				continue
			}
			if !isListItem(lines[i]) {
				break
			}
			itemIndices := []int{i}
			i++
			for i < len(lines) {
				line := lines[i]
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine == "" {
					itemIndices = append(itemIndices, i)
					i++
					break
				}
				if isListItem(line) {
					break
				}
				if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
					itemIndices = append(itemIndices, i)
					i++
					continue
				}
				break
			}
			items = append(items, itemIndices)
		}
	}

	return revealConfig{directiveLines: directive, items: items}
}

func isListItem(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		return true
	}
	idx := 0
	for idx < len(trimmed) && trimmed[idx] >= '0' && trimmed[idx] <= '9' {
		idx++
	}
	if idx == 0 {
		return false
	}
	if idx < len(trimmed) && trimmed[idx] == '.' {
		if idx+1 < len(trimmed) && trimmed[idx+1] == ' ' {
			return true
		}
	}
	return false
}
