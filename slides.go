package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type slidesLoadedMsg struct {
	slides        []string
	paths         []string
	title         string
	author        string
	revealConfigs []revealConfig
	commandBlocks [][]string
}

type slideReloadedMsg struct {
	slideIndex   int
	content      string
	config       revealConfig
	path         string
	commandBlock []string
}

func loadSlides() tea.Msg {
	files, err := os.ReadDir(".")
	if err != nil {
		return errMsg(err)
	}

	var slides []string
	var filenames []string
	var title string
	var author string
	var configs []revealConfig
	var paths []string
	var commandBlocks [][]string

	// Load title from _title.md if it exists (check current dir first, then slides dir)
	titlePaths := []string{"_title.md", "slides/_title.md"}
	for _, path := range titlePaths {
		if titleContent, err := os.ReadFile(path); err == nil {
			title = strings.TrimSpace(string(titleContent))
			break
		}
	}

	// Load author from _author.md if it exists (check current dir first, then slides dir)
	authorPaths := []string{"_author.md", "slides/_author.md"}
	for _, path := range authorPaths {
		if authorContent, err := os.ReadFile(path); err == nil {
			author = strings.TrimSpace(string(authorContent))
			break
		}
	}

	// Collect markdown files (excluding files starting with underscore)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" && !strings.HasPrefix(file.Name(), "_") {
			filenames = append(filenames, file.Name())
		}
	}

	// Sort filenames to ensure consistent order
	sort.Strings(filenames)

	// Read file contents
	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return errMsg(err)
		}
		slide := string(content)
		slides = append(slides, slide)
		configs = append(configs, analyzeReveal(slide))
		paths = append(paths, filename)
		commandBlocks = append(commandBlocks, parseCommandBlocks(slide))
	}

	return slidesLoadedMsg{
		slides:        slides,
		title:         title,
		author:        author,
		revealConfigs: configs,
		paths:         paths,
		commandBlocks: commandBlocks,
	}
}

func reloadSlide(slideIndex int) tea.Cmd {
	return func() tea.Msg {
		files, err := os.ReadDir(".")
		if err != nil {
			return errMsg(err)
		}

		var filenames []string

		// Collect markdown files (excluding files starting with underscore)
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".md" && !strings.HasPrefix(file.Name(), "_") {
				filenames = append(filenames, file.Name())
			}
		}

		// Sort filenames to ensure consistent order
		sort.Strings(filenames)

		// Check if slideIndex is valid
		if slideIndex < 0 || slideIndex >= len(filenames) {
			return errMsg(fmt.Errorf("invalid slide index: %d", slideIndex))
		}

		// Read the specific slide content
		content, err := os.ReadFile(filenames[slideIndex])
		if err != nil {
			return errMsg(err)
		}

		slide := string(content)
		return slideReloadedMsg{
			slideIndex:   slideIndex,
			content:      slide,
			config:       analyzeReveal(slide),
			path:         filenames[slideIndex],
			commandBlock: parseCommandBlocks(slide),
		}
	}
}
