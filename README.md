# Janketty

A 7/10 TUI slideshow application built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Glamour](https://github.com/charmbracelet/glamour).

## Features

- 🎨 Just-okay markdown rendering with syntax highlighting
- ⌨️  Normal-ass keyboard navigation
- 📱 Regular-ass TUI
- 🚀 Fast and lightweight

## Usage

### Installation

```bash
go build -o janketty main.go
```

### Running

```bash
./janketty
```

The application will automatically load all `.md` files from the current directory in alphabetical order.

### Controls

- `→` or `l` - Next slide
- `←` or `h` - Previous slide
- `e` - edit with vim
- `q` or `Ctrl+C` - Quit

