package main

import (
	"encoding/xml"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CacheSuite struct {
	suite.Suite
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}

func (s *CacheSuite) TestNewDirectory() {
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		cache *processCache
		args   args
		want *processCache
	}{
		{
			name: "XML root",
			cache: &processCache{
				currentDirectory: []string{"output-folder", "target-folder"},
				directoryCounter: make(map[string]int),
			},
			args: args{
				name: "xml-tag",
			},
			want: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				directoryCounter: map[string]int{
					"output-folder/target-folder/xml-tag": 0,
				},
				ioActions: []ioAction{
					{actionType: newDirectory, path: "output-folder/target-folder/xml-tag/0", ready: true},
				},
			},
		},
		{
			name: "repeated tag",
			cache: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				directoryCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				name: "repeated-tag",
			},
			want: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0", "repeated-tag", "2"},
				directoryCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 2,
				},
				ioActions: []ioAction{
					{actionType: newDirectory, path: "output-folder/target-folder/xml-tag/0/repeated-tag/2", ready: true},
				},
			},
		},
		{
			name: "new-tag",
			cache: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				directoryCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				name: "new-tag",
			},
			want: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0", "new-tag", "0"},
				directoryCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
					"output-folder/target-folder/xml-tag/0/new-tag": 0,
				},
				ioActions: []ioAction{
					{actionType: newDirectory, path: "output-folder/target-folder/xml-tag/0/new-tag/0", ready: true},
				},
			},
		},
	}
	for _, tt := range tests {
		tt.cache.newDirectory(tt.args.name)
		s.Assert().Equal(tt.want, tt.cache)
	}
}

func (s *CacheSuite) TestNewFile() {
	type args struct {
		name string
	}
	tests := []struct {
		cache *processCache
		args   args
		want   *processCache
	}{
		{
			cache: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				fileCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				name: "repeated-tag",
			},
			want: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				fileCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 2,
				},
				ioActions: []ioAction{
					{actionType: writeFile, path: "output-folder/target-folder/xml-tag/0/repeated-tag.2.xml", lines: []string{xml.Header}},
				},
				totalFiles: 1,
				file: true,
			},
		},
		{
			cache: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				fileCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				name: "new-tag",
			},
			want: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				fileCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
					"output-folder/target-folder/xml-tag/0/new-tag": 0,
				},
				ioActions: []ioAction{
					{actionType: writeFile, path: "output-folder/target-folder/xml-tag/0/new-tag.0.xml", lines: []string{xml.Header}},
				},
				totalFiles: 1,
				file: true,
			},
		},
	}
	for _, tt := range tests {
		tt.cache.openFile(tt.args.name)
		s.Assert().Equal(tt.want, tt.cache)
	}
}

func (s *CacheSuite) TestExitDirectory() {
	tests := []struct {
		cache *processCache
		want   *processCache
	}{
		{
			cache: &processCache{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
			},
			want: &processCache{
				currentDirectory: []string{"output-folder", "target-folder"},
			},
		},
	}
	for _, tt := range tests {
		tt.cache.exitDirectory()
		s.Assert().Equal(tt.want, tt.cache)
	}
}

func (s *CacheSuite) TestCloseFile() {
	tests := []struct {
		cache *processCache
		want   *processCache
	}{
		{
			cache: &processCache{
				ioActions: []ioAction{
					{actionType: writeFile},
				},
				file: true,
			},
			want: &processCache{
				ioActions: []ioAction{
					{actionType: writeFile, ready: true},
				},
				file: false,
			},
		},
	}
	for _, tt := range tests {
		tt.cache.closeFile()
		s.Assert().Equal(tt.want, tt.cache)
	}
}

func (s *CacheSuite) TestAppendLine() {
	type args struct {
		line string
	}
	tests := []struct {
		args args
		cache *processCache
		want   *processCache
	}{
		{
			args: args {
				line: "this is a new line",
			},
			cache: &processCache{
				ioActions: []ioAction{
					{actionType: writeFile, lines: []string{"first line"}},
				},
				file: true,
			},
			want: &processCache{
				ioActions: []ioAction{
					{actionType: writeFile, lines: []string{"first line", "this is a new line"}},
				},
				file: true,
			},
		},
	}
	for _, tt := range tests {
		tt.cache.appendLine(tt.args.line)
		s.Assert().Equal(tt.want, tt.cache)
	}
}

func (s *CacheSuite) TestAppendFile() {
	type args struct {
		name string
		text string
	}
	tests := []struct {
		args args
		cache *processCache
		want   *processCache
	}{
		{
			args: args {
				name: "filename",
				text: "text for new file",
			},
			cache: &processCache{
				currentDirectory: []string{"output", "target", "xml-tag", "0"},
				ioActions: []ioAction{
					{actionType: writeFile, lines: []string{"xml file"}, ready: true},
				},
			},
			want: &processCache{
				currentDirectory: []string{"output", "target", "xml-tag", "0"},
				ioActions: []ioAction{
					{actionType: writeFile, lines: []string{"xml file"}, ready: true},
					{actionType: writeFile, lines: []string{xml.Header + "text for new file"}, ready: true, path: "output/target/xml-tag/0/filename.xml"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt.cache.appendFile(tt.args.name, tt.args.text)
		s.Assert().Equal(tt.want.currentDirectory, tt.cache.currentDirectory)
		s.Assert().ElementsMatch(tt.want.ioActions, tt.cache.ioActions)
	}
}