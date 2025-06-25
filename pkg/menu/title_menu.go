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

/* Title Menu KeyMap. */
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

/* Keybinds for the Category Menu. */
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
	selected   []types.Title
	confirmed  bool
	errMsg     string
}

/* Initializes the title model. */
func initialTitleModel(titles []types.Title, tabs []types.Category) titleModel {
	return titleModel{
		Tabs:       tabs,
		TabContent: []types.Title{},
		keys:       titleKeys,
		help:       help.New(),
		inputStyle: inputStyle,
		choices:    titles,
		selected:   []types.Title{},
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
			m.toggleTitle(choice)
		case key.Matches(msg, m.keys.ToggleAll):
			m.toggleAllTitles()
		case key.Matches(msg, m.keys.Confirm):
			/* Must select at least one title/episode. */
			if len(m.selected) != 0 {
				m.confirmed = true
				return m, m.resetWindowTitleAndQuit()
			} else {
				m.errMsg = "You must select at least one title/episode."
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

/* Title Model Style. */
var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	singleTabBorder   = tabBorderWithBottom("│", " ", "│")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1).Align(lipgloss.Center)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	singleTabStyle    = inactiveTabStyle.Border(singleTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

/* Renders the UI based on the data in the title model, `m`. */
func (m titleModel) View() string {
	doc := strings.Builder{}
	var renderedTabs []string

	longestContent := getLongestTitleString(m.choices)
	content, spaces := m.getTitleMenuContent(longestContent)
	menuHeight := strings.Count(content, "\n")

	for i, t := range m.Tabs {
		var style lipgloss.Style

		contentWidth := lipgloss.Width(windowStyle.Render(longestContent)) - windowStyle.GetHorizontalPadding() + spaces
		tabCount := len(m.Tabs)
		tabWidth := contentWidth / tabCount

		isFirst, isLast, isActive := i == 0, i == tabCount-1, i == int(m.activeTab)

		if tabCount == 1 {
			style = singleTabStyle
		} else if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		if tabCount > 1 {
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
		}

		style = style.Width(tabWidth)

		renderedTabs = append(renderedTabs, style.Render(t.String()))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	menuWidth := lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize()

	belowMenuContent := ""
	if m.errMsg != "" {
		belowMenuContent += "\n" + errMsgStyle.Render(m.errMsg) + "\n"
	}
	belowMenuContent += "\n" + m.help.View(m.keys) + "\n"

	doc.WriteString(windowStyle.Width(menuWidth).Height(menuHeight).Render(content))
	doc.WriteString(belowMenuContent)
	return docStyle.Render(doc.String())
}

/*
Returns the content to render in the Title menu,
as well as the number of extra characters that do not pertain
to the longest title/episode name (used in Tab width calculation).
*/
func (m titleModel) getTitleMenuContent(longestName string) (string, int) {
	var titleContent string
	var extraChars int

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if containsTitle(m.selected, choice) {
			checked = "x"
		}

		line := fmt.Sprintf("%s [%s] %s", cursor, checked, choice.Name)
		if longestName == choice.Name {
			extraChars = len(line) - len(longestName)
		}

		style := lipgloss.NewStyle()
		if checked == "x" {
			style = style.Inherit(selectedStyle)
		}
		if m.cursor == i {
			style = style.Inherit(highlightStyle)
		}

		titleContent += style.Render(line) + "\n"
	}

	return titleContent, extraChars
}

/*
Returns the longest title or episode name from titles `titles`.
(Useful for determining the maximum width with which to render the Title Menu.)
*/
func getLongestTitleString(titles []types.Title) string {
	maxWidth := ""

	for _, title := range titles {
		if len(title.Name) > len(maxWidth) {
			maxWidth = title.Name
		}
		for _, episode := range title.Episodes {
			if len(episode.Name) > len(maxWidth) {
				maxWidth = episode.Name
			}
		}
	}

	return maxWidth
}

/* Returns true, if `t` is in `titles`, and returns false otherwise. */
func containsTitle(titles []types.Title, t types.Title) bool {
	for _, title := range titles {
		if title.Name == t.Name {
			return true
		}
	}

	return false
}

/*
Launch the Title/Episode Menu.
Returns selected titles/episodes, or exits if the user quits.
If `debug` is enabled, then this function prints out the selected titles/episodes.
*/
func LaunchTitleMenu(titles []types.Title, tabs []types.Category, debug bool) []types.Title {
	p := tea.NewProgram(initialTitleModel(titles, tabs))
	if m, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Title Menu has encountered an error: %v", err)
		os.Exit(1)
	} else {
		m, ok := m.(titleModel)
		if ok {
			/* Debug: Print selected titles and episodes. */
			if debug {
				fmt.Println("\nSELECTED TITLES AND EPISODES:")
				for _, title := range m.selected {
					fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Link)
					for _, episode := range title.Episodes {
						fmt.Printf("\t%s -> %s\n", episode.Name, episode.Link)
					}
				}
			}
			/* User has not confirmed their selection. Exit. */
			if !m.confirmed {
				fmt.Fprintf(os.Stderr, "Title Menu: Operation aborted.\n")
				os.Exit(1)
			}

			return m.selected
		}
	}

	return []types.Title{}
}

/*
Adds/removes the title `title` from the selection of title model `m`.
Uses the URL of the title to check equality.
*/
func (m *titleModel) toggleTitle(title types.Title) {
	for i, t := range m.selected {
		if t.Link == title.Link {
			m.selected = append(m.selected[:i], m.selected[i+1:]...)
			return
		}
	}
	m.selected = append(m.selected, title)
}

/* Adds/removes all titles of title model `m`. */
// TODO: Toggle all titles that match current activeTab (i.e., match the category tab)
func (m *titleModel) toggleAllTitles() {
	if len(m.selected) < len(m.choices) {
		m.selected = make([]types.Title, len(m.choices))
		copy(m.selected, m.choices)
	} else {
		m.selected = nil
	}
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

/* Returns a command that resets the window title and quits the menu. */
func (m titleModel) resetWindowTitleAndQuit() tea.Cmd {
	return tea.Sequence(
		tea.SetWindowTitle(""),
		tea.Quit,
	)
}
