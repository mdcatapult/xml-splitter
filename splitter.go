package main

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	defaultSkip       = "(<\\?xml)|(<!DOCTYPE)"
	openTagRegex      = "<([a-zA-Z:_][a-zA-Z0-9:_.-]*)(\\s*>|\\s+([a-zA-Z0-9:_.-]+\\s*=\\s*(\"[^\"]*\"|'[^']*')\\s*)*>)"
	emptyTagRegex     = "<([a-zA-Z:_][a-zA-Z0-9:_.-]*)(\\s*/>|\\s+([a-zA-Z0-9:_.-]+\\s*=\\s*(\"[^\"]*\"|'[^']*')\\s*)*/>)"
	closeTagRegex     = "</\\s*([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)\\s*>"
	whitespaceRegex   = "^\\s*$"
	openTagStartRegex = "<([a-zA-Z:_][a-zA-Z0-9:_.-]*)\\s+([a-zA-Z0-9:_.-]+\\s*=\\s*(\"[^\"]*\"|'[^']*')\\s*)*$"
	openTagEndRegex   = "^\\s*([a-zA-Z0-9:_.-]+\\s*=\\s*(\"[^\"]*\"|'[^']*')\\s*)*>"
)

var openingTag = regexp.MustCompile(openTagRegex)
var closingTag = regexp.MustCompile(closeTagRegex)
var emptyTag = regexp.MustCompile(emptyTagRegex)
var whitespace = regexp.MustCompile(whitespaceRegex)
var openTagStart = regexp.MustCompile(openTagStartRegex)
var openTagEnd = regexp.MustCompile(openTagEndRegex)

type TagType int

const (
	Opening TagType = iota
	Closing
	Empty
	OpeningTagStart
	OpeningTagEnd
)

type Tag struct {
	Type  TagType
	Name  string
	Full  string
	Start int
	End   int
}

func write(data string, name string, path ...string) {
	file := strings.Join(path, "/")
	file = fmt.Sprintf("%s/%s.xml", file, name)
	bytes := []byte(xml.Header + data)
	err := ioutil.WriteFile(file, bytes, 0644)
	handleError(err)
}

type XMLSplitter struct {
	path string
	conf Config
}

func (s *XMLSplitter) getLineStructure(line string) map[int]Tag {
	lineStructure := make(map[int]Tag)

	var NewTag = func(v []int, t TagType) Tag {
		return Tag{
			Type:  t,
			Name:  line[v[2]:v[3]],
			Full:  line[v[0]:v[1]],
			Start: v[0],
			End:   v[1],
		}
	}

	tags := openingTag.FindAllStringSubmatchIndex(line, -1)
	for _, tag := range tags {
		lineStructure[tag[0]] = NewTag(tag, Opening)
	}

	tags = closingTag.FindAllStringSubmatchIndex(line, -1)
	for _, tag := range tags {
		lineStructure[tag[0]] = NewTag(tag, Closing)
	}

	tags = emptyTag.FindAllStringSubmatchIndex(line, -1)
	for _, tag := range tags {
		lineStructure[tag[0]] = NewTag(tag, Empty)
	}

	tags = openTagStart.FindAllStringSubmatchIndex(line, -1)
	for _, tag := range tags {
		lineStructure[tag[0]] = NewTag(tag, OpeningTagStart)
	}

	tags = openTagEnd.FindAllStringSubmatchIndex(line, -1)
	for _, tag := range tags {
		lineStructure[tag[0]] = NewTag(tag, OpeningTagEnd)
	}

	return lineStructure
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

func (s *XMLSplitter) ProcessFile(scanner *bufio.Scanner) int {

	// cache is used to keep track of files/folders and xml depth so we don't overwrite files.
	cache := &processCache{
		currentDirectory: []string{s.conf.out, filepath.Base(strings.TrimSuffix(s.path, filepath.Ext(s.path)))},
		directoryCounter: make(map[string]int),
		fileCounter:      make(map[string]int),
	}

	isMultilineTag := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		skipMatches := s.conf.skip.FindStringSubmatch(line)
		if len(skipMatches) > 0 {
			continue
		}

		if openTagStart.MatchString(line) {
			isMultilineTag = true
			cache.line = line
			continue
		}

		if openTagEnd.MatchString(line) {
			line = cache.line + line
			cache.line = ""
			isMultilineTag = false
		}

		if isMultilineTag {
			cache.line += " " + line
			continue
		}

		if s.conf.strip.String() != "" {
			line = s.conf.strip.ReplaceAllString(line, "")
		}

		s.processLine(line, cache)

		if len(cache.ioActions) > 10 {
			err := cache.flushIO()
			handleError(err)
		}
	}

	err := cache.flushIO()
	handleError(err)

	return cache.totalFiles
}

func (s *XMLSplitter) processLine(line string, cache *processCache) {

	lineStructure := s.getLineStructure(line)

	for i := 0; i < len(line); {
		if tag, ok := lineStructure[i]; ok {

			if cache.file && !whitespace.MatchString(cache.innerText) {
				cache.appendLine(strings.TrimSpace(cache.innerText))
				cache.innerText = ""
			}

			switch tag.Type {
			case Opening:

				if cache.depth < s.conf.depth {
					cache.newDirectory(tag.Name)
					cache.appendFile("root", tag.Full[:len(tag.Full)-1]+"/>")
				} else if !cache.file {
					cache.openFile(tag.Name)
					cache.appendLine(tag.Full)
				} else {
					cache.appendLine(tag.Full)
				}
				cache.depth++

			case Closing:

				if cache.depth > s.conf.depth+1 {
					cache.appendLine(tag.Full)
				} else if cache.depth == s.conf.depth+1 {
					cache.appendLine(tag.Full)
					cache.closeFile()
				} else if tag.Name == cache.currentDirectory[len(cache.currentDirectory)-2] && cache.depth <= s.conf.depth {
					cache.exitDirectory()
				}
				cache.depth--

			case Empty:

				if cache.depth < s.conf.depth {
					cache.newDirectory(tag.Name)
					cache.appendFile("root", tag.Full)
					cache.exitDirectory()
				} else if !cache.file {
					cache.openFile(tag.Name)
					cache.appendLine(tag.Full)
					cache.closeFile()
				} else {
					cache.appendLine(tag.Full)
				}
			}
			i = tag.End

		} else {

			if cache.file {
				cache.innerText += string(line[i])
			}
			i++
			if cache.file && i == len(line) {
				cache.innerText += "\n"
			}

		}
	}
}
