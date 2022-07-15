package save

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type SaveMsg struct {
	FileName    string
	ChunkNumber int
}

type LoadMsg struct {
	Saves Saves `yaml:"saves"`
}

type Saves map[string]int

func Save(msg SaveMsg, configFile string, saves LoadMsg) error {
	if len(saves.Saves) == 0 {
		var s LoadMsg
		s, err := Load(configFile)
		if err != nil {
			return fmt.Errorf("could not load file: %w", err)
		}
		saves = s
	}
	saves.Saves[msg.FileName] = msg.ChunkNumber

	data, err := yaml.Marshal(&saves)
	if err != nil {
		return fmt.Errorf("could not marshal load message: %w", err)
	}
	err = ioutil.WriteFile("words.yaml", data, 0)
	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}
	return nil
}

func Load(configFile string) (LoadMsg, error) {
	data, err := ioutil.ReadFile("items.yaml")
	if err != nil {
		return LoadMsg{}, fmt.Errorf("could not read file: %w", err)
	}
	msg := LoadMsg{}
	err = yaml.Unmarshal(data, &msg)
	if err != nil {
		return LoadMsg{}, fmt.Errorf("could not unmarshall data to yaml from file: %w", err)
	}
	return msg, nil
}
