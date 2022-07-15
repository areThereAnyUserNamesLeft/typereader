package saving

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type SaveMsg struct {
	FileName    string
	ChunkNumber int
}

type LoadMsg struct {
	Saves Saves `yaml:"saves"`
}

type Saves map[string]int

func VerifySaveFile(saveFile string) bool {
	_, err := os.Stat(saveFile)
	if os.IsNotExist(err) {
		_, err := os.Create(saveFile)
		if err != nil {
			return false
		}
	} else if err != nil {
		return false
	}
	return true
}

func Save(msg SaveMsg, saveFile string, saves LoadMsg) error {
	if !VerifySaveFile(saveFile) {
		return fmt.Errorf("could not verify file")
	}
	if len(saves.Saves) == 0 {
		var s LoadMsg
		s, err := Load(saveFile)
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
	err = ioutil.WriteFile(saveFile, data, 0)
	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}
	return nil
}

func Load(saveFile string) (LoadMsg, error) {
	data, err := ioutil.ReadFile(saveFile)
	if err != nil {
		return LoadMsg{}, fmt.Errorf("could not read file: %w", err)
	}
	msg := LoadMsg{
		Saves: make(map[string]int),
	}
	err = yaml.Unmarshal(data, &msg)
	if err != nil {
		return LoadMsg{}, fmt.Errorf("could not unmarshall data to yaml from file: %w", err)
	}
	return msg, nil
}
