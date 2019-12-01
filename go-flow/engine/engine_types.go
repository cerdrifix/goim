package engine

import "github.com/google/uuid"

type ProcessParameter struct {
	Name          string `json:"name"`
	ParameterType string `json:"type"`
	Value         string `json:"value"`
}

type ProcessEvent struct {
	EventType  string             `json:"type"`
	Name       string             `json:"name"`
	Parameters []ProcessParameter `json:"parameters"`
}

type ProcessEventsPrePost struct {
	Pre  []ProcessEvent `json:"pre"`
	Post []ProcessEvent `json:"post"`
}

type ProcessTransaction struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	To          string               `json:"to"`
	Events      ProcessEventsPrePost `json:"events"`
}

type TimeSpan struct {
	Unit  string `json:"unit"`
	Value int    `json:"value"`
}

type ProcessTrigger struct {
	Name     string         `json:"name"`
	TimeSpan TimeSpan       `json:"timeSpan"`
	Events   []ProcessEvent `json:"event"`
}

type ProcessNode struct {
	Name         string               `json:"name"`
	NodeType     string               `json:"type"`
	Events       ProcessEventsPrePost `json:"events"`
	Triggers     []ProcessTrigger     `json:"triggers"`
	Transactions []ProcessTransaction `json:"transactions"`
}

type ProcessMap struct {
	Id          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Nodes       []ProcessNode `json:"nodes"`
}

type CreateProcessPayload struct {
	ProcessName string                 `json:"processName"`
	Variables   map[string]interface{} `json:"variables"`
}
