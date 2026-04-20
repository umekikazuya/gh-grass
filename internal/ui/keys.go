package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap は MainModel 全体で共通利用するキーバインド定義。
type keyMap struct {
	Quit    key.Binding
	Confirm key.Binding
	Back    key.Binding
	Help    key.Binding
	DoneKey key.Binding // stateResult/stateError/stateHelp で終了に使う q キー
}

var keys = keyMap{
	Quit:    key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("ctrl+c/esc", "quit")),
	Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	Back:    key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "back")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	DoneKey: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

// helpEntries はヘルプ画面で表示するキー一覧。
func helpEntries() []key.Binding {
	return []key.Binding{keys.Confirm, keys.Back, keys.Help, keys.DoneKey, keys.Quit}
}
