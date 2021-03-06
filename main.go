package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	in    string
	out   string
	files int
	skip  *regexp.Regexp
	strip *regexp.Regexp
	depth int
	buffer int
}

func GetConfig() (Config, error) {
	c := Config{}
	var skip, strip, in, out string
	flag.StringVar(&in, "in", "", "the folder to process (glob)")
	flag.StringVar(&out, "out", "", "the folder output to")
	flag.IntVar(&c.depth, "depth", 1, "the nesting depth at which to split the XML")
	flag.IntVar(&c.files, "files", 1, "number of files to process concurrently")
	flag.StringVar(&skip, "skip", defaultSkip, "regex for lines that should be skipped")
	flag.StringVar(&strip, "strip", "", "regex of values to strip from lines")
	flag.IntVar(&c.buffer, "buffer", 20, "max number of files to hold in buffer before writing")
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

// Generic function to handle errors
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func getFiles(directory string) []string {
	var files []string
	path := fmt.Sprintf("%s/*.xml*", directory)
	xmlfiles, err := filepath.Glob(path)
	if err != nil {
		log.Panic(err)
	}
	path = fmt.Sprintf("%s/*.owl*", directory)
	owlfiles, err := filepath.Glob(path)
	if err != nil {
		log.Panic(err)
	}
	files = append(xmlfiles, owlfiles...)
	return files
}

func main() {
	config, err := GetConfig()
	if err != nil {
		return
	}

	files := getFiles(config.in)
	fileSem := make(chan bool, config.files)
	for _, path := range files {
		fileSem <- true
		go func(path string) {
			s := XMLSplitter{path: path, conf: config}
			scanner, err := getScanner(s.path, strings.HasSuffix(s.path, ".gz"))
			handleError(err)
			filesCreated := s.ProcessFile(scanner, &writer{})
			fmt.Printf("%d files generated from %s\n", filesCreated, path)
			<-fileSem
		}(path)
	}
	for i := 0; i < cap(fileSem); i++ {
		fileSem <- true
	}
}
