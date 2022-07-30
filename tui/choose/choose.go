package choose

import (
	"fmt"
	"strings"

	"github.com/areThereAnyUserNamesLeft/typereader/saving"
	"github.com/areThereAnyUserNamesLeft/typereader/state"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type Model struct {
	Parent      tea.Model
	Options     Options
	Items       []Item
	quitting    bool
	index       int
	limit       int
	numSelected int
	Paginator   paginator.Model

	// styles
	cursorStyle       lipgloss.Style
	itemStyle         lipgloss.Style
	selectedItemStyle lipgloss.Style
}

type Item struct {
	Text     string
	FilePath string
	Position int
	Selected bool
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

		m.Options.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		start, end := m.Paginator.GetSliceBounds(len(m.Items))
		switch keypress := msg.String(); keypress {
		case "down", "j", "ctrl+n":
			m.index = clamp(m.index+1, 0, len(m.Items)-1)
			if m.index >= end {
				m.Paginator.NextPage()
			}
		case "up", "k", "ctrl+p":
			m.index = clamp(m.index-1, 0, len(m.Items)-1)
			if m.index <= start {
				m.Paginator.PrevPage()
			}
		case "right", "l", "ctrl+f":
			m.index = clamp(m.index+m.Options.Height, 0, len(m.Items)-1)
			m.Paginator.NextPage()
		case "left", "h", "ctrl+b":
			m.index = clamp(m.index-m.Options.Height, 0, len(m.Items)-1)
			m.Paginator.PrevPage()
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case " ", "x":
			if m.limit == 1 {
				break // no op
			}

			if m.Items[m.index].Selected {
				m.Items[m.index].Selected = false
				m.numSelected--
			} else if m.numSelected < m.limit {
				m.Items[m.index].Selected = true
				m.numSelected++
			}
		case "enter":
			m.quitting = true
			// If the user hasn't selected any items in a multi-select.
			// Then we select the item that they have pressed enter on. If they
			// have selected items, then we simply return them.
			if m.numSelected < 1 {
				m.Items[m.index].Selected = true
			}
			kvs := make(map[string]string)
			kvs["Filepath"] = m.Items[m.index].FilePath
			kvs["Position"] = fmt.Sprintf("%d", m.Items[m.index].Position)
			return m.Parent.Update(state.StateChangeMsg{
				State: state.Type,
				KVs:   kvs,
			})

		}
	}

	var cmd tea.Cmd
	m.Paginator, cmd = m.Paginator.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	start, end := m.Paginator.GetSliceBounds(len(m.Items))
	for i, item := range m.Items[start:end] {
		if i == m.index%m.Options.Height {
			s.WriteString(m.Options.CursorStyle.Render(m.Options.Cursor))
		} else {
			s.WriteString(strings.Repeat(" ", runewidth.StringWidth(m.Options.Cursor)))
		}

		if item.Selected {
			s.WriteString(m.Options.SelectedItemStyle.Render(m.Options.SelectedPrefix + item.Text))
		} else if i == m.index%m.Options.Height {
			s.WriteString(m.Options.CursorStyle.Render(m.Options.CursorPrefix + item.Text))
		} else {
			s.WriteString(m.Options.ItemStyle.Render(m.Options.UnselectedPrefix + item.Text))
		}
		if i != m.Options.Height {
			s.WriteRune('\n')
		}
	}

	if m.Paginator.TotalPages <= 1 {
		return s.String()
	}

	s.WriteString(strings.Repeat("\n", m.Options.Height-m.Paginator.ItemsOnPage(len(m.Items))+1))
	s.WriteString("  " + m.Paginator.View())

	return s.String()
}

func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// Options is the customization options for the choose command.
type Options struct {
	Options []Choice `arg:"" optional:"" help:"Options to choose from."`

	Limit             int            `help:"Maximum number of options to pick" default:"1" group:"Selection"`
	NoLimit           bool           `help:"Pick unlimited number of options (ignores limit)" group:"Selection"`
	Height            int            `help:"Height of the list" default:"10"`
	Cursor            string         `help:"Prefix to show on item that corresponds to the cursor position" default:"> "`
	CursorPrefix      string         `help:"Prefix to show on the cursor item (hidden if limit is 1)" default:"[•] "`
	SelectedPrefix    string         `help:"Prefix to show on selected items (hidden if limit is 1)" default:"[✕] "`
	UnselectedPrefix  string         `help:"Prefix to show on selected items (hidden if limit is 1)" default:"[ ] "`
	CursorStyle       lipgloss.Style `embed:"" prefix:"cursor." set:"defaultForeground=212" set:"name=indicator"`
	ItemStyle         lipgloss.Style `embed:"" prefix:"item." hidden:"" set:"defaultForeground=255" set:"name=item"`
	SelectedItemStyle lipgloss.Style `embed:"" prefix:"selected." set:"defaultForeground=212" set:"name=selected item"`
}

type Choice struct {
	Text     string
	Filepath string
	Position int
}

var (
	subduedStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#847A85", Dark: "#979797"})
	verySubduedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"})
)

func New(s saving.Saves) (m Model) {
	m.Options.Options = append(m.Options.Options, MakeChoices(s)...)
	// create new chooser
	pager := paginator.New()
	pager.PerPage = 4
	pager.Type = paginator.Dots
	pager.ActiveDot = subduedStyle.Render("•")
	pager.InactiveDot = verySubduedStyle.Render("•")
	// disable default keys - i defined my own in the chooser
	pager.UseHLKeys = false
	pager.UseLeftRightKeys = false
	pager.UseJKKeys = false
	pager.UsePgUpPgDownKeys = false

	var items []Item // empties
	m.Options.Height = 10
	m.Options.Limit = 1
	m.Options.Cursor = ">"
	m.Options.SelectedPrefix = "[✕] "
	m.Options.CursorPrefix = "[•] "
	m.Options.UnselectedPrefix = "[ ] "
	m.Options.CursorStyle = lipgloss.NewStyle().SetString(`embed:"" prefix:"cursor." set:"defaultforeground=212" set:"name=indicator"`)
	m.Options.ItemStyle = lipgloss.NewStyle().SetString(`embed:"" prefix:"item." hidden:"" set:"defaultforeground=255" set:"name=item"`)
	m.Options.SelectedItemStyle = lipgloss.NewStyle().SetString(`embed:"" prefix:"selected." set:"defaultforeground=212" set:"name=selected item"`)
	for _, option := range m.Options.Options {
		items = append(items, Item{
			Text:     option.Text,
			FilePath: option.Filepath,
			Position: option.Position,
			Selected: false,
		})
	}
	m.Paginator = pager
	m.Items = items
	return m
}

func MakeChoices(mp map[string]int) (out []Choice) {
	for k, v := range mp {
		ch := Choice{}
		ch.Text = fmt.Sprintf("FILE: %s - LOCATION %d", k, v)
		ch.Position = v
		ch.Filepath = k
		out = append(out, ch)
	}
	return
}
