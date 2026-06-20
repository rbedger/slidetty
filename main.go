package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	slides            []string
	slidePaths        []string
	currentSlide      int
	renderer          *glamour.TermRenderer
	progress          progress.Model
	width             int
	height            int
	title             string
	author            string
	err               error
	revealConfigs     []revealConfig
	revealProgress    map[int]int
	commandBlocks     [][]string // commands for each slide
	notification      string
	notificationTimer int
}

func initialModel() model {
	// Initialize glamour renderer with theme from _theme.md
	theme := loadTheme()
	var r *glamour.TermRenderer
	wrapWidth := getWordWrapWidth(0, 0)
	if theme == "auto" {
		r, _ = glamour.NewTermRenderer(
			glamour.WithStylesFromJSONBytes(defaultFlashyTheme),
			glamour.WithWordWrap(wrapWidth),
		)
	} else {
		r, _ = glamour.NewTermRenderer(
			glamour.WithStylePath(theme),
			glamour.WithWordWrap(wrapWidth),
		)
	}

	// Initialize progress bar with VSCode blue solid fill
	prog := progress.New(progress.WithSolidFill("#007acc"))

	return model{
		slides:         []string{},
		slidePaths:     []string{},
		currentSlide:   0,
		renderer:       r,
		progress:       prog,
		title:          "",
		author:         "",
		revealConfigs:  nil,
		revealProgress: make(map[int]int),
		commandBlocks:  [][]string{},
	}
}

func (m model) Init() tea.Cmd {
	return loadSlides
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.renderer != nil {
			theme := loadTheme()
			var r *glamour.TermRenderer
			wrapWidth := getWordWrapWidth(msg.Width, msg.Height)
			if theme == "auto" {
				r, _ = glamour.NewTermRenderer(
					glamour.WithStylesFromJSONBytes(defaultFlashyTheme),
					glamour.WithWordWrap(wrapWidth),
				)
			} else {
				r, _ = glamour.NewTermRenderer(
					glamour.WithStylePath(theme),
					glamour.WithWordWrap(wrapWidth),
				)
			}
			m.renderer = r
		}
		m.progress.Width = msg.Width - 4
		return m, nil

	case slidesLoadedMsg:
		m.slides = msg.slides
		m.slidePaths = msg.paths
		m.title = msg.title
		m.author = msg.author
		m.revealConfigs = msg.revealConfigs
		m.commandBlocks = msg.commandBlocks
		m.revealProgress = make(map[int]int, len(msg.revealConfigs))
		for idx, cfg := range msg.revealConfigs {
			if cfg.totalItems() > 0 {
				m.revealProgress[idx] = 1
			}
		}
		if len(m.slides) == 0 {
			return m, nil
		}
		if m.currentSlide >= len(m.slides) {
			m.currentSlide = len(m.slides) - 1
		}
		percentage := float64(m.currentSlide+1) / float64(len(m.slides))
		m.progress.SetPercent(percentage)

		return m, nil

	case slideReloadedMsg:
		if msg.slideIndex >= 0 && msg.slideIndex < len(m.slides) {
			m.slides[msg.slideIndex] = msg.content
			if len(m.revealConfigs) != len(m.slides) {
				newConfigs := make([]revealConfig, len(m.slides))
				copy(newConfigs, m.revealConfigs)
				m.revealConfigs = newConfigs
			}
			if len(m.slidePaths) != len(m.slides) {
				newPaths := make([]string, len(m.slides))
				copy(newPaths, m.slidePaths)
				m.slidePaths = newPaths
			}
			if len(m.commandBlocks) != len(m.slides) {
				newCommandBlocks := make([][]string, len(m.slides))
				copy(newCommandBlocks, m.commandBlocks)
				m.commandBlocks = newCommandBlocks
			}
			m.revealConfigs[msg.slideIndex] = msg.config
			m.commandBlocks[msg.slideIndex] = msg.commandBlock
			if msg.path != "" {
				m.slidePaths[msg.slideIndex] = msg.path
			}
			current, ok := m.revealProgress[msg.slideIndex]
			total := msg.config.totalItems()
			minVisible := 0
			if total > 0 {
				minVisible = 1
			}
			if !ok {
				current = minVisible
			}
			if current < minVisible {
				current = minVisible
			}
			if current > total {
				current = total
			}
			if total == 0 {
				delete(m.revealProgress, msg.slideIndex)
			} else {
				m.revealProgress[msg.slideIndex] = current
			}
		}
		return m, nil

	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return m, reloadSlide(m.currentSlide)

	case errMsg:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			return m, tea.Quit

		case "e":
			if len(m.slides) == 0 || m.currentSlide < 0 || m.currentSlide >= len(m.slides) {
				return m, nil
			}
			filePath := m.slidePaths[m.currentSlide]
			if filePath == "" {
				return m, nil
			}
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}
			c := exec.Command(editor, filePath)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return editorFinishedMsg{err: err}
			})

		case "r":
			if len(m.slides) > 0 {
				return m, reloadSlide(m.currentSlide)
			}
			return m, nil

		case "d":
			// Handle first command hotkey
			if m.currentSlide < len(m.commandBlocks) && len(m.commandBlocks[m.currentSlide]) > 0 {
				commands := m.commandBlocks[m.currentSlide]
				if len(commands) > 0 {
					if err := copyToClipboard(commands[0]); err != nil {
						m.notification = fmt.Sprintf("Copy error: %v", err)
					} else {
						// Truncate command text to fit notification bar
						displayCmd := safeTruncate(commands[0], m.width-12)
						m.notification = fmt.Sprintf("Copied: %s", displayCmd)
					}
					m.notificationTimer = 3
					return m, doTick()
				}
			}
			return m, nil

		case "down", "j":
			if adjustReveal(&m, m.currentSlide, 1) {
				return m, nil
			}
			if m.currentSlide < len(m.slides)-1 {
				m.currentSlide++
				percentage := float64(m.currentSlide+1) / float64(len(m.slides))
				m.progress.SetPercent(percentage)
			}
			return m, nil

		case "up", "k":
			if adjustReveal(&m, m.currentSlide, -1) {
				return m, nil
			}
			if m.currentSlide > 0 {
				m.currentSlide--
				percentage := float64(m.currentSlide+1) / float64(len(m.slides))
				m.progress.SetPercent(percentage)
			}
			return m, nil

		case "right", "l":
			if m.currentSlide < len(m.slides)-1 {
				m.currentSlide++
				percentage := float64(m.currentSlide+1) / float64(len(m.slides))
				m.progress.SetPercent(percentage)
			}
			return m, nil

		case "left", "h":
			if m.currentSlide > 0 {
				m.currentSlide--
				percentage := float64(m.currentSlide+1) / float64(len(m.slides))
				m.progress.SetPercent(percentage)
			}
			return m, nil

		case "f", "g", "t", "y", "u", "i", "o", "p", "z":
			// Handle command hotkeys (only if current slide has commands)
			if m.currentSlide < len(m.commandBlocks) && len(m.commandBlocks[m.currentSlide]) > 0 {
				commands := m.commandBlocks[m.currentSlide]
				// Map keys to indices (d=0 is handled above, start from f=1)
				keyMap := map[string]int{
					"f": 1, "g": 2, "t": 3, "y": 4,
					"u": 5, "i": 6, "o": 7, "p": 8, "z": 9,
				}

				if cmdNum, exists := keyMap[msg.String()]; exists && cmdNum < len(commands) {
					if err := copyToClipboard(commands[cmdNum]); err != nil {
						m.notification = fmt.Sprintf("Copy error: %v", err)
					} else {
						// Truncate command text to fit notification bar
						displayCmd := safeTruncate(commands[cmdNum], m.width-12)
						m.notification = fmt.Sprintf("Copied: %s", displayCmd)
					}
					m.notificationTimer = 3 // Show for 3 seconds
					return m, doTick()
				}
			}
			return m, nil

		default:
			return m, nil
		}

	case tickMsg:
		if m.notificationTimer > 0 {
			m.notificationTimer--
			if m.notificationTimer <= 0 {
				m.notification = ""
			} else {
				return m, doTick()
			}
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", m.err)
	}

	if len(m.slides) == 0 {
		return "Loading slides...\n\nPress 'q' to quit."
	}

	// Render current slide with glamour
	slideContent := m.slides[m.currentSlide]
	if m.currentSlide < len(m.revealConfigs) {
		slideContent = applyReveal(slideContent, m.revealConfigs[m.currentSlide], m.revealProgress[m.currentSlide])
	}
	// Strip command blocks from rendered content
	slideContent = stripCommandBlocks(slideContent)

	// Calculate available height for content (reserve lines for bottom bars)
	contentHeight := m.height - 2 // status + progress
	var commandHotkeyLines []string
	if m.currentSlide < len(m.commandBlocks) && len(m.commandBlocks[m.currentSlide]) > 0 {
		commandHotkeyLines = renderCommandHotkeys(m.commandBlocks[m.currentSlide], m.width)
		contentHeight -= len(commandHotkeyLines) // reserve lines for each command hotkey
	}
	if m.notification != "" {
		contentHeight-- // additional line for notification
	}

	// Determine if we should wrap in a beautiful border panel
	useBorder := m.width >= 50 && contentHeight >= 6
	var rendered string
	var slideTitle string
	var bodyContent string

	if useBorder {
		// Extract H1 header for dashboard style header panel
		var title string
		title, bodyContent = extractH1(slideContent)
		slideTitle = title

		var err error
		rendered, err = m.renderer.Render(bodyContent)
		if err != nil {
			rendered = "Error rendering markdown: " + err.Error()
		}
	} else {
		var err error
		rendered, err = m.renderer.Render(slideContent)
		if err != nil {
			rendered = "Error rendering markdown: " + err.Error()
		}
	}

	innerHeight := contentHeight
	if useBorder {
		innerHeight -= 4 // borders (2) + header (1) + divider (1)
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	// Split rendered content into lines and fit to available inner height
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	// Pad content to fill the available inner height
	content := strings.Join(lines, "\n")
	contentLines := len(lines)
	if contentLines < innerHeight {
		padding := strings.Repeat("\n", innerHeight-contentLines)
		content += padding
	}

	// Wrap in border if applicable
	var framedContent string
	if useBorder {
		// Calculate the header text
		headerText := " ⚡ "
		if slideTitle != "" {
			emoji := getRelevantEmoji(slideTitle)
			headerText += emoji + " " + strings.ToUpper(slideTitle)
		} else {
			headerText += "JANKETTY"
		}
		headerText += " ⚡"

		pagination := fmt.Sprintf("[%02d/%02d]", m.currentSlide+1, len(m.slides))

		// Right align the pagination inside the box (width - 2 border characters)
		innerWidth := m.width - 2
		spacing := innerWidth - lipgloss.Width(headerText) - lipgloss.Width(pagination) - 2 // additional padding spaces
		if spacing < 0 {
			spacing = 0
		}
		headerLineText := " " + headerText + strings.Repeat(" ", spacing) + pagination + " "

		// Styled header row (VSCode Yellow text on dark gray panel)
		headerRow := lipgloss.NewStyle().
			Background(lipgloss.Color("#252526")).
			Foreground(lipgloss.Color("#dcdcaa")).
			Bold(true).
			Width(innerWidth).
			Render(headerLineText)

		// Styled double border horizontal separator divider (VSCode Blue)
		divider := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#007acc")).
			Render(strings.Repeat("═", innerWidth))

		// Padded slide content
		contentStyle := lipgloss.NewStyle().
			Width(innerWidth).
			Height(innerHeight).
			Padding(0, 2)
		paddedContent := contentStyle.Render(content)

		// Combine all parts
		combined := lipgloss.JoinVertical(lipgloss.Left, headerRow, divider, paddedContent)

		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#007acc")). // VSCode blue
			Width(m.width).
			Height(contentHeight) // outer height
		framedContent = borderStyle.Render(combined)
	} else {
		plainStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(contentHeight)
		framedContent = plainStyle.Render(content)
	}

	// Get the static progress bar (animations are disabled to prevent freezing)
	progressBar := m.progress.ViewAs(m.progress.Percent())

	// Create three-section status line with chevrons
	slideInfo := fmt.Sprintf("📖 Slide %d/%d", m.currentSlide+1, len(m.slides))

	titleText := m.title
	if titleText == "" {
		titleText = "Janketty"
	}
	titleText = "🏷️ " + titleText

	authorText := m.author
	if authorText == "" {
		authorText = "Unknown"
	}
	authorText = "👤 " + authorText

	// Define styles for the three sections (VSCode Dark theme colors)
	leftStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#007acc")). // VSCode Blue
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Padding(0, 1)

	centerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#252526")). // VSCode Panel dark gray
		Foreground(lipgloss.Color("#d4d4d4")). // VSCode text gray
		Bold(true).
		PaddingLeft(1).
		PaddingRight(0)

	rightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")). // VSCode Status Bar gray
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Padding(0, 1)

	// Chevron styles (colors transition seamlessly)
	leftChevronStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#252526")). // center background
		Foreground(lipgloss.Color("#007acc"))  // left background

	rightChevronStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")). // right background
		Foreground(lipgloss.Color("#252526"))  // center background

	// Calculate section widths (approximate thirds)
	totalWidth := m.width
	chevronWidth := 1
	sectionWidth := (totalWidth - 2*chevronWidth) / 3

	// Adjust for any remaining width
	leftWidth := sectionWidth
	centerWidth := sectionWidth
	rightWidth := totalWidth - leftWidth - centerWidth - 2*chevronWidth

	if leftWidth < 0 {
		leftWidth = 0
	}
	if centerWidth < 0 {
		centerWidth = 0
	}
	if rightWidth < 0 {
		rightWidth = 0
	}

	// Truncate text safely to fit the padded sections
	slideInfo = safeTruncate(slideInfo, leftWidth-2)
	authorText = safeTruncate(authorText, centerWidth-2)
	titleText = safeTruncate(titleText, rightWidth-2)

	// Create sections with proper width and alignment
	leftSection := leftStyle.Width(leftWidth).Render(slideInfo)
	centerSection := centerStyle.Width(centerWidth).Align(lipgloss.Center).Render(authorText)
	rightSection := rightStyle.Width(rightWidth).Align(lipgloss.Right).Render(titleText)

	// Create chevrons using Nerd Font Powerline symbol U+E0B0
	leftChevron := leftChevronStyle.Render("\uE0B0")
	rightChevron := rightChevronStyle.Render("\uE0B0")

	// Combine all sections
	statusLine := lipgloss.JoinHorizontal(lipgloss.Top, leftSection, leftChevron, centerSection, rightChevron, rightSection)

	// Create notification bar if there's a notification
	var notificationBar string
	if m.notification != "" {
		notificationBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#007acc")). // VSCode Blue
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Width(m.width).
			Padding(0, 1).
			Render(m.notification)
	}

	// Build final layout
	result := framedContent
	if notificationBar != "" {
		result += "\n" + notificationBar
	}
	// Add each command hotkey line
	for _, hotkeyLine := range commandHotkeyLines {
		result += "\n" + hotkeyLine
	}
	result += "\n" + statusLine + "\n" + progressBar

	return result
}

func initProject() error {
	// Check if slides directory already exists
	if _, err := os.Stat("slides"); err == nil {
		return fmt.Errorf("slides directory already exists")
	}

	// Create slides directory
	if err := os.MkdirAll("slides", 0755); err != nil {
		return fmt.Errorf("failed to create slides directory: %v", err)
	}

	// Create _title.md
	titleContent := "My Presentation"
	if err := os.WriteFile("slides/_title.md", []byte(titleContent), 0644); err != nil {
		return fmt.Errorf("failed to create _title.md: %v", err)
	}

	// Create _author.md
	authorContent := "Your Name"
	if err := os.WriteFile("slides/_author.md", []byte(authorContent), 0644); err != nil {
		return fmt.Errorf("failed to create _author.md: %v", err)
	}

	// Create example slides
	slide1 := `# Welcome to Janketty

Welcome to your new presentation!

This is your first slide. You can edit this file and add more slides.

:reveal:
- Navigate with arrow keys or h/j/k/l
- Press 'q' to quit
- Press 'e' to edit current slide
- Press 'r' to reload slides`

	slide2 := `# Features

Janketty supports many great features:

:reveal:
- **Markdown rendering** with beautiful syntax highlighting
- **Progressive reveal** for bullet points
- **Command hotkeys** for copying commands to clipboard
- **Live editing** of slides
- **Responsive design** that adapts to your terminal

Try pressing 'j' and 'k' to reveal items progressively!`

	slide3 := `# Getting Started

Here's how to work with Janketty:

:reveal:
1. **Create slides** - Add numbered markdown files (01-slide.md, 02-slide.md, etc.)
2. **Edit content** - Press 'e' to edit the current slide
3. **Add commands** - Use ` + "```commands```" + ` blocks for copyable commands
4. **Customize theme** - Create _theme.md to set your preferred style

` + "```commands" + `
echo "Hello, World!"
ls -la
git status
` + "```" + `

Press 'd' to copy the first command above!`

	// Write example slides
	slides := map[string]string{
		"01-welcome.md":         slide1,
		"02-features.md":        slide2,
		"03-getting-started.md": slide3,
	}

	for filename, content := range slides {
		if err := os.WriteFile(filepath.Join("slides", filename), []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %v", filename, err)
		}
	}

	fmt.Println("✅ Slideshow initialized successfully!")
	fmt.Println("\nCreated files:")
	fmt.Println("  slides/")
	fmt.Println("  ├── _title.md")
	fmt.Println("  ├── _author.md")
	fmt.Println("  ├── 01-welcome.md")
	fmt.Println("  ├── 02-features.md")
	fmt.Println("  └── 03-getting-started.md")
	fmt.Println("\nRun 'janketty' to start your presentation!")

	return nil
}

func main() {
	// Check for init command
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := initProject(); err != nil {
			fmt.Printf("Error initializing project: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run normal slideshow
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
