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
	}
)

func main() {
	app := &cli.App{
		Name:  "typereader",
		Usage: "read as you type",
		Flags: flags,
		Action: func(cCtx *cli.Context) error {
			termenv.ClearScreen()
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
			text := ""
			if cCtx.Args().First() != "" {
				text, err = tui.FromFile(cCtx.Args().First())
				if err != nil {
					fmt.Printf("Not a valid filepath %s", cCtx.Args().First())
				}
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
				Menu:     &dirMenu,
				Typing: typing.Model{
					Saves:    saves,
					SaveFile: saveFile,
					Theme:    theme.DefaultTheme(),
				},
			}
			// if we have opted for a text file - use it
			if text != "" {
				m.HandleText(text)
				m.State = state.Type
				program = tea.NewProgram(m)
				// if we have more than one save - choose it
			} else if len(saves.Saves) >= 1 {
				m.State = state.Choose
				m.Choose = choose.New(saves.Saves)
				m.Choose.Parent = &m
				program = tea.NewProgram(m)

			} else {
				m.State = state.Menu
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
