package menu

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

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

func NewDirMenu(dir string) (Model, error) {
	s := ""
	m := Model{
		WindowSize: tea.WindowSizeMsg{},
		WorkingDir: dir,
		Options:    []list.Item{},
		List:       list.Model{},
		Chosen:     s,
	}
	wd, err := os.Getwd()
	if err != nil {
		return m, fmt.Errorf("could not get working dir: %w", err)
	}
	if dir != "" {
		wd = dir
	}
	m.WorkingDir = wd
	files, err := ioutil.ReadDir(wd)

	if err != nil {
		return m, fmt.Errorf("failed to list directory: %w", err)
	}
	files = remove(files)
	m.Options = make([]list.Item, len(files))
	for k, v := range files {
		p := Item{}
		refString := fmt.Sprintf("%s/%s", wd, v.Name())
		p.Filepath = &refString
		p.Desc = wd + "/" + v.Name()
		p.Filename = v.Name()
		m.Options[k] = p
	}
	m.List = list.New(m.Options, list.NewDefaultDelegate(), 0, 0)
	m.List.Title = "Please choose your file"
	return m, nil
}

func remove(files []fs.FileInfo) []fs.FileInfo {
	for k, v := range files {
		if v.IsDir() {
			return remove(append(files[:k], files[k+1:]...))
		}
	}
	return files
}
