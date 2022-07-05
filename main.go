package main

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strings"

	// "github.com/areThereAnyUserNamesLeft/typereader/tui"
	"github.com/areThereAnyUserNamesLeft/typereader/tui"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/menu"
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

func FromFile(path string) (string, error) {
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func main() {
	app := &cli.App{
		Name:  "typereader",
		Usage: "read as you type",
		Flags: []cli.Flag{},
		Action: func(cCtx *cli.Context) error {
			termenv.ClearScreen()
			text, err := FromFile(cCtx.Args().First())
			if err != nil {
				panic(err)
			}
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

			menu := NewMenu()
			program := tea.NewProgram(
				tui.Model{
					TextFile: cCtx.Args().First(),
					State:    tui.Menu,
					Menu:     menu,
				},
			)
			eg, _ := errgroup.WithContext(context.Background())
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

func NewMenu() menu.Model {
	m := menu.Model{}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	m.WorkingDir = wd
	files, err := ioutil.ReadDir(wd)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	files = remove(files)
	m.Positions = make([]list.Item, len(files))
	for k, v := range files {
		p := menu.Item{}
		// p.Filepath = wd + "/" + v.Name()
		p.Desc = wd + "/" + v.Name()
		p.Filename = v.Name()
		m.Positions[k] = p
	}

	m.List = list.New(m.Positions, list.NewDefaultDelegate(), 0, 0)
	m.List.Title = "Files"
	return m
}

func remove(files []fs.FileInfo) []fs.FileInfo {
	for k, v := range files {
		if v.IsDir() {
			return remove(append(files[:k], files[k+1:]...))
		}
	}
	return files
}
