package main

import (
	"bufio"
	"fmt"
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
	file             *os.File
	writer           *bufio.Writer
}

func (p *processCache) newDirectory(tag Tag) string {
	p.currentDirectory = append(p.currentDirectory, tag.Name)
	dirKey := strings.Join(p.currentDirectory, "/")
	if _, ok := p.directoryCounter[dirKey]; ok {
		p.directoryCounter[dirKey]++
		p.currentDirectory = append(p.currentDirectory, strconv.Itoa(p.directoryCounter[dirKey]))
	} else {
		p.directoryCounter[dirKey] = 0
		p.currentDirectory = append(p.currentDirectory, "0")
	}
	return fmt.Sprintf("%s/%d", dirKey, p.directoryCounter[dirKey])
}

func (p *processCache) newFile(tag Tag) string {
	filekey := strings.Join(p.currentDirectory, "/") + "/" + tag.Name
	if _, ok := p.fileCounter[filekey]; ok {
		p.fileCounter[filekey]++
	} else {
		p.fileCounter[filekey] = 0
	}
	return fmt.Sprintf("%s.%d.xml", filekey, p.fileCounter[filekey])
}
