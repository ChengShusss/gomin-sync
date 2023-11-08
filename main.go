package main

import (
	"fmt"
	command "gomin-sync/internal/command"
	"os"
)

type OperationFunc func()

type Operation struct {
	Name string
	Desp string
	Func OperationFunc
}

type OperationSet struct {
	set map[string]*Operation
}

func (s *OperationSet) AddOperation(name, desp string, f OperationFunc) {
	s.set[name] = &Operation{name, desp, f}
}

func (s *OperationSet) ParseAndHandle(operation string) {
	op, ok := s.set[operation]
	if !ok {
		s.PrintInfo()
		return
	}
	op.Func()
}

var operationSet = OperationSet{
	set: map[string]*Operation{},
}

func (s *OperationSet) PrintInfo() {
	maxlen := 0
	for _, op := range s.set {
		if len(op.Name) > maxlen {
			maxlen = len(op.Name) + 5
		}
	}

	for _, op := range s.set {
		format := fmt.Sprintf("%%-%ds%%s\n", maxlen)
		fmt.Printf(format, op.Name, op.Desp)
	}

}

func main() {
	operationSet.AddOperation(
		"sync",
		"sync dir with remote repo",
		command.SyncDir)

	if len(os.Args) == 1 {
		operationSet.PrintInfo()
		return
	}

	operationSet.ParseAndHandle(os.Args[1])
}
