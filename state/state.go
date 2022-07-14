package state

type State int

const (
	Unknown State = iota
	Menu
	Type
)

type StateChangeMsg struct {
	State State
	KVs   map[string]string
}
