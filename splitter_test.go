package main

import (
	"bufio"
	"encoding/xml"
	"github.com/stretchr/testify/mock"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SplitterSuite struct {
	suite.Suite
	splitter XMLSplitter
	file     os.File
}

func TestSplitterSuite(t *testing.T) {
	suite.Run(t, new(SplitterSuite))
}

func (s *SplitterSuite) TestGetLineStructure() {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want map[int]Tag
	}{
		{
			name: "Happy path",
			args: args{
				line: `<Opening><Empty />Some text in here for posterity.<OpeningWithAttributes time="1 o'clock" date="Wednesday 6th November"></ Closing>`,
			},
			want: map[int]Tag{
				0: {
					Type:  Opening,
					Name:  "Opening",
					Full:  "<Opening>",
					Start: 0,
					End:   9,
				},
				9: {
					Type:  Empty,
					Name:  "Empty",
					Full:  "<Empty />",
					Start: 9,
					End:   18,
				},
				50: {
					Type:  Opening,
					Name:  "OpeningWithAttributes",
					Full:  `<OpeningWithAttributes time="1 o'clock" date="Wednesday 6th November">`,
					Start: 50,
					End:   120,
				},
				120: {
					Type:  Closing,
					Name:  "Closing",
					Full:  "</ Closing>",
					Start: 120,
					End:   131,
				},
			},
		},
	}
	for i, tt := range tests {
		s.T().Logf("Test get line structure: %s, case %d", tt.name, i)
		s.Assert().Equal(tt.want, s.splitter.getLineStructure(tt.args.line))
	}
}

func (s *SplitterSuite) TestProcessFile() {
	reader := &mockReader{data: `<uniprot xmlns="http://uniprot.org/uniprot"
 xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://uniprot.org/uniprot http://www.uniprot.org/docs/uniprot.xsd">
  <entry>  <accession>Q6GZX4</accession>  <name>001R_FRG3G</name>  <protein>    <recommendedName>      <fullName>Putative transcription factor 001R</fullName>    </recommendedName>  </protein></entry>
  <entry>  <accession>Q6GZX4</accession>  <name>001R_FRG3G</name>  <protein>    <recommendedName>      <fullName>Putative transcription factor 001R</fullName>    </recommendedName>  </protein></entry>
</uniprot>`}
	reader.On("Read", mock.Anything)
	config := Config{
		out: "out",
		skip: regexp.MustCompile(defaultSkip),
		strip: regexp.MustCompile(""),
		depth: 1,
	}
	splitter := XMLSplitter{path: "sprot", conf: config}
	writer := &mockWriter{}
	writer.On("write", []ioAction{
		{
			actionType: newDirectory,
			path: "out/sprot/uniprot/0",
			ready: true,
		},
		{
			actionType: writeFile,
			path: "out/sprot/uniprot/0/root.xml",
			lines: []string{xml.Header + `<uniprot xmlns="http://uniprot.org/uniprot"  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"  xsi:schemaLocation="http://uniprot.org/uniprot http://www.uniprot.org/docs/uniprot.xsd">`},
			ready: true,
		},
		{
			actionType: writeFile,
			path: "out/sprot/uniprot/0/entry.0.xml",
			lines: []string{
				xml.Header,
				"<entry>",
				"<accession>",
				"Q6GZX4",
				"</accession>",
				"<name>",
				"001R_FRG3G",
				"</name>",
				"<protein>",
				"<recommendedName>",
				"<fullName>",
				"Putative transcription factor 001R",
				"</fullName>",
				"</recommendedName>",
				"</protein>",
				"</entry>",
			},
			ready: true,
		},
		{
			actionType: writeFile,
			path: "out/sprot/uniprot/0/entry.1.xml",
			lines: []string{
				xml.Header,
				"<entry>",
				"<accession>",
				"Q6GZX4",
				"</accession>",
				"<name>",
				"001R_FRG3G",
				"</name>",
				"<protein>",
				"<recommendedName>",
				"<fullName>",
				"Putative transcription factor 001R",
				"</fullName>",
				"</recommendedName>",
				"</protein>",
				"</entry>",
			},
			ready: true,
		},
	}).Return([]ioAction{}, nil)
	totalFiles := splitter.ProcessFile(bufio.NewScanner(reader), writer)

	s.Assert().Equal(3, totalFiles)
	reader.AssertNumberOfCalls(s.T(), "Read", 2)
	writer.AssertNumberOfCalls(s.T(), "write", 1)
}
