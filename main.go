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

var workingDir = flag.String("in", "", "the folder to process (glob)")
var ext = flag.String("ext", "xml", "file extension to process")
var useGzip = flag.Bool("gzip", false, "use gzip to decompress files")
var outputDir = flag.String("out", "", "the folder output to")
var filesConcurrency = flag.Int("files", 1, "number of files to process concurrently")
var skipRegex = flag.String("skip", "(<?xml)|(<!DOCTYPE)", "regex for lines that should be skipped")
var stripRegex = flag.String("strip", "", "regex of values to trip from lines")
var splitRegex = flag.String("split",  "", "regex to split documents, matched expression will be written i.e. '(.*</PMC_ARTICLE>)(.*)'")



func handleError(err error) {
	if err != nil {
		panic(err)
	}
}


func processFile(filePath string, resolve func()) {

	var target = filepath.Base(strings.TrimSuffix(filePath, filepath.Ext(filePath)))

	var file, err = os.Open(filePath)
	handleError(err)
	defer file.Close()

	reader := bufio.NewReader(file)

	if *useGzip {
		target = strings.TrimSuffix(target, filepath.Ext(target))
		gunzip, gerr := gzip.NewReader(file)
		handleError(gerr)
		reader = bufio.NewReader(gunzip)
		defer gunzip.Close()
	}

	scanner := bufio.NewScanner(reader)

	lineCntr := 0
	fileCntr := 0
	var entry []string
	skip := regexp.MustCompile(*skipRegex)
	strip := regexp.MustCompile(*stripRegex)
	split := regexp.MustCompile(*splitRegex)
	for scanner.Scan() {
		lineCntr += 1
		var line = scanner.Text()
		if len(skip.FindStringSubmatch(line)) > 0  {
			continue
		}

		if len(*stripRegex) > 0 {
			line = strip.ReplaceAllString(line, "")
		}

		found := split.FindStringSubmatch(line)
		if len(found) >=1  {
			index := 0
			if len(found) == 3 {
				index = 1
			}
			entry = append(entry, found[index])
			bytes := []byte(strings.Join(entry, ""))
			mkerr := os.MkdirAll(fmt.Sprintf("%s/%s/", strings.TrimRight(*outputDir,"/"), target), 0755)
			handleError(mkerr)
			var newFile = fmt.Sprintf("%s/%s/%d.xml", strings.TrimRight(*outputDir,"/"), target, fileCntr)
			fmt.Println(newFile)
			werr := ioutil.WriteFile(newFile, bytes, 0644)
			handleError(werr)
			fileCntr += 1
			entry = entry[:0]
			if len(found) == 3 {
				entry = append(entry, found[2])
			}
		} else {
			entry = append(entry, line)
		}
	}

	resolve()
}


func check(key string, value *string) {
	if len(*value) == 0 {
		flag.PrintDefaults()
		fmt.Println()
		log.Fatal(fmt.Sprintf("'-%s' must be supplied", key))
	}
}

func main() {

	flag.Parse()
	check("in", workingDir)
	check("out", outputDir)
	check("split", splitRegex)

	path := fmt.Sprintf("%s/*.%s", strings.TrimRight(*workingDir, "/"), *ext)
	files, err := filepath.Glob(path)
	if err != nil {
		log.Panic(err)
	}

	fileSem := make(chan bool, *filesConcurrency)

	for _, filePath := range files {
		fileSem <- true
		go processFile(filePath, func() {
			<- fileSem
		})
	}
	for i := 0; i < cap(fileSem); i++ {
		fileSem <- true
	}

}

