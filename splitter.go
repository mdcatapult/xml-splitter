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
	defaultSkip   = `(<\?xml)|(<!DOCTYPE)`
	openTagRegex  = `<([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)\s*([a-zA-Z0-9:_.-]+\s*=\s*("[^"]*"|'[^']*')\s*)*>`
	emptyTagRegex = `<([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)\s*([a-zA-Z0-9:_.-]+\s*=\s*("[^"]*"|'[^']*')\s*)*\/>`
	closeTagRegex = `<\/\s*([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)\s*>`
	whitespaceRegex = `^\s*$`
)

var openTag = regexp.MustCompile(openTagRegex)
var closingTag = regexp.MustCompile(closeTagRegex)
var emptyTag = regexp.MustCompile(emptyTagRegex)
var whitespace = regexp.MustCompile(whitespaceRegex)

type TagType int

const (
	Opening TagType = iota
	Closing
	Empty
)

type Tag struct {
	Type  TagType
	Name  string
	Full  string
	Start int
	End   int
}

func tabs(depth int) string {
	return strings.Repeat("  ", depth)
}

func writeFile(data string, name string, path ...string) {
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

func (s *XMLSplitter) GetLineStructure(line string) map[int]Tag {
	lineStructure := make(map[int]Tag)

	var NewTag = func (v []int, t TagType) Tag {
		return Tag{
			Type: t,
			Name: line[v[2]:v[3]],
			Full: line[v[0]:v[1]],
			Start: v[0],
			End: v[1],
		}
	}

	tags := openTag.FindAllStringSubmatchIndex(line, -1)
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

	return lineStructure
}

func (s *XMLSplitter) GetScanner(target string) (*bufio.Scanner, error) {
	var scanner *bufio.Scanner
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("File '%s' not Found", target))
	}
	file, err := os.Open(target)
	handleError(err)

	if s.conf.gzip {
		target = strings.TrimSuffix(target, filepath.Ext(target))
		gunzip, gerr := gzip.NewReader(file)
		handleError(gerr)

		scanner = bufio.NewScanner(bufio.NewReader(gunzip))
	} else {
		scanner = bufio.NewScanner(file)
	}

	return scanner, nil
}

func (s *XMLSplitter) ProcessFile() int {
	scanner, serr := s.GetScanner(s.path)
	handleError(serr)

	// cache is used to keep track of files/folders and xml depth so we don't overwrite files.
	cache := processCache{
		currentDirectory: []string{s.conf.out, filepath.Base(strings.TrimSuffix(s.path, filepath.Ext(s.path)))},
		directoryCounter: make(map[string]int),
		fileCounter:      make(map[string]int),
	}
	
	var currentFile *os.File
	var writer *bufio.Writer

	// Convenience function is scoped to this function since all I/O occurs here.
	var closeFile = func () {
		err := writer.Flush()
		handleError(err)
		err = currentFile.Close()
		handleError(err)
		currentFile = nil
		writer = nil
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		skipMatches := s.conf.skip.FindStringSubmatch(line)
		if len(skipMatches) > 0 {
			continue
		}

		if s.conf.strip.String() != "" {
			line = s.conf.strip.ReplaceAllString(line, "")
		}

		lineStructure := s.GetLineStructure(line)

		for i := 0; i < len(line); {
			if tag, ok := lineStructure[i]; ok {
				if currentFile != nil && !whitespace.MatchString(cache.innerText) {

					// We have reached a new xml tag and have recorded some text so write it to file.
					_, err := writer.WriteString(tabs(cache.depth-s.conf.depth) + strings.TrimSpace(cache.innerText) + "\n")
					handleError(err)
					cache.innerText = ""
				}
				switch tag.Type {
				case Opening:
					if cache.depth < s.conf.depth {

						// We have reached an opening tag that is below the depth we wish to split at.
						// Make a new directory to contain more deeply nested data.
						// Write the opening tag as an empty tag in it's own "root.xml" file inside this folder.
						directory := cache.newDirectory(tag)
						err := os.MkdirAll(directory, 0755)
						handleError(err)
						writeFile(tag.Full[:len(tag.Full)-1]+"/>", "root", cache.currentDirectory...)
						cache.totalFiles++
					} else if currentFile == nil {

						// We have reached an opening tag that is at or above the depth at which we wish to split the file.
						// Open a new file and write the xml header and opening tag.
						var err error
						file := cache.newFile(tag)
						currentFile, err = os.Create(file)
						handleError(err)
						writer = bufio.NewWriter(currentFile)
						_, err = writer.WriteString(xml.Header + tabs(cache.depth-s.conf.depth) + tag.Full + "\n")
						handleError(err)
						cache.totalFiles++
					} else {

						// We have an opening tag but aleady have a file open.
						// We are above the split depth so we can just write to file.
						_, err := writer.WriteString(tabs(cache.depth-s.conf.depth) + tag.Full + "\n")
						handleError(err)
					}
					cache.depth++

				case Closing:
					if cache.depth > s.conf.depth + 1 {

						// If we are more than one level above the split depth we must have an open file and so can just write.
						_, err := writer.WriteString(tabs(cache.depth-s.conf.depth-1) + tag.Full + "\n")
						handleError(err)
					} else if cache.depth == s.conf.depth + 1 {

						// Closing tag one level above the split depth => closes root tag of the file.
						// Close the file after writing the tag.
						_, err := writer.WriteString(tabs(cache.depth-s.conf.depth-1) + tag.Full + "\n")
						handleError(err)
						closeFile()
					} else if tag.Name == cache.currentDirectory[len(cache.currentDirectory)-2] && cache.depth <= s.conf.depth {

						// Closing tag for the containing directory.
						cache.currentDirectory = cache.currentDirectory[:len(cache.currentDirectory)-2]
					}
					cache.depth--

				case Empty:
					if cache.depth < s.conf.depth {

						// Empty tag below the split depth.
						// Create a directory indicating the tag and it's index then write it to file.
						directory := cache.newDirectory(tag)
						err := os.MkdirAll(directory, 0755)
						handleError(err)
						writeFile(tag.Full, "root", cache.currentDirectory...)
						cache.totalFiles++
						cache.currentDirectory = cache.currentDirectory[:len(cache.currentDirectory)-2]
					} else if currentFile == nil {

						// Empty tag above the split depth.
						var err error
						file := cache.newFile(tag)
						currentFile, err = os.Create(file)
						handleError(err)
						writer = bufio.NewWriter(currentFile)
						_, err = writer.WriteString(xml.Header + tabs(cache.depth-s.conf.depth) + tag.Full + "\n")
						handleError(err)
						closeFile()
						cache.totalFiles++
					} else {
						_, err := writer.WriteString(tabs(cache.depth-s.conf.depth) + tag.Full + "\n")
						handleError(err)
					}
				}

				// Set i to the index of the line at the end of the tag.
				i = tag.End
			} else {
				// Capture text as long as we have an open file to write to
				if currentFile != nil {
					cache.innerText += string(line[i])
				}
				i++
			}
		}
	}
	return cache.totalFiles
}