package menu

import "github.com/charmbracelet/bubbles/key"

/* Title Menu KeyMap. */
type TitleKeyMap struct {
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
func (k TitleKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

/* Keybindings to be shown in the expanded-help view. */
func (k TitleKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Toggle, k.ToggleAll, k.Confirm}, // First Column
		{k.Help, k.Quit}, // Second Column
	}
}

/* Keybinds for the Category Menu. */
var TitleKeys = TitleKeyMap{
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
