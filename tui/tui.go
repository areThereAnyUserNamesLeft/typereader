package tui

import (
	"github.com/areThereAnyUserNamesLeft/typereader/tui/typing"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	Menu State = iota
	Type
)

type Model struct {
	State      State
	ConfigPath string
	TextFile   string
	Typing     typing.Model
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	switch m.State {
	// case Menu:
	// 	return m.Menu.Update(msg)

	case Type:
		return m.Typing.Update(msg)

	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.State {
	// case Menu:
	// 	return  m.Menu.View()

	case Type:
		return m.Typing.View()

	default:
		return ""
	}
}
