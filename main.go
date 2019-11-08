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
	gzip  bool
	files int
	skip  *regexp.Regexp
	strip *regexp.Regexp
	depth int
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

// Generic function to handle errors
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	config, err := GetConfig()
	if err != nil {
		return
	}

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
