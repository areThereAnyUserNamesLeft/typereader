package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/areThereAnyUserNamesLeft/typereader/model"
	"github.com/areThereAnyUserNamesLeft/typereader/theme"
	tea "github.com/charmbracelet/bubbletea"
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
		Flags: []cli.Flag{
			// &cli.StringFlag{
			// 	Name:     "lang",
			// 	Value:    "english",
			// 	Usage:    "language for the greeting",
			// 	Required: true,
			// },
		},
		Action: func(cCtx *cli.Context) error {
			text, err := FromFile(cCtx.Args().Get(0))
			if err != nil {
				panic(err)
			}
			text = strings.ReplaceAll(text, "’", "'")
			chunks := [][]rune{}
			texts := strings.Split(text, "\n\n")

			for i := range texts {
				text = strings.Trim(texts[i], "\n")
				chunks = append(chunks, []rune(text))
			}

			program := tea.NewProgram(model.Model{
				Chunk: chunks,
				Theme: theme.DefaultTheme(),
			})
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
