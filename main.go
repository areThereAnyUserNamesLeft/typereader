package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"github.com/areThereAnyUserNamesLeft/typereader/saving"
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

var (
	flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config-directory",
			Usage: "Location of your saves and configration files",
			Value: "$HOME/.config/typereader/",
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
			err := createConfigDir(cCtx.String("config-directory"))
			if err != nil {
				return fmt.Errorf("could not create config dir: %w", err)
			}
			saveFile := fmt.Sprintf("%s/saves.yaml", cCtx.String("config-directory"))
			if !saving.VerifySaveFile(saveFile) {
				return fmt.Errorf("could not create savefile")
			}
			saves, err := saving.Load(fmt.Sprintf("%s/saves.yaml", cCtx.String("config-directory")))
			if err != nil {
				return fmt.Errorf("could not load saves: %w", err)
			}
			text, err := tui.FromFile(cCtx.Args().First())
			if err != nil {
				fmt.Println("Not a valid filepath %s", cCtx.Args().First())
			}
			if err != nil {
				fmt.Println("Not a valid filepath %s", cCtx.Args().First())
			}
			dirMenu, err := menu.NewDirMenu(cCtx.Args().First())
			if err != nil {
				dirMenu, err = menu.NewDirMenu("")
			}
			// Replace out all weird quotes for keyboard friendly alternatives
			program := &tea.Program{}
			m := tui.Model{
				TextFile: cCtx.Args().First(),
				State:    state.Menu,
				Menu:     &menu,
				Typing: typing.Model{
					Saves:    saves,
					SaveFile: saveFile,
					Theme:    theme.DefaultTheme(),
				},
			}
			if text != "" {
				m.HandleText(text)
				m.State = state.Type
				program = tea.NewProgram(m)
			} else {
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

func createConfigDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return fmt.Errorf("could not mkdir config dir at %s: %w", dir, err)
		}
	}
	if err != nil {
		return err
	}
	return nil
}
