package typing

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/areThereAnyUserNamesLeft/typereader/saving"
	"github.com/areThereAnyUserNamesLeft/typereader/theme"

	// "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	width = 100.

	// charsPerWord is the average characters per word used by most typing tests
	// to calculate your WPM score.
	charsPerWord = 5.
)

var (
	wpms []float64
)

type Model struct {
	Parent     *tea.Model
	WindowSize tea.WindowSizeMsg
	SaveMsg    saving.SaveMsg
	Saves      saving.LoadMsg
	TextFile   string
	SaveFile   string
	Choice     string // for dubugging
	Next       string // for dubugging
	spew       string // for dubugging
	Percent    float64
	Chunks     [][]rune
	Typed      []rune
	Start      time.Time
	Mistakes   int
	Score      float64
	Theme      *theme.Theme
}

type TextUpdateMsg struct {
	TextFile  string
	Paragraph int
	Text      string
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// Update updates the bubbletea model by handling the progress bar update
// and adding typed characters to the state if they are valid typing characters
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TextUpdateMsg:
		m.TextFile = msg.TextFile
		_, c := msg.HandleText()
		m.Chunks = c
		m.SaveMsg.ChunkNumber = msg.Paragraph
		return m, nil
	case tea.KeyMsg:
		// Start counting time only after the first keystroke
		if m.Start.IsZero() {
			m.Start = time.Now()
		}

		// User wants to cancel the typing test
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if msg.String() == " " {
			msg.Runes = []rune{' '}
		}

		// Deleting characters
		if msg.Type == tea.KeyBackspace && len(m.Typed) > 0 {
			m.Typed = m.Typed[:len(m.Typed)-1]
		}

		// Ensure we are adding characters only that we want the user to be able to type
		// There may be need to add some outliers like I did with <SPACE>
		if msg.Type != tea.KeyRunes && msg.String() != " " {
			return m, nil
		}

		// Bounce to the next chunk when we are done with the current one
		if len(m.Typed) >= len(m.Chunks[m.SaveMsg.ChunkNumber]) {
			m.Percent = 0
			m.Typed = []rune{}

			m.SaveMsg.ChunkNumber++
			absoluteFilename, err := filepath.Abs(m.TextFile)
			if err != nil {
				panic(fmt.Sprintf("cannot get absolute filepath for: %s: err = %s", m.TextFile, err.Error()))
			}
			m.SaveMsg.FileName = absoluteFilename
			err = saving.Save(
				m.SaveMsg,
				m.SaveFile,
				m.Saves,
			)
			if err != nil {
				panic("could not save progress: " + err.Error())
			}
			return m, nil
		}

		char := msg.Runes[0]
		m.Choice = string(msg.Runes[0])
		next := rune(m.Chunks[m.SaveMsg.ChunkNumber][len(m.Typed)])
		if len(m.Typed) == len(m.Chunks[m.SaveMsg.ChunkNumber])-1 {
			m.Next = string(m.Chunks[m.SaveMsg.ChunkNumber+1][0])
		} else {
			m.Next = string(m.Chunks[m.SaveMsg.ChunkNumber][len(m.Typed)+1])
		}

		m.Typed = append(m.Typed, msg.Runes...)

		if char == next {
			m.Score += 1.
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.WindowSize = msg
		return m, nil
	default:
		return m, nil
	}
}

// View shows the current state of the typing test.
// It displays a progress bar for the progression of the typing test,
// the typed characters (with errors displayed in red) and remaining
// characters to be typed in a faint display
func (m Model) View() string {
	remaining := m.Chunks[m.SaveMsg.ChunkNumber][len(m.Typed):]

	var typed string
	for i, c := range m.Typed {
		if c == rune(m.Chunks[m.SaveMsg.ChunkNumber][i]) {
			typed += m.Theme.StringColor(m.Theme.Text.Typed, string(c)).String()
		} else if c == ' ' && rune(m.Chunks[m.SaveMsg.ChunkNumber][i]) == '\n' { // && c == ' ' {
			typed += m.Theme.StringColor(m.Theme.Text.Typed, string('\n')).String()
		} else {
			typed += m.Theme.StringColor(m.Theme.Text.Error, string(m.Chunks[m.SaveMsg.ChunkNumber][i])).String()
		}
		m.spew = fmt.Sprintf("c = '%s'", string(c))
	}

	var wpm float64
	// Start counting wpm after at least two characters are typed
	if len(m.Typed) > 1 {
		wpm = (m.Score / charsPerWord) / (time.Since(m.Start).Minutes())
	}

	if len(m.Typed) > charsPerWord {
		wpms = append(wpms, wpm)
	}

	wpmsCount := wpms
	if len(wpmsCount) <= 0 {
		wpmsCount = []float64{0}
	}

	text := fmt.Sprintf(
		"%s%s\n",
		typed,
		m.Theme.StringColor(m.Theme.Text.Untyped, string(remaining)).Faint(),
	)
	wpmText := fmt.Sprintf(
		"WPM: %f",
		wpm,
	)
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("202")).Align(lipgloss.Center).Width(m.WindowSize.Width)

	return style.Render(text) + "\n" + style.Render(wpmText) + "\n" + style.Render(fmt.Sprintf("\n Paragraph: %d/%d", m.SaveMsg.ChunkNumber, len(m.Chunks)))
}

func (t TextUpdateMsg) HandleText() (string, [][]rune) {
	return HandleText(t.Text)
}

func HandleText(text string) (string, [][]rune) {
	// Replace out all weird quotes for keyboard friendly alternatives
	text = strings.ReplaceAll(text, "’", "'")
	text = strings.ReplaceAll(text, "“", "\"")
	text = strings.ReplaceAll(text, "”", "\"")
	text = strings.ReplaceAll(text, "—", "-")
	chunks := [][]rune{}
	// Break text to be typed one paragraph at a time
	texts := strings.Split(text, "\n\n")

	for i := range texts {
		// Trim out the other new lines
		text = strings.Trim(texts[i], "\n")
		chunks = append(chunks, []rune(text))
	}
	return text, chunks
}
