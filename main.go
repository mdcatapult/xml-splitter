package main

import (
	"bufio"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
		in    string
		out   string
		ext   string
		gzip  bool
		files int
		skip  *regexp.Regexp
		strip *regexp.Regexp
		level int
}

type XMLSplitter struct {
	path string
	conf Config
}

const (
	defaultSkip   = `(<\?xml)|(<!DOCTYPE)`
	openTagRegex  = `<([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)[\s]*([a-zA-Z]+=("[^"]*"|'[^']*')[\s]*)*>`
	emptyTagRegex = `<([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)[\s]*([a-zA-Z]+=("[^"]*"|'[^']*')[\s]*)*\/>`
	closeTagRegex = `<\/[\s]*([a-zA-Z:_]?[a-zA-Z0-9:_.-]*)[\s]*>`
)

var openTag = regexp.MustCompile(openTagRegex)
var closingTag = regexp.MustCompile(closeTagRegex)
var emptyTag = regexp.MustCompile(emptyTagRegex)

func GetConfig() (Config, error) {
	c := Config{}
	var skip, strip string
	flag.StringVar(&c.in, "in", "", "the folder to process (glob)")
	flag.StringVar(&c.out, "out", "", "the folder output to")
	flag.IntVar(&c.level, "depth", 1, "The nesting depth at which to split the XML")
	flag.StringVar(&c.ext, "ext", "xml", "file extension to process")
	flag.BoolVar(&c.gzip, "gzip", false, "use gzip to decompress files")
	flag.IntVar(&c.files, "files", 1, "number of files to process concurrently")
	flag.StringVar(&skip, "skip", defaultSkip, "regex for lines that should be skipped")
	flag.StringVar(&strip, "strip", "", "regex of values to trim from lines")
	flag.Parse()
	if len(c.in) == 0 || len(c.out) == 0 {
		flag.PrintDefaults()
		return c, errors.New("values must be provided for -in and -out")
	}
	c.strip = regexp.MustCompile(strip)
	c.skip = regexp.MustCompile(skip)
	return c, nil
}

func (s *XMLSplitter) WriteLines(lines []string, target string, suffix int) error {
	bytes := []byte(strings.Join(lines, "\n"))
	mkerr := os.MkdirAll(fmt.Sprintf("%s/%s/", strings.TrimRight(s.conf.out, "/"), target), 0755)
	handleError(mkerr)
	newFile := fmt.Sprintf("%s/%s/%d.xml", strings.TrimRight(s.conf.out, "/"), target, suffix)
	return ioutil.WriteFile(newFile, bytes, 0644)
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

type TagType int
const (
	Opening TagType = iota
	Closing
	Empty
)
func (tt TagType) String() string {
	return [...]string{"Opening", "Closing", "Empty"}[tt]
}

type Tag struct {
	Type  TagType
	Name  string
	Depth int
	Start int
	End   int
}

func (s *XMLSplitter) ProcessFile() int {
	scanner, serr := s.GetScanner(s.path)
	handleError(serr)

	currentDepth := 0
	//currentTag := ""

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

		lineStructure := make(map[int]Tag)

		tags := openTag.FindAllStringSubmatchIndex(line,-1)
		for _, tag := range tags {
			lineStructure[tag[0]] = Tag{
				Type:  Opening,
				Name:  line[tag[2]:tag[3]],
				Start: tag[0],
				End:   tag[1],
			}
		}

		tags = closingTag.FindAllStringSubmatchIndex(line, -1)
		for _, tag := range tags {
			lineStructure[tag[0]] = Tag{
				Type:  Closing,
				Name:  line[tag[2]:tag[3]],
				Start: tag[0],
				End:   tag[1],
			}
		}

		tags = emptyTag.FindAllStringSubmatchIndex(line, -1)
		for _, tag := range tags {
			lineStructure[tag[0]] = Tag{
				Type:  Empty,
				Name:  line[tag[2]:tag[3]],
				Start: tag[0],
				End:   tag[1],
			}
		}

		for i := 0; i < len(line); {
			if tag, ok := lineStructure[i]; ok {
				switch tag.Type {
				case Opening:
					currentDepth++
				case Closing:
					currentDepth--
				case Empty:
				default:
				}
				i = tag.End
			} else {
				i++
			}
		}
	}
	return 0
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

	path := fmt.Sprintf("%s/*.%s", strings.TrimRight(config.in, "/"), config.ext)
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
