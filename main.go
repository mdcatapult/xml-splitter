package main

import (
	"bufio"
	"compress/gzip"
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
}

// container for command line arguments
var config = Config{
	*flag.String("in", "", "the folder to process (glob)"),
	*flag.String("out", "", "the folder output to"),
	*flag.String("split", "", "regex to split documents, matched expression will be written i.e. '(.*</Entry>)(.*)'"),
	*flag.String("ext", "xml", "file extension to process"),
	*flag.Bool("gzip", false, "use gzip to decompress files"),
	*flag.Int("files", 1, "number of files to process concurrently"),
	*flag.String("skip", "(<?xml)|(<!DOCTYPE)", "regex for lines that should be skipped"),
	*flag.String("strip", "", "regex of values to trip from lines"),
}

// loads arguments from command line
func checkConfig() {
	if len(config.in) == 0 || len(config.out) == 0 || len(config.split) == 0 {
		flag.PrintDefaults()
		fmt.Println()
		log.Fatal(fmt.Sprintf("Values must be provided for -in, -out & -split"))
	}
}

// Generic function to handle errors
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func addLines(line string, lines []string) []string {
	skip := regexp.MustCompile(config.skip)
	strip := regexp.MustCompile(config.strip)
	split := regexp.MustCompile(config.split)

	if line == "" {
		return lines
	}

	if len(skip.FindStringSubmatch(line)) > 0 {
		return lines
	}

	if len(config.strip) > 0 {
		line = strip.ReplaceAllString(line, "")
	}

	found := split.FindStringSubmatch(line)
	if len(found) >= 1 {
		index := 0
		if len(found) == 3 {
			index = 1
		}
		lines = append(lines, found[index])
		if len(found) == 3 {
			lines = append(lines, found[2])
		}
		lines = append(lines, "")
	} else {
		lines = append(lines, line)
	}
	return lines
}

func writeLines(lines []string, target string, suffix int) error {
	bytes := []byte(strings.Join(lines, ""))
	mkerr := os.MkdirAll(fmt.Sprintf("%s/%s/", strings.TrimRight(config.out, "/"), target), 0755)
	handleError(mkerr)
	newFile := fmt.Sprintf("%s/%s/%d.xml", strings.TrimRight(config.out, "/"), target, suffix)
	fmt.Println(newFile)
	return ioutil.WriteFile(newFile, bytes, 0644)
}

func getScanner(target string) *bufio.Scanner {
	file, err := os.Open(target)
	handleError(err)
	defer file.Close()

	reader := bufio.NewReader(file)

	if config.gzip {
		target = strings.TrimSuffix(target, filepath.Ext(target))
		gunzip, gerr := gzip.NewReader(file)
		handleError(gerr)
		reader = bufio.NewReader(gunzip)
		defer gunzip.Close()
	}

	return bufio.NewScanner(reader)
}

func processFile(filePath string, resolve func()) {

	target := filepath.Base(strings.TrimSuffix(filePath, filepath.Ext(filePath)))
	scanner := getScanner(target)

	lineCntr := 0
	fileCntr := 0
	var lines []string
	for scanner.Scan() {
		lineCntr += 1
		lines = addLines(scanner.Text(), lines)
		if lines[len(lines)-1] == "" {
			werr := writeLines(lines, target, fileCntr)
			handleError(werr)
			fileCntr += 1
			lines = lines[:0]
		}
	}

	resolve()
}

func main() {

	flag.Parse()
	checkConfig()

	path := fmt.Sprintf("%s/*.%s", strings.TrimRight(config.in, "/"), config.ext)
	files, err := filepath.Glob(path)
	if err != nil {
		log.Panic(err)
	}

	fileSem := make(chan bool, config.files)

	for _, filePath := range files {
		fileSem <- true
		go processFile(filePath, func() {
			<-fileSem
		})
	}
	for i := 0; i < cap(fileSem); i++ {
		fileSem <- true
	}

}
