package consensus

import "encoding/json"

type CommandType int

const (
	CommandPut CommandType = iota
)

type PutCommand struct {
	Key   string
	Value string
}

type command struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
