package menu

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Title Model Style. */
var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

type titleKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Toggle    key.Binding
	ToggleAll key.Binding
	Confirm   key.Binding
	Help      key.Binding
	Quit      key.Binding
}

/* Keybindings to be shown in the mini-help view. */
func (k titleKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

/* Keybindings to be shown in the expanded-help view. */
func (k titleKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Toggle, k.ToggleAll, k.Confirm}, // First Column
		{k.Help, k.Quit}, // Second Column
	}
}

var titleKeys = titleKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "w", "k"),
		key.WithHelp("↑/w/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "s", "j"),
		key.WithHelp("↓/s/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "a", "shift+tab"),
		key.WithHelp("←/a/shift+tab", "move to left tab"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "d", "tab"),
		key.WithHelp("→/d/tab", "move to right tab"),
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

/* Title/Episode Menu Model. */
type titleModel struct {
	Tabs       []types.Category
	TabContent []types.Title
	activeTab  types.Category
	keys       titleKeyMap
	help       help.Model
	inputStyle lipgloss.Style
	cursor     int
	choices    []types.Title
	selected   map[*types.Title]struct{}
	confirmed  bool
	errMsg     string
}

/* Initializes the title model. */
func initialTitleModel() titleModel {
	return titleModel{
		Tabs: []types.Category{
			types.CategoryMovie,
			types.CategoryTV,
			types.CategoryAnime,
		},
		TabContent: []types.Title{},
		keys:       titleKeys,
		help:       help.New(),
		inputStyle: inputStyle,
		choices:    []types.Title{},
		selected:   make(map[*types.Title]struct{}),
	}
}

/*
Returns an initial command for the application to run.
In this case, sets a suitable window title for the title model.
*/
func (m titleModel) Init() tea.Cmd {
	return tea.SetWindowTitle("Fancaps-Scraper-Go Title/Episode Picker")
}

/* Handles incoming events and updates the title model `m` accordingly. */
func (m titleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, m.keys.Left):
			m.setTabWrapLeft()
		case key.Matches(msg, m.keys.Right):
			m.setTabWrapRight()
		case key.Matches(msg, m.keys.Toggle):
			choice := m.choices[m.cursor]
			_, ok := m.selected[&choice]
			if ok {
				delete(m.selected, &choice)
			} else {
				m.selected[&choice] = struct{}{}
			}
		case key.Matches(msg, m.keys.ToggleAll):
			if len(m.selected) < len(m.choices) {
				for _, choice := range m.choices {
					_, ok := m.selected[&choice]
					if !ok {
						m.selected[&choice] = struct{}{}
					}
				}
			} else {
				for _, choice := range m.choices {
					_, ok := m.selected[&choice]
					if ok {
						delete(m.selected, &choice)
					}
				}
			}
		case key.Matches(msg, m.keys.Confirm):
			/* Must select at least one title/episode. */
			if len(m.selected) != 0 {
				m.confirmed = true
				return m, tea.Quit
			} else {
				m.errMsg = "You must select at least one title/episode."
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

/* Renders the UI based on the data in the title model, `m`. */
func (m titleModel) View() string {
	doc := strings.Builder{}
	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == int(m.activeTab)

		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}

		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t.String()))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab].Name))
	return docStyle.Render(doc.String())
}

/*
Launch the Title/Episode Menu.
Returns selected titles/episodes and whether the user confirmed their choice.
*/
func GetTitleMenu() (map[*types.Title]struct{}, bool) {
	p := tea.NewProgram(initialTitleModel())
	if m, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Title Menu has encountered an error: %v", err)
		os.Exit(1)
	} else {
		m, ok := m.(titleModel)
		if ok {
			return m.selected, m.confirmed
		}
	}

	return map[*types.Title]struct{}{}, false
}

/*
Set the tab of model `m` to either move left,
or wrap-around to the end of the list of tabs.
*/
func (m *titleModel) setTabWrapLeft() {
	if m.activeTab <= 0 {
		m.activeTab = types.Category(len(m.Tabs) - 1)
	} else {
		m.activeTab--
	}
}

/*
Set the tab of model `m` to either move right,
or wrap-around to the beginning of the list of tabs.
*/
func (m *titleModel) setTabWrapRight() {
	if m.activeTab >= types.Category(len(m.Tabs)-1) {
		m.activeTab = 0
	} else {
		m.activeTab++
	}
}

/*
Set the cursor of model `m` to either move up,
or wrap-around to the end of the list of choices.
*/
func (m *titleModel) setCursorWrapUp() {
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
func (m *titleModel) setCursorWrapDown() {
	if m.cursor >= len(m.choices)-1 {
		m.cursor = 0
	} else {
		m.cursor++
	}
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}
