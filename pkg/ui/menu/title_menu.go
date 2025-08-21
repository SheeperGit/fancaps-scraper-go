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
	keys "sheeper.com/fancaps-scraper-go/pkg/ui/menu/keys"
)

const menuLineFormat = "%c [%c] %s" // Format of a menu line.

/* Title Menu Model. */
type titleModel struct {
	/* Base Title Model fields. */

	tabList   TabList[types.Category] // Menu tabs.
	choices   []*types.Title          // Available choices.
	selected  []*types.Title          // Selected choices.
	confirmed bool                    // True if the user confirmed their selection, false otherwise.
	errMsg    string                  // Error message. If empty, no errors.
	keys      keys.TitleKeyMap        // Menu Keybinds.
	help      help.Model              // Help view.

	/* Cached View() fields. */

	menuWidth    int // Menu width.
	menuHeight   int // Menu height.
	contentWidth int // Content width.
}

/* Initializes the title model. */
func initialTitleModel(titles []*types.Title, menuLines uint8) titleModel {
	contentPadding := getContentPadding(menuLineFormat)
	contentWidth := lipgloss.Width(windowStyle.Render(ui.GetLongestTitle(titles))) + contentPadding - windowStyle.GetHorizontalPadding()

	/* Count up titles per category. */
	catStats := make(map[types.Category]int, len(types.CategoryName))
	for c := types.Category(0); c < types.Category(len(types.CategoryName)); c++ {
		catStats[c] = 0 // Initializing every statistic guarantees its existence in the map.
	}
	for _, t := range titles {
		catStats[t.Category]++
	}

	tabList := initTabList(catStats)

	menuWidth := lipgloss.Width(getTabRow(tabList.Tabs(), tabList.ActiveTab().id, contentWidth)) - windowStyle.GetHorizontalFrameSize()
	menuHeight := int(menuLines) + windowStyle.GetVerticalFrameSize()

	return titleModel{
		tabList:  tabList,
		choices:  titles,
		selected: []*types.Title{},
		keys:     keys.TitleKeys,
		help:     help.New(),

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
			choice := m.choices[m.tabList.ActiveTab().cursor]
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

	tabRow := getTabRow(m.tabList.Tabs(), m.tabList.ActiveTab().id, m.contentWidth)
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

/* Returns the render of the tab row at the top of the menu. */
func getTabRow(tabs []types.Category, activeTab types.Category, contentWidth int) string {
	tabCount := len(tabs)

	padding := (len(types.CategoryName) - tabCount) // Pad tab width proportional to the amount of missing tabs.
	tabWidth := (contentWidth / tabCount) + padding

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
		if m.tabList.ActiveTab().id == choice.Category {
			cursor := ' '
			if m.tabList.ActiveTab().cursor == i {
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
			if m.tabList.ActiveTab().cursor == i {
				style = style.Inherit(ui.HighlightStyle)
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
		if title.Url == t.Url {
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
func LaunchTitleMenu(titles []*types.Title, tabs []types.Category, menuLines uint8, debug bool) []*types.Title {
	p := tea.NewProgram(initialTitleModel(titles, menuLines))
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
					fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Url)
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
		if t.Url == title.Url {
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
		if t.Category == m.tabList.ActiveTab().id {
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
			if t.Category != m.tabList.ActiveTab().id {
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
	tabs, i := m.tabList.tabs, m.tabList.activeIndex
	if i == len(tabs)-1 { // If at the last tab going right, wrap to first.
		m.tabList.activeIndex = 0
	} else { // Otherwise, go to next tab.
		m.tabList.activeIndex = i + 1
	}
}

/*
Set the tab of model `m` to either move right,
or wrap-around to the beginning of the list of tabs.
*/
func (m *titleModel) setTabWrapRight() {
	tabs, i := m.tabList.tabs, m.tabList.activeIndex
	if i == len(tabs)-1 { // If at the last tab going right, go to first tab.
		m.tabList.activeIndex = 0
	} else { // Otherwise, go to next tab.
		m.tabList.activeIndex = i + 1
	}
}

/*
Set the cursor of model `m` to either move up,
or wrap-around to the end of the list of choices.
*/
func (m *titleModel) setCursorWrapUp() {
	pos, newPos := m.tabList.ActiveTab().cursor, -1
	if pos <= 0 || m.choices[pos-1].Category != m.choices[pos].Category {
		newPos = pos + m.tabList.stats[m.choices[pos].Category] - 1
	} else {
		newPos = pos - 1
	}

	m.tabList.ActiveTab().cursor = newPos
}

/*
Set the cursor of model `m` to either move down,
or wrap-around to the beginning of the list of choices.
*/
func (m *titleModel) setCursorWrapDown() {
	pos, newPos := m.tabList.ActiveTab().cursor, -1
	if pos >= len(m.choices)-1 || m.choices[pos+1].Category != m.choices[pos].Category {
		newPos = pos - m.tabList.stats[m.choices[pos].Category] + 1
	} else {
		newPos = pos + 1
	}

	m.tabList.ActiveTab().cursor = newPos
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
