package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type processCache struct {
	depth            int
	currentDirectory []string
	directoryCounter map[string]int
	fileCounter      map[string]int
	totalFiles       int
	innerText        string
	line             string
	file             bool
	ioActions []ioAction
}

type ioActionType int

const (
	writeFile ioActionType = iota
	newDirectory
)

type ioAction struct {
	actionType ioActionType
	path string
	lines []string
	ready bool
}

func (p *processCache) newDirectory(name string) {
	p.currentDirectory = append(p.currentDirectory, name)
	dirKey := strings.Join(p.currentDirectory, "/")
	if _, ok := p.directoryCounter[dirKey]; ok {
		p.directoryCounter[dirKey]++
		p.currentDirectory = append(p.currentDirectory, strconv.Itoa(p.directoryCounter[dirKey]))
	} else {
		p.directoryCounter[dirKey] = 0
		p.currentDirectory = append(p.currentDirectory, "0")
	}
	p.ioActions = append(p.ioActions, ioAction{actionType: newDirectory, path: fmt.Sprintf("%s/%d", dirKey, p.directoryCounter[dirKey]), ready: true})
}

func (p *processCache) exitDirectory() {
	p.currentDirectory = p.currentDirectory[:len(p.currentDirectory)-2]
}

func (p *processCache) openFile(prefix string) {
	filekey := strings.Join(p.currentDirectory, "/") + "/" + prefix
	if _, ok := p.fileCounter[filekey]; ok {
		p.fileCounter[filekey]++
	} else {
		p.fileCounter[filekey] = 0
	}
	p.ioActions = append(p.ioActions, ioAction{actionType: writeFile, path: fmt.Sprintf("%s.%d.xml", filekey, p.fileCounter[filekey]), lines: []string{xml.Header}})
	p.file = true
	p.totalFiles++
}

func (p *processCache) closeFile() {
	p.ioActions[len(p.ioActions)-1].ready = true
	p.file = false
}

func (p *processCache) appendLine(line string) {
	p.ioActions[len(p.ioActions)-1].lines = append(p.ioActions[len(p.ioActions)-1].lines, line)
}

func (p *processCache) appendFile(name, text string) {
	p.ioActions = append(p.ioActions, ioAction{actionType: writeFile, path: strings.Join(append(p.currentDirectory, name), "/") + ".xml", ready: true, lines: []string{xml.Header + text}})
	p.totalFiles++
}

func (p *processCache) flushIO() error {
	for len(p.ioActions) > 0 && p.ioActions[0].ready {
		action := p.ioActions[0]
		if action.ready {
			switch action.actionType {
			case writeFile:
				if err := ioutil.WriteFile(action.path, []byte(strings.Join(action.lines, "")), 0644); err != nil {
					return err
				}
			case newDirectory:
				if err := os.MkdirAll(action.path, 0755); err != nil {
					return err
				}
			}
			p.ioActions = p.ioActions[1:]
		}
	}
	return nil
}