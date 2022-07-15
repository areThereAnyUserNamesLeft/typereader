package menu

import (
	"fmt"

	"github.com/areThereAnyUserNamesLeft/typereader/state"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	width = 100.

	// charsPerWord is the average characters per word used by most typing tests
	// to calculate your WPM score.
	charsPerWord = 5.
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

var (
	wpms []float64
)

type Item struct {
	Filename, Desc string
	Filepath       *string
}

type Model struct {
	WindowSize tea.WindowSizeMsg
	WorkingDir string
	Options    []list.Item
	List       list.Model
	Chosen     string
	Parent     tea.Model
	HasChosen  bool
}

func (i Item) Title() string       { return i.Filename }
func (i Item) Description() string { return i.Desc }
func (i Item) FilterValue() string { return i.Filename }

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

type StateChangeMsg struct {
	State state.State
	KVs   map[string]string
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == tea.KeyEnter.String() {
			i, ok := m.List.SelectedItem().(Item)
			if ok {
				message := state.StateChangeMsg{
					State: state.Type,
					KVs:   map[string]string{"Filepath": *i.Filepath},
				}
				fmt.Printf("%#v", message)
				return m.Parent.Update(message)
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return docStyle.Render(m.List.View())
}
