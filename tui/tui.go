package tui

import (
	"fmt"
	"io/ioutil"

	"github.com/areThereAnyUserNamesLeft/typereader/state"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/menu"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/typing"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	WindowSize tea.WindowSizeMsg
	State      state.State
	ConfigPath string
	TextFile   string
	Typing     typing.Model
	Menu       *menu.Model
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
	case state.StateChangeMsg:
		if msg.State == state.Type {
			m.State = msg.State
			m.TextFile = msg.KVs["Filepath"]
			text, err := FromFile(msg.KVs["Filepath"])
			if err != nil {
				fmt.Println("this is not a valid filepath %s", msg.KVs["Filepath"])
			}
			m.HandleText(text)
			txtMsg := typing.TextUpdateMsg{
				Text: text,
			}
			return m.Typing.Update(txtMsg)
		}
	case tea.WindowSizeMsg:
		if m.State == state.Menu {
			m.Menu.Update(msg)
		}
		if m.State == state.Type {
			m.Typing.Update(msg)
		}
		m.WindowSize = msg
	}
	switch m.State {
	case state.Menu:
		return m.Menu.Update(msg)
	case state.Type:
		return m.Typing.Update(msg)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.State {
	case state.Menu:
		return m.Menu.View()

	case state.Type:
		return m.Typing.View()

	default:
		return ""
	}
}

func (m Model) HandleText(text string) Model {
	t, c := typing.HandleText(text)
	m.TextFile = t
	m.Typing.Chunks = c
	return m
}

func FromFile(path string) (string, error) {
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(text), nil
}
