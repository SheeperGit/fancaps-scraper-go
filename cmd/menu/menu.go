package menu

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	MOVIE_TEXT = "Movies"
	TV_TEXT    = "TV Series"
	ANIME_TEXT = "Anime"
)

type model struct {
	cursor   int
	choices  []string
	selected map[string]struct{}
}

func initialModel() model {
	return model{
		choices:  []string{MOVIE_TEXT, TV_TEXT, ANIME_TEXT},
		selected: make(map[string]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Fancaps-Scraper-Go Category Picker")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			choice := m.choices[m.cursor]
			_, ok := m.selected[choice]
			if ok {
				delete(m.selected, choice)
			} else {
				m.selected[choice] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Select Categories to Scrape from:\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[choice]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress 'q' to quit.\n"

	return s
}

func GetCategoryMenu() map[string]struct{} {
	p := tea.NewProgram(initialModel())
	if m, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Category Menu has encountered an error: %v", err)
		os.Exit(1)
	} else {
		m, ok := m.(model)
		if ok {
			return m.selected
		}
	}

	return map[string]struct{}{}
}
