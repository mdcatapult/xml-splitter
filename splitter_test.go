package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type mockFile struct {
	data string
	done bool
}

func (m *mockFile) Read(bytes []byte) (int, error) {
	copy(bytes, m.data)
	if m.done {
		return 0, io.EOF
	}
	m.done = true
	return len(m.data), nil
}

func (m *mockFile) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

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

}
