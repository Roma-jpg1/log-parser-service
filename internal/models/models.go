package models

type Log struct {
	ID       int
	Filename string
	Status   string
}

type Node struct {
	ID    int
	LogID int
	Name  string
	Type  string
}

type Port struct {
	ID     int
	NodeID int
	Name   string
	Status string
}

type NodeInfo struct {
	ID     int
	NodeID int
	Key    string
	Value  string
}
