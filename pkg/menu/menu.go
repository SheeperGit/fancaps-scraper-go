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
	cursor    int
	choices   []string
	selected  map[string]struct{}
	confirmed bool
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
		/* Cancel. Return nothing. */
		case "ctrl+c", "q", "esc":
			m.confirmed = false
			return m, tea.Quit
		/* Move cursor up. */
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		/* Move cursor down. */
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		/* Toggle cursor selection. */
		case "enter", " ":
			choice := m.choices[m.cursor]
			_, ok := m.selected[choice]
			if ok {
				delete(m.selected, choice)
			} else {
				m.selected[choice] = struct{}{}
			}
		/* Confirm selection. Return model. */
		case "p":
			/* Must select at least one category. */
			// TODO: Message that selection must be non-empty.
			if len(m.selected) != 0 {
				m.confirmed = true
				return m, tea.Quit
			} else {
				fmt.Printf("Select at least one category.")
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

func GetCategoryMenu() (map[string]struct{}, bool) {
	p := tea.NewProgram(initialModel())
	if m, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Category Menu has encountered an error: %v", err)
		os.Exit(1)
	} else {
		m, ok := m.(model)
		if ok {
			return m.selected, m.confirmed
		}
	}

	return map[string]struct{}{}, false
}
