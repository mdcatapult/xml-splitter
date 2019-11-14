package main

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ioActionType int

const (
	writeFile ioActionType = iota
	newDirectory
)

type ioAction struct {
	actionType ioActionType
	path       string
	lines      []string
	ready      bool
}

type ioActionWriter interface {
	write([]ioAction) ([]ioAction, error)
}

type writer struct{}

func (w *writer) write(actions []ioAction) ([]ioAction, error) {
	for len(actions) > 0 && actions[0].ready {
		action := actions[0]
		switch action.actionType {
		case writeFile:
			if err := ioutil.WriteFile(action.path, []byte(strings.Join(action.lines, "")), 0644); err != nil {
				return nil, err
			}
		case newDirectory:
			if err := os.MkdirAll(action.path, 0755); err != nil {
				return nil, err
			}
		}
		actions = actions[1:]
	}
	return actions, nil
}

func getScanner(target string, isZipped bool) (*bufio.Scanner, error) {
	var scanner *bufio.Scanner
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("File '%s' not Found", target))
	}
	file, err := os.Open(target)
	handleError(err)

	if isZipped {
		target = strings.TrimSuffix(target, filepath.Ext(target))
		gunzip, gerr := gzip.NewReader(file)
		handleError(gerr)

		scanner = bufio.NewScanner(bufio.NewReader(gunzip))
	} else {
		scanner = bufio.NewScanner(file)
	}

	return scanner, nil
}
