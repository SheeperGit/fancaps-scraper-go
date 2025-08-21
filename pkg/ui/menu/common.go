package menu

import (
	// "github.com/charmbracelet/bubbles/help"
	// "github.com/charmbracelet/bubbles/key"

	"github.com/charmbracelet/lipgloss"
)

type Tab[T ~int] struct {
	id     T   // Tab identifier.
	cursor int // Cursor position within the tab.
}

type TabList[T ~int] struct {
	tabs        []Tab[T]  // Info of all tabs.
	activeIndex int       // Active tab index in the tab list.
	stats       map[T]int // Tab statistics.
}

/* Format of a menu line. */
const menuLineFormat = "%c [%c] %s"

/* Menu Styles. */
var inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7"))

/*
Returns a TabList from tab statistics `tabStats`.
`tabStats` is assumed to contain at least one element with an item.
*/
func initTabList[T ~int](stats map[T]int) TabList[T] {
	tabs, cursor := []Tab[T]{}, 0
	for id := T(0); id < T(len(stats)); id++ {
		if tabItems := stats[id]; tabItems > 0 {
			tabs = append(tabs, Tab[T]{
				id:     id,
				cursor: cursor,
			})
			cursor += tabItems // Update next initial tab cursor position.
		}
	}

	return TabList[T]{
		tabs:        tabs,
		activeIndex: 0,
		stats:       stats,
	}
}

/* Returns a list of tabs from a TabList. */
func (tl *TabList[T]) Tabs() []T {
	out := make([]T, len(tl.tabs))
	for i := range tl.tabs {
		out[i] = tl.tabs[i].id
	}

	return out
}

/* Returns the active tab from a TabList. */
func (tl *TabList[T]) ActiveTab() *Tab[T] {
	return &tl.tabs[tl.activeIndex]
}

// /* Base key map. */
// type baseKeyMap struct {
// 	Up        key.Binding
// 	Down      key.Binding
// 	Toggle    key.Binding
// 	ToggleAll key.Binding
// 	Confirm   key.Binding
// 	Help      key.Binding
// 	Quit      key.Binding
// }

// /* Base/Default keybindings to be shown in the mini-help view. */
// func (k baseKeyMap) ShortHelp() []key.Binding {
// 	return []key.Binding{k.Help, k.Quit}
// }

// /* Base/Default keybindings to be shown in the expanded-help view. */
// func (k baseKeyMap) FullHelp(extraPrimaryKeys ...key.Binding) [][]key.Binding {
// 	primaryKeys := []key.Binding{k.Up, k.Down, k.Toggle, k.ToggleAll, k.Confirm}
// 	primaryKeys = append(primaryKeys, extraPrimaryKeys...)

// 	return [][]key.Binding{
// 		primaryKeys,      // First Column
// 		{k.Help, k.Quit}, // Second Column
// 	}
// }

// /* Base/Default keybinds. */
// var baseKeys = baseKeyMap{
// 	Up: key.NewBinding(
// 		key.WithKeys("up", "w", "k"),
// 		key.WithHelp("↑/w/k", "move up"),
// 	),
// 	Down: key.NewBinding(
// 		key.WithKeys("down", "s", "j"),
// 		key.WithHelp("↓/s/j", "move down"),
// 	),
// 	Toggle: key.NewBinding(
// 		key.WithKeys("enter", " "),
// 		key.WithHelp("enter/[space]", "toggle selection"),
// 	),
// 	ToggleAll: key.NewBinding(
// 		key.WithKeys("t"),
// 		key.WithHelp("t", "toggle all"),
// 	),
// 	Confirm: key.NewBinding(
// 		key.WithKeys("p"),
// 		key.WithHelp("p", "confirm"),
// 	),
// 	Help: key.NewBinding(
// 		key.WithKeys("?"),
// 		key.WithHelp("?", "toggle help"),
// 	),
// 	Quit: key.NewBinding(
// 		key.WithKeys("q", "esc", "ctrl+c"),
// 		key.WithHelp("q", "quit"),
// 	),
// }

// /* Base Tea Model. */
// type baseModel[T comparable] struct {
// 	keys       baseKeyMap
// 	help       help.Model
// 	inputStyle lipgloss.Style
// 	cursor     int
// 	choices    []T
// 	selected   map[T]struct{}
// 	confirmed  bool
// 	errMsg     string
// }

// /* Initializes the base model. */
// func initialBaseModel[T comparable](choices []T) baseModel[T] {
// 	return baseModel[T]{
// 		keys:       baseKeys,
// 		help:       help.New(),
// 		inputStyle: inputStyle,
// 		choices:    choices,
// 		selected:   make(map[T]struct{}),
// 	}
// }

// /*
// Set the cursor of model `m` to either move up,
// or wrap-around to the end of the list of choices.
// */
// func (m *baseModel[T]) setCursorWrapUp() {
// 	if m.cursor <= 0 {
// 		m.cursor = len(m.choices) - 1
// 	} else {
// 		m.cursor--
// 	}
// }

// /*
// Set the cursor of model `m` to either move down,
// or wrap-around to the beginning of the list of choices.
// */
// func (m *baseModel[T]) setCursorWrapDown() {
// 	if m.cursor >= len(m.choices)-1 {
// 		m.cursor = 0
// 	} else {
// 		m.cursor++
// 	}
// }
