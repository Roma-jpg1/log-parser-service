package models

import "time"

type Log struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Status     string    `json:"status"`
	ErrorText  string    `json:"error_text,omitempty"`
	NodesCount int       `json:"nodes_count"`
	PortsCount int       `json:"ports_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type Node struct {
	ID       int    `json:"id"`
	LogID    int    `json:"log_id"`
	SourceID string `json:"source_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

type Port struct {
	ID       int    `json:"id"`
	LogID    int    `json:"log_id"`
	NodeID   int    `json:"node_id"`
	SourceID string `json:"source_id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

type NodeInfo struct {
	ID       int    `json:"id"`
	NodeID   int    `json:"node_id"`
	SourceID string `json:"source_id,omitempty"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type ParsedLog struct {
	Nodes     []Node
	Ports     []Port
	NodeInfos []NodeInfo
}
