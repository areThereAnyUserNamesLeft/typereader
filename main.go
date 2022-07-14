package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"github.com/areThereAnyUserNamesLeft/typereader/state"
	"github.com/areThereAnyUserNamesLeft/typereader/theme"
	"github.com/areThereAnyUserNamesLeft/typereader/tui"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/menu"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/typing"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	words        = 15
	defaultWidth = 60
)

func main() {
	app := &cli.App{
		Name:  "typereader",
		Usage: "read as you type",
		Flags: []cli.Flag{},
		Action: func(cCtx *cli.Context) error {
			termenv.ClearScreen()
			text, err := tui.FromFile(cCtx.Args().First())
			if err != nil {
				fmt.Println("Not a valid filepath %s", cCtx.Args().First())
			}
			menu, err := NewMenu(cCtx.Args().First())
			if err != nil {
				menu, err = NewMenu("")
			}
			// Replace out all weird quotes for keyboard friendly alternatives
			program := &tea.Program{}
			if text != "" {
				m := tui.Model{
					WindowSize: tea.WindowSizeMsg{},
					State:      state.Type,
					TextFile:   cCtx.Args().First(),
					Menu:       &menu,
					Typing: typing.Model{
						WindowSize: tea.WindowSizeMsg{},
						Theme: &theme.Theme{
							Text: theme.DefaultTheme().Text,
						},
					},
				}.HandleText(text)
				program = tea.NewProgram(m)
			} else {
				m := tui.Model{
					TextFile: cCtx.Args().First(),
					State:    state.Menu,
					Menu:     &menu,
					Typing: typing.Model{
						Theme: theme.DefaultTheme(),
					},
				}
				m.Menu.Parent = &m
				program = tea.NewProgram(m)

			}
			eg, _ := errgroup.WithContext(cCtx.Context)

			eg.Go(func() error {
				return program.Start()
			})
			return eg.Wait()
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func NewMenu(dir string) (menu.Model, error) {
	s := ""
	m := menu.Model{
		WindowSize: tea.WindowSizeMsg{},
		WorkingDir: dir,
		Positions:  []list.Item{},
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
	m.Positions = make([]list.Item, len(files))
	for k, v := range files {
		p := menu.Item{}
		refString := fmt.Sprintf("%s/%s", wd, v.Name())
		p.Filepath = &refString
		p.Desc = wd + "/" + v.Name()
		p.Filename = v.Name()
		m.Positions[k] = p
	}
	m.List = list.New(m.Positions, list.NewDefaultDelegate(), 0, 0)
	m.List.Title = "Files"
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
