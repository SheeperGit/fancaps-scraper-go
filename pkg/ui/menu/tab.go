package menu

type Tab[T ~int] struct {
	id     T   // Tab identifier.
	cursor int // Cursor position within the tab.
}

type TabList[T ~int] struct {
	tabs        []Tab[T]  // Info of all tabs.
	activeIndex int       // Active tab index in the tab list.
	stats       map[T]int // Tab statistics.
}

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
