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
		split string
		ext   string
		gzip  bool
		files int
		skip  string
		strip string
		level int
}

type XMLSplitter struct {
	path string
	conf Config
}

const (
	defaultSkip   = "(<?xml)|(<!DOCTYPE)"
	fileDelimiter = "LINE_DELIMITER"
)

func GetConfig() (Config, error) {
	c := Config{}
	flag.StringVar(&c.in, "in", "", "the folder to process (glob)")
	flag.StringVar(&c.out, "out", "", "the folder output to")
	flag.StringVar(&c.split, "split", "", "The XML closing tag to split after i.e. '</Entry>'")
	flag.StringVar(&c.ext, "ext", "xml", "file extension to process")
	flag.BoolVar(&c.gzip, "gzip", false, "use gzip to decompress files")
	flag.IntVar(&c.files, "files", 1, "number of files to process concurrently")
	flag.StringVar(&c.skip, "skip", defaultSkip, "regex for lines that should be skipped")
	flag.StringVar(&c.strip, "strip", "", "regex of values to trip from lines")
	flag.Parse()
	if len(c.in) == 0 || len(c.out) == 0 || len(c.split) == 0 {
		flag.PrintDefaults()
		return c, errors.New("values must be provided for -in, -out & -split")
	}
	return c, nil
}

func (s *XMLSplitter) GetLines(line string) []string {
	var lines []string
	skip := regexp.MustCompile(s.conf.skip)
	strip := regexp.MustCompile(s.conf.strip)
	split := regexp.MustCompile(s.conf.split)

	if line == "" {
		return lines
	}

	if len(skip.FindStringSubmatch(line)) > 0 {
		return lines
	}

	if len(s.conf.strip) > 0 {
		line = strip.ReplaceAllString(line, "")
	}

	found := split.FindAllStringSubmatchIndex(line, -1)
	if len(found) >= 0 {
		previous := 0
		for _, v := range found {
			lines = append(lines, line[previous:v[1]])
			lines = append(lines, fileDelimiter)
			previous = v[1]
		}
		if len(line[previous:]) > 0 {
			lines = append(lines, line[previous:])
		}
	} else {
		lines = append(lines, line)
	}
	return lines
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

func (s *XMLSplitter) ProcessFile() int {
	scanner, serr := s.GetScanner(s.path)
	handleError(serr)

	target := filepath.Base(strings.TrimSuffix(s.path, filepath.Ext(s.path)))
	lineCntr := 0
	fileCntr := 0
	var lines []string

	for scanner.Scan() {
		lineCntr += 1
		newlines := s.GetLines(scanner.Text())
		if len(newlines) > 1 {
			for _, v := range newlines {
				if v == fileDelimiter {
					werr := s.WriteLines(lines, target, fileCntr)
					handleError(werr)
					fileCntr += 1
					lines = lines[:0]
					continue
				}
				lines = append(lines, v)
			}
		} else {
			lines = append(lines, newlines...)
		}
	}

	return fileCntr
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
