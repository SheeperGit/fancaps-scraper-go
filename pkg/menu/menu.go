package menu

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	ToggleAll key.Binding
	Confirm   key.Binding
	Help      key.Binding
	Quit      key.Binding
}

/* Keybindings to be shown in the mini-help view. */
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

/* Keybindings to be shown in the expanded-help view. */
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle, k.ToggleAll, k.Confirm}, // First Column
		{k.Help, k.Quit}, // Second Column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "w", "k"),
		key.WithHelp("↑/w/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "s", "j"),
		key.WithHelp("↓/s/j", "move down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/[space]", "toggle selection"),
	),
	ToggleAll: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle all"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "confirm"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

/* Enum for Categories. */
type Category int

const (
	CategoryMovie Category = iota
	CategoryTV
	CategoryAnime
	CategoryUnknown
)

var CategoryName = map[Category]string{
	CategoryMovie:   "Movies",
	CategoryTV:      "TV Series",
	CategoryAnime:   "Anime",
	CategoryUnknown: "Category Unknown",
}

/* Convert a category enumeration to its corresponding string representation. */
func (cat Category) String() string {
	return CategoryName[cat]
}

/* Menu Model. */
type model struct {
	keys       keyMap
	help       help.Model
	inputStyle lipgloss.Style
	cursor     int
	choices    []Category
	selected   map[Category]struct{}
	confirmed  bool
}

/* Initializes the model. */
func initialModel() model {
	return model{
		keys:       keys,
		help:       help.New(),
		inputStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7")),
		choices:    []Category{CategoryMovie, CategoryTV, CategoryAnime},
		selected:   make(map[Category]struct{}),
	}
}

/*
Returns an initial command for the application to run.
In this case, sets a suitable window title.
*/
func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Fancaps-Scraper-Go Category Picker")
}

/* Handles incoming events and updates the model `model` accordingly. */
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		/* Truncate help menu width based on message width. */
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Toggle):
			choice := m.choices[m.cursor]
			_, ok := m.selected[choice]
			if ok {
				delete(m.selected, choice)
			} else {
				m.selected[choice] = struct{}{}
			}
		case key.Matches(msg, m.keys.ToggleAll):
			if len(m.selected) < len(m.choices) {
				for _, choice := range m.choices {
					_, ok := m.selected[choice]
					if !ok {
						m.selected[choice] = struct{}{}
					}
				}
			} else {
				for _, choice := range m.choices {
					_, ok := m.selected[choice]
					if ok {
						delete(m.selected, choice)
					}
				}
			}
		case key.Matches(msg, m.keys.Confirm):
			/* Must select at least one category. */
			// TODO: Message that selection must be non-empty.
			if len(m.selected) != 0 {
				m.confirmed = true
				return m, tea.Quit
			} else {
				fmt.Printf("Select at least one category.")
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			m.confirmed = false
			return m, tea.Quit
		}

	}

	return m, nil
}

/* Menu Styles. */
var highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

/* Renders the UI based on the data in the model, `model`. */
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

		line := fmt.Sprintf("%s [%s] %s", cursor, checked, choice)
		if m.cursor == i {
			line = highlightStyle.Render(line)
		}
		s += line + "\n"
	}

	s += "\n" + m.help.View(m.keys) + "\n"

	return s
}

/*
Launch the Category Menu.
Returns selected categories and whether the user confirmed their choice.
*/
func GetCategoriesMenu() (map[Category]struct{}, bool) {
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

	return map[Category]struct{}{}, false
}
