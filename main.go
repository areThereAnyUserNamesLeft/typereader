package main

import (
	"fmt"
	"log"
	"os"

	"github.com/areThereAnyUserNamesLeft/typereader/saving"
	"github.com/areThereAnyUserNamesLeft/typereader/state"
	"github.com/areThereAnyUserNamesLeft/typereader/theme"
	"github.com/areThereAnyUserNamesLeft/typereader/tui"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/choose"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/menu"
	"github.com/areThereAnyUserNamesLeft/typereader/tui/typing"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	words        = 15
	defaultWidth = 60
)

var (
	flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config-directory",
			Usage: "Location of your saves and configration files",
			Value: "$HOME/.config/typereader/",
		},
		&cli.BoolFlag{
			Name:    "use-saves",
			Usage:   "Use saves rather than choosing from current directory",
			Value:   false,
			Aliases: []string{"s", "S"},
		},
	}
)

func main() {
	app := &cli.App{
		Name:  "typereader",
		Usage: "read as you type",
		Flags: flags,
		Action: func(cCtx *cli.Context) error {
			termenv.ClearScreen()
			text := ""
			configDir := os.ExpandEnv(cCtx.String("config-directory"))
			err := createConfigDir(configDir)
			if err != nil {
				return fmt.Errorf("could not create config dir: %w", err)
			}
			saveFile := fmt.Sprintf("%s/saves.yaml", configDir)
			if !saving.VerifySaveFile(saveFile) {
				return fmt.Errorf("could not create savefile")
			}
			saves, err := saving.Load(fmt.Sprintf("%s/saves.yaml", configDir))
			if err != nil {

			}
			if cCtx.Args().First() != "" {
				text, err = tui.FromFile(os.ExpandEnv(cCtx.Args().First()))
				if err != nil {
					fmt.Printf("Not a valid filepath %s", cCtx.Args().First())
				}
			}
			// Replace out all weird quotes for keyboard friendly alternatives
			program := &tea.Program{}
			m := tui.Model{
				TextFile: cCtx.Args().First(),
				Typing: typing.Model{
					Saves:    saves,
					SaveFile: saveFile,
					Theme:    theme.DefaultTheme(),
				},
			}
			if cCtx.Bool("use-saves") {
				// if we have more than one save - choose it
				m.State = state.Choose
				m.Choose = choose.New(saves.Saves)
				m.Choose.Parent = &m
				program = tea.NewProgram(m)
			} else if text != "" {
				// if we have opted for a text file - use it
				m := m.HandleText(text)
				m.State = state.Type
				m.TextFile = cCtx.Args().First()
				program = tea.NewProgram(m)

			} else {
				// if the first argument is a dir chose from that otherwise use current dir
				m.Menu.Parent = &m
				m.State = state.Menu
				dirMenu, err := menu.NewDirMenu(cCtx.Args().First())
				if err != nil {
					dirMenu, _ = menu.NewDirMenu("")
				}
				m.Menu = &dirMenu
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

func createConfigDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not mkdir config dir at %s: %w", dir, err)
		}
	}
	if err != nil {
		return err
	}
	return nil
}
