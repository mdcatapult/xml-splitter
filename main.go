package main

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultSkip   = `(<\?xml)|(<!DOCTYPE)`
	openTagRegex  = `<([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)[\s]*([a-zA-Z]+[\s]*=[\s]*("[^"]*"|'[^']*')[\s]*)*>`
	emptyTagRegex = `<([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)[\s]*([a-zA-Z]+[\s]*=[\s]*("[^"]*"|'[^']*')[\s]*)*\/>`
	closeTagRegex = `<\/[\s]*([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)[\s]*>`
)

var openTag = regexp.MustCompile(openTagRegex)
var closingTag = regexp.MustCompile(closeTagRegex)
var emptyTag = regexp.MustCompile(emptyTagRegex)

type Config struct {
	in    string
	out   string
	gzip  bool
	files int
	skip  *regexp.Regexp
	strip *regexp.Regexp
	depth int
}

type XMLSplitter struct {
	path string
	conf Config
}

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
	Depth int
	Start int
	End   int
}

func GetConfig() (Config, error) {
	c := Config{}
	var skip, strip, in, out string
	flag.StringVar(&in, "in", "", "the folder to process (glob)")
	flag.StringVar(&out, "out", "", "the folder output to")
	flag.IntVar(&c.depth, "depth", 1, "The nesting depth at which to split the XML")
	flag.BoolVar(&c.gzip, "gzip", false, "use gzip to decompress files")
	flag.IntVar(&c.files, "files", 1, "number of files to process concurrently")
	flag.StringVar(&skip, "skip", defaultSkip, "regex for lines that should be skipped")
	flag.StringVar(&strip, "strip", "", "regex of values to trim from lines")
	flag.Parse()
	if len(in) == 0 || len(out) == 0 {
		flag.PrintDefaults()
		return Config{}, errors.New("values must be provided for -in and -out")
	}
	if c.depth < 1 {
		return Config{}, errors.New("depth must be greater than or equal to 1")
	}
	c.in = strings.TrimRight(in, "/")
	c.out = strings.TrimRight(out, "/")
	c.strip = regexp.MustCompile(strip)
	c.skip = regexp.MustCompile(skip)
	return c, nil
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

	currDepth := 0
	currDir := []string{s.conf.out, filepath.Base(strings.TrimSuffix(s.path, filepath.Ext(s.path)))}
	dirCounter := make(map[string]int)
	fileCounter := make(map[string]int)
	totalFiles := 0
	var currFile *os.File
	var writer *bufio.Writer

	var createDirectory = func(tag Tag) {
		currDir = append(currDir, tag.Name)
		dir := strings.Join(currDir, "/")
		if _, ok := dirCounter[dir]; ok {
			dirCounter[dir]++
			currDir = append(currDir, strconv.Itoa(dirCounter[dir]))
		} else {
			dirCounter[dir] = 0
			currDir = append(currDir, "0")
		}
		err := os.MkdirAll(strings.Join(currDir, "/"), 0755)
		handleError(err)
	}

	var openFileAndWriter = func(tag Tag) {
		var err error
		filepart := strings.Join(currDir, "/") + "/" + tag.Name
		if _, ok := fileCounter[filepart]; ok {
			fileCounter[filepart]++
		} else {
			fileCounter[filepart] = 0
		}
		currFile, err = os.Create(fmt.Sprintf("%s.%d.xml", filepart, fileCounter[filepart]))
		handleError(err)
		writer = bufio.NewWriter(currFile)
	}

	var closeFileAndWriter = func() {
		err := writer.Flush()
		handleError(err)
		err = currFile.Close()
		handleError(err)
		currFile = nil
	}

	var tabs = func(tagType TagType) string {
		if tagType == Closing {
			return strings.Repeat("  ", currDepth-s.conf.depth-1)
		}
		return strings.Repeat("  ", currDepth-s.conf.depth)
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

		lineStructure := GetLineStructure(line)

		for i := 0; i < len(line); {
			if tag, ok := lineStructure[i]; ok {
				switch tag.Type {
				case Opening:
					if currDepth < s.conf.depth {
						createDirectory(tag)
						Write(tag.Full[:len(tag.Full)-1]+"/>", "root", currDir...)
						totalFiles++
					} else if currFile == nil {
						openFileAndWriter(tag)
						_, err := writer.WriteString(xml.Header + tabs(Opening) + tag.Full + "\n")
						handleError(err)
						totalFiles++
					} else {
						_, err := writer.WriteString(tabs(Opening) + tag.Full + "\n")
						handleError(err)
					}
					currDepth++

				case Closing:
					if currDepth > s.conf.depth+1 {
						_, err := writer.WriteString(tabs(Closing) + tag.Full + "\n")
						handleError(err)
					} else if currDepth == s.conf.depth+1 {
						_, err := writer.WriteString(tabs(Closing) + tag.Full + "\n")
						handleError(err)
						closeFileAndWriter()
					}

					if tag.Name == currDir[len(currDir)-2] && currDepth == s.conf.depth {
						currDir = currDir[:len(currDir)-2]
					}
					currDepth--

				case Empty:
					if currDepth < s.conf.depth {
						createDirectory(tag)
						Write(tag.Full, "root", currDir...)
						totalFiles++
						currDir = currDir[:len(currDir)-2]
					} else if currFile == nil {
						openFileAndWriter(tag)
						_, err := writer.WriteString(xml.Header + tabs(Empty) + tag.Full + "\n")
						handleError(err)
						closeFileAndWriter()
						totalFiles++
					} else {
						_, err := writer.WriteString(tabs(Empty) + tag.Full + "\n")
						handleError(err)
					}
				}
				i = tag.End
			} else {
				i++
			}
		}
	}
	return totalFiles
}

func GetLineStructure(line string) map[int]Tag {
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

func Write(data string, name string, path ...string) {
	file := strings.Join(path, "/")
	file = fmt.Sprintf("%s/%s.xml", file, name)
	bytes := []byte(xml.Header + data)
	err := ioutil.WriteFile(file, bytes, 0644)
	handleError(err)
}

// Generic function to handle errors
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	config, err := GetConfig()
	handleError(err)

	path := fmt.Sprintf("%s/*.xml", config.in)
	files, err := filepath.Glob(path)
	if err != nil {
		log.Panic(err)
	}

	fileSem := make(chan bool, config.files)
	for _, path := range files {
		fileSem <- true
		go func() {
			s := XMLSplitter{path: path, conf: config}
			filesCreated := s.ProcessFile()
			fmt.Println(fmt.Sprintf("%d files generated from %s", filesCreated, path))
			<-fileSem
		}()
	}
	for i := 0; i < cap(fileSem); i++ {
		fileSem <- true
	}
}
