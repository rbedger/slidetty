package main

import (
	"os"
	"strings"
)

var defaultFlashyTheme = []byte(`{
  "document": {
    "block_prefix": "\n",
    "block_suffix": "\n",
    "color": "#d4d4d4",
    "margin": 2
  },
  "block_quote": {
    "indent": 2,
    "indent_token": "┃ ",
    "color": "#6a9955",
    "italic": true
  },
  "paragraph": {},
  "list": {
    "level_indent": 2
  },
  "heading": {
    "color": "#569cd6",
    "bold": true
  },
  "h1": {
    "prefix": "⚡ ",
    "suffix": " ⚡",
    "color": "#dcdcaa",
    "bold": true
  },
  "h2": {
    "prefix": "❯ ",
    "color": "#4fc1ff"
  },
  "h3": {
    "prefix": "  ◈ ",
    "color": "#c586c0"
  },
  "text": {},
  "strikethrough": {
    "crossed_out": true
  },
  "emph": {
    "color": "#b5cea8",
    "italic": true
  },
  "strong": {
    "color": "#ce9178",
    "bold": true
  },
  "hr": {
    "color": "#3c3c3c",
    "format": "══════════════════════════════════════════════════════"
  },
  "item": {
    "block_prefix": "• "
  },
  "enumeration": {
    "block_prefix": ". "
  },
  "link": {
    "color": "#9cdcfe",
    "underline": true
  },
  "link_text": {
    "color": "#569cd6",
    "bold": true
  },
  "code": {
    "color": "#9cdcfe",
    "background_color": "#1e1e1e"
  },
  "code_block": {
    "color": "#d4d4d4",
    "background_color": "#1e1e1e",
    "margin": 2
  },
  "table": {
    "color": "#d4d4d4"
  },
  "table_header": {
    "color": "#569cd6",
    "bold": true
  },
  "table_border": {
    "color": "#3c3c3c"
  }
}`)

// loadTheme searches for a theme file in standard locations and returns its path or "auto".
func loadTheme() string {
	// Try current directory first, then slides directory
	paths := []string{"_theme.md", "slides/_theme.md"}
	for _, path := range paths {
		if themeContent, err := os.ReadFile(path); err == nil {
			theme := strings.TrimSpace(string(themeContent))
			if theme != "" {
				return theme
			}
		}
	}
	return "auto" // fallback to auto if no theme file or empty
}
