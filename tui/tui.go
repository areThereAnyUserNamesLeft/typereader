package tui

import (
	"github.com/areThereAnyUserNamesLeft/typereader/tui/menu"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/typing"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	Unknown State = iota
	Menu
	Type
)

type Model struct {
	WindowSize tea.WindowSizeMsg
	State      State
	ConfigPath string
	TextFile   string
	Typing     typing.Model
	Menu       menu.Model
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		if m.State == Menu {
			m.Menu.Update(msg)
		}
		if m.State == Type {
			m.Typing.Update(msg)
		}
		m.WindowSize = msg
	}
	switch m.State {
	case Menu:
		return m.Menu.Update(msg)
	case Type:
		return m.Typing.Update(msg)

	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.State {
	case Menu:
		return m.Menu.View()

	case Type:
		return m.Typing.View()

	default:
		return ""
	}
}
