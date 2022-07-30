package state

type State int

const (
	Unknown State = iota
	Menu
	Type
	Choose
)

type StateChangeMsg struct {
	State State
	KVs   map[string]string
}
