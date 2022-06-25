package model

import (
	"fmt"
	"time"

	"github.com/areThereAnyUserNamesLeft/typereader/theme"
	// "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
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
	currentChunk int
	Choice       string // for dubugging
	Next         string // for dubugging
	spew         string // for dubugging
	Percent      float64
	Chunk        [][]rune
	Typed        []rune
	Start        time.Time
	Mistakes     int
	Score        float64
	Theme        *theme.Theme
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// Update updates the bubbletea model by handling the progress bar update
// and adding typed characters to the state if they are valid typing characters
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Start counting time only after the first keystroke
		if m.Start.IsZero() {
			m.Start = time.Now()
		}

		m.spew = ""

		// User wants to cancel the typing test
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if msg.String() == " " {
			m.spew = fmt.Sprint("it's a space")
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
		if len(m.Typed) >= len(m.Chunk[m.currentChunk]) {
			m.currentChunk++
			m.Typed = []rune{}
			return m, nil
		}

		char := msg.Runes[0]
		m.Choice = string(msg.Runes[0])
		next := rune(m.Chunk[m.currentChunk][len(m.Typed)])
		if len(m.Typed) == len(m.Chunk[m.currentChunk])-1 {
			m.Next = string(m.Chunk[m.currentChunk+1][0])
		} else {
			m.Next = string(m.Chunk[m.currentChunk][len(m.Typed)+1])
		}

		m.Typed = append(m.Typed, msg.Runes...)

		if char == next {
			m.Score += 1.
		}
		return m, nil

	case tea.WindowSizeMsg:
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
	remaining := m.Chunk[m.currentChunk][len(m.Typed):]

	var typed string
	for i, c := range m.Typed {
		if c == rune(m.Chunk[m.currentChunk][i]) {
			typed += m.Theme.StringColor(m.Theme.Text.Typed, string(c)).String()
		} else {
			typed += m.Theme.StringColor(m.Theme.Text.Error, string(m.Chunk[m.currentChunk][i])).String()
		}
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

	s := fmt.Sprintf(
		"%s%s\n\tWPM: %f",
		typed,
		m.Theme.StringColor(m.Theme.Text.Untyped, string(remaining)).Faint(),
		wpm,
	)

	return s
}
