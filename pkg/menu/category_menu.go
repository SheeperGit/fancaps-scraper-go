package menu

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Category Menu KeyMap. */
type catKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	ToggleAll key.Binding
	Confirm   key.Binding
	Help      key.Binding
	Quit      key.Binding
}

/* Keybindings to be shown in the mini-help view. */
func (k catKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

/* Keybindings to be shown in the expanded-help view. */
func (k catKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle, k.ToggleAll, k.Confirm}, // First Column
		{k.Help, k.Quit}, // Second Column
	}
}

/* Keybinds for the Category Menu. */
var categoryKeys = catKeyMap{
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

/* Category Menu Model. */
type categoryModel struct {
	keys       catKeyMap
	help       help.Model
	inputStyle lipgloss.Style
	cursor     int
	choices    []types.Category
	selected   map[types.Category]struct{}
	confirmed  bool
	errMsg     string
}

/* Initializes the category model. */
func initialCategoryModel() categoryModel {
	return categoryModel{
		keys:       categoryKeys,
		help:       help.New(),
		inputStyle: inputStyle,
		choices: []types.Category{
			types.CategoryMovie, types.CategoryTV, types.CategoryAnime,
		},
		selected: make(map[types.Category]struct{}),
	}
}

/*
Returns an initial command for the application to run.
In this case, sets a suitable window title for the category model.
*/
func (m categoryModel) Init() tea.Cmd {
	return tea.SetWindowTitle("Fancaps-Scraper-Go Category Picker")
}

/* Handles incoming events and updates the model `m` accordingly. */
func (m categoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		/* Truncate help menu width based on message width. */
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.setCursorWrapUp()
		case key.Matches(msg, m.keys.Down):
			m.setCursorWrapDown()
		case key.Matches(msg, m.keys.Toggle):
			m.toggle()
		case key.Matches(msg, m.keys.ToggleAll):
			m.toggleAll()
		case key.Matches(msg, m.keys.Confirm):
			/* Must select at least one category. */
			if len(m.selected) != 0 {
				m.confirmed = true
				return m, m.resetWindowTitleAndQuit()
			} else {
				m.errMsg = "You must select at least one category."
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			m.confirmed = false
			return m, m.resetWindowTitleAndQuit()
		}

	}

	return m, nil
}

/* Renders the UI based on the data in the model, `m`. */
func (m categoryModel) View() string {
	s := "Select Categories to Scrape from:\n\n"

	for i, choice := range m.choices {
		cursor := ' '
		if m.cursor == i {
			cursor = '>'
		}

		checked := ' '
		if _, ok := m.selected[choice]; ok {
			checked = 'x'
		}

		line := fmt.Sprintf(menuLineFormat, cursor, checked, choice)

		style := lipgloss.NewStyle()
		if checked == 'x' {
			style = style.Inherit(selectedStyle)
		}
		if m.cursor == i {
			style = style.Inherit(highlightStyle)
		}

		s += style.Render(line) + "\n"
	}

	if m.errMsg != "" {
		s += "\n" + errMsgStyle.Render(m.errMsg) + "\n"
	}

	s += "\n" + m.help.View(m.keys) + "\n"

	return s
}

/*
Launches the Category Menu.
Returns selected categories, or exits if the user quits.
*/
func LaunchCategoriesMenu() map[types.Category]struct{} {
	p := tea.NewProgram(initialCategoryModel())
	if m, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Category Menu has encountered an error: %v", err)
		os.Exit(1)
	} else {
		m, ok := m.(categoryModel)
		if ok {
			/* User has not confirmed their selection. Exit. */
			if !m.confirmed {
				fmt.Fprintf(os.Stderr, "Category Menu: Operation aborted.\n")
				os.Exit(1)
			}
			return m.selected
		}
	}

	return map[types.Category]struct{}{}
}

/*
Set the cursor of model `m` to either move up,
or wrap-around to the end of the list of choices.
*/
func (m *categoryModel) setCursorWrapUp() {
	if m.cursor <= 0 {
		m.cursor = len(m.choices) - 1
	} else {
		m.cursor--
	}
}

/*
Set the cursor of model `m` to either move down,
or wrap-around to the beginning of the list of choices.
*/
func (m *categoryModel) setCursorWrapDown() {
	if m.cursor >= len(m.choices)-1 {
		m.cursor = 0
	} else {
		m.cursor++
	}
}

func (m *categoryModel) toggle() {
	choice := m.choices[m.cursor]
	_, ok := m.selected[choice]
	if ok {
		delete(m.selected, choice)
	} else {
		m.selected[choice] = struct{}{}
	}
}

func (m *categoryModel) toggleAll() {
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
}

func (m categoryModel) resetWindowTitleAndQuit() tea.Cmd {
	return tea.Sequence(
		tea.SetWindowTitle(""),
		tea.Quit,
	)
}
