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
	"sheeper.com/fancaps-scraper-go/pkg/ui"
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

/* Title Menu Model. */
type titleModel struct {
	/* Base Title Model fields. */

	Tabs       []types.Category // Menu tabs.
	TabContent []*types.Title   // Active tab content.
	activeTab  types.Category   // Currently viewed tab.
	keys       titleKeyMap      // Menu Keybinds.
	help       help.Model       // Help view.
	inputStyle lipgloss.Style   // Input style.
	cursor     int              // Position of current selection.
	choices    []*types.Title   // Available choices.
	selected   []*types.Title   // Selected choices.
	confirmed  bool             // True if the user confirmed their selection, false otherwise.
	errMsg     string           // Error message. If empty, no errors.

	/* Cached View() fields. */

	catStats     *types.CatStats // Amount of titles per category.
	menuWidth    int             // Menu width.
	menuHeight   int             // Menu height.
	contentWidth int             // Content width.
}

/* Initializes the title model. */
func initialTitleModel(titles []*types.Title) titleModel {
	contentPadding := getContentPadding(menuLineFormat)
	contentWidth := lipgloss.Width(windowStyle.Render(ui.GetLongestTitle(titles))) + contentPadding - windowStyle.GetHorizontalPadding()

	catStats := types.GetCatStats(titles)
	tabs := catStats.UsedCategories()
	activeTab := tabs[0]
	menuWidth := lipgloss.Width(getTabRow(tabs, activeTab, contentWidth)) - windowStyle.GetHorizontalFrameSize()
	menuHeight := catStats.Max + windowStyle.GetVerticalFrameSize()

	return titleModel{
		Tabs:       tabs,
		TabContent: []*types.Title{},
		activeTab:  activeTab,
		keys:       titleKeys,
		help:       help.New(),
		inputStyle: inputStyle,
		choices:    titles,
		selected:   []*types.Title{},

		catStats:     catStats,
		menuWidth:    menuWidth,
		menuHeight:   menuHeight,
		contentWidth: contentWidth,
	}
}

/*
Returns an initial command for the application to run.
In this case, sets a suitable window title for the title model.
*/
func (m titleModel) Init() tea.Cmd {
	return tea.SetWindowTitle("Fancaps-Scraper-Go Title Picker")
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
			m.toggle(choice)
		case key.Matches(msg, m.keys.ToggleAll):
			m.toggleAll()
		case key.Matches(msg, m.keys.Confirm):
			/* Must select at least one title. */
			if len(m.selected) != 0 {
				m.confirmed = true
				return m, m.resetWindowTitleAndQuit()
			} else {
				m.errMsg = "You must select at least one title."
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

	tabRow := getTabRow(m.Tabs, m.activeTab, m.contentWidth)
	doc.WriteString(tabRow)

	content := m.getTitleMenuContent()

	belowMenuContent := ""
	if m.errMsg != "" {
		belowMenuContent += "\n" + ui.ErrStyle.Render(m.errMsg) + "\n"
	}
	belowMenuContent += "\n" + m.help.View(m.keys) + "\n"

	doc.WriteString(windowStyle.Width(m.menuWidth).Height(m.menuHeight).Render(content))
	doc.WriteString(belowMenuContent)

	return docStyle.Render(doc.String())
}

func getTabRow(tabs []types.Category, activeTab types.Category, contentWidth int) string {
	tabCount := len(tabs)

	tabWidth := contentWidth / tabCount
	if tabCount == 1 {
		tabWidth += 2 // Increase tab width by 2 to make the single-tab menu look less smushed.
	}

	var renderedTabs []string
	for i, t := range tabs {
		tabStyle := lipgloss.NewStyle().Width(tabWidth)
		isFirst, isLast, isActive := i == 0, i == tabCount-1, t == activeTab

		if tabCount == 1 {
			tabStyle = tabStyle.Inherit(singleTabStyle)
		} else if isActive {
			tabStyle = tabStyle.Inherit(activeTabStyle)
		} else {
			tabStyle = tabStyle.Inherit(inactiveTabStyle)
		}

		if tabCount > 1 {
			border, _, _, _, _ := tabStyle.GetBorder()
			if isFirst && isActive {
				border.BottomLeft = "│"
			} else if isFirst && !isActive {
				border.BottomLeft = "├"
			} else if isLast && isActive {
				border.BottomRight = "│"
			} else if isLast && !isActive {
				border.BottomRight = "┤"
			}
			tabStyle = tabStyle.Border(border)
		}
		renderedTabs = append(renderedTabs, tabStyle.Render(t.String()))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...) + "\n"
}

/* Returns the content to render in the Title menu. */
func (m titleModel) getTitleMenuContent() string {
	var titleContent string

	for i, choice := range m.choices {
		/* Only render titles matching the category of the active tab. */
		if m.activeTab == choice.Category {
			cursor := ' '
			if m.cursor == i {
				cursor = '>'
			}

			checked := ' '
			if containsTitle(m.selected, choice) {
				checked = 'x'
			}

			line := fmt.Sprintf(menuLineFormat, cursor, checked, choice.Name)

			style := lipgloss.NewStyle()
			if checked == 'x' {
				style = style.Inherit(ui.SuccessStyle)
			}
			if m.cursor == i {
				style = style.Inherit(highlightStyle)
			}

			titleContent += style.Render(line) + "\n"
		}
	}

	return titleContent
}

/*
Returns the number of characters that don't pertain to a title name
from a format string `s`.
*/
func getContentPadding(s string) int {
	lastS := strings.LastIndex(s, "%s")
	if lastS == -1 {
		return 0
	}

	count := 0
	for i := 0; i < lastS; i++ {
		/* Count `%c` as one char. */
		if i+1 < lastS && s[i:i+2] == "%c" {
			i++
			continue
		}
		count++
	}

	return count
}

/* Returns true, if `t` is in `titles`, and returns false otherwise. */
func containsTitle(titles []*types.Title, t *types.Title) bool {
	for _, title := range titles {
		if title.Link == t.Link {
			return true
		}
	}

	return false
}

/*
Launches the Title Menu.
Returns non-empty selected titles, or exits if the user quits.
If `debug` is enabled, then this function prints out the selected titles.
*/
func LaunchTitleMenu(titles []*types.Title, tabs []types.Category, debug bool) []*types.Title {
	p := tea.NewProgram(initialTitleModel(titles))
	if m, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Title Menu has encountered an error: %v", err)
		os.Exit(1)
	} else {
		m, ok := m.(titleModel)
		if ok {
			/* Debug: Print selected titles. */
			if debug {
				fmt.Println("\nSELECTED TITLES:")
				for _, title := range m.selected {
					fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Link)
				}
			}
			fmt.Println()

			/* User has not confirmed their selection. Exit. */
			if !m.confirmed {
				fmt.Fprintf(os.Stderr, "Title Menu: Operation aborted.\n")
				os.Exit(1)
			}

			return m.selected
		}
	}

	return []*types.Title{}
}

/*
Adds/removes the title `title` to/from the selection of title model `m`.
Uses the URL of the title to check equality.
*/
func (m *titleModel) toggle(title *types.Title) {
	for i, t := range m.selected {
		if t.Link == title.Link {
			m.selected = append(m.selected[:i], m.selected[i+1:]...)
			return
		}
	}
	m.selected = append(m.selected, title)
}

/* Adds/removes all titles to/from the selection of title model `m` on the active tab. */
func (m *titleModel) toggleAll() {
	/* Get all titles from the active tab. */
	var activeTabCategories []*types.Title
	for _, t := range m.choices {
		if t.Category == m.activeTab {
			activeTabCategories = append(activeTabCategories, t)
		}
	}

	/* Determine whether all titles in the active tab are selected. */
	allSelected := true
	for _, t := range activeTabCategories {
		if !containsTitle(m.selected, t) {
			allSelected = false
			break
		}
	}

	if allSelected {
		/* New selection becomes all previously selected titles excluding all titles from the active tab. */
		var newSelected []*types.Title
		for _, t := range m.selected {
			if t.Category != m.activeTab {
				newSelected = append(newSelected, t)
			}
		}
		m.selected = newSelected
	} else {
		/* If a title from the active tab is not selected, add it to the selection. */
		for _, t := range activeTabCategories {
			if !containsTitle(m.selected, t) {
				m.selected = append(m.selected, t)
			}
		}
	}
}

/*
Set the tab of model `m` to either move left,
or wrap-around to the end of the list of tabs.
*/
func (m *titleModel) setTabWrapLeft() {
	i := m.getCategoryTabIndex(m.activeTab)

	switch i {
	case 0: // Go to last tab, if at the first tab.
		m.activeTab = m.Tabs[len(m.Tabs)-1]
	case -1: // Fallback on first tab, if category to switch to was not found.
		fmt.Fprintf(os.Stderr, "title menu error: failed to find %s in tabs.\ndefaulting to first tab...\n\n", m.activeTab.String())
		m.activeTab = m.Tabs[0]
	default: // Go to previous tab, otherwise.
		m.activeTab = m.Tabs[i-1]
	}

	/* Set cursor to the beginning of the switched tab. */
	m.cursor = m.getTabStartIndex()
}

/*
Set the tab of model `m` to either move right,
or wrap-around to the beginning of the list of tabs.
*/
func (m *titleModel) setTabWrapRight() {
	i := m.getCategoryTabIndex(m.activeTab)

	switch i {
	case len(m.Tabs) - 1: // Go to first tab, if at the last tab.
		m.activeTab = m.Tabs[0]
	case -1: // Fallback on last tab, if category to switch to was not found.
		fmt.Fprintf(os.Stderr, "title menu error: failed to find %s in tabs.\ndefaulting to last tab...\n\n", m.activeTab.String())
		m.activeTab = m.Tabs[len(m.Tabs)-1]
	default: // Go to next tab, otherwise.
		m.activeTab = m.Tabs[i+1]
	}

	/* Set cursor to the beginning of the switched tab. */
	m.cursor = m.getTabStartIndex()
}

/*
Returns the index of the category `cat` relative to the tab list.
Returns -1, if `cat` is not found in the tab list.
*/
func (m *titleModel) getCategoryTabIndex(cat types.Category) int {
	for i, tab := range m.Tabs {
		if tab == cat {
			return i
		}
	}

	return -1
}

/* Returns the starting cursor index of the active tab. */
func (m *titleModel) getTabStartIndex() int {
	tabStartIndex := 0
	for _, cat := range m.Tabs {
		if cat == m.activeTab {
			break
		}
		tabStartIndex += m.catStats.Amts[cat]
	}

	return tabStartIndex
}

/*
Set the cursor of model `m` to either move up,
or wrap-around to the end of the list of choices.
*/
func (m *titleModel) setCursorWrapUp() {
	if m.cursor <= 0 || m.choices[m.cursor-1].Category != m.choices[m.cursor].Category {
		m.cursor = m.cursor + m.catStats.Amts[m.choices[m.cursor].Category] - 1
	} else {
		m.cursor--
	}
}

/*
Set the cursor of model `m` to either move down,
or wrap-around to the beginning of the list of choices.
*/
func (m *titleModel) setCursorWrapDown() {
	if m.cursor >= len(m.choices)-1 || m.choices[m.cursor+1].Category != m.choices[m.cursor].Category {
		m.cursor = m.cursor - m.catStats.Amts[m.choices[m.cursor].Category] + 1
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
