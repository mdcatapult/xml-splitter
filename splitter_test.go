package main

import (
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type SplitterSuite struct {
	suite.Suite
	splitter XMLSplitter
	file os.File
}

func TestSplitterSuite(t *testing.T) {
	suite.Run(t, new(SplitterSuite))
}

func (s *SplitterSuite) TestGetLineStructure(t *testing.T) {
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
					Type: Opening,
					Name: "Opening",
					Full: "<Opening>",
					Start: 0,
					End: 9,
				},
				9: {
					Type: Empty,
					Name: "Empty",
					Full: "<Empty />",
					Start: 9,
					End: 18,
				},
				50: {
					Type: Opening,
					Name: "OpeningWithAttributes",
					Full: `<OpeningWithAttributes time="1 o'clock" date="Wednesday 6th November">`,
					Start: 50,
					End: 120,
				},
				120: {
					Type: Closing,
					Name: "Closing",
					Full: "</ Closing>",
					Start: 120,
					End: 131,
				},
			},
		},
	}
	for _, tt := range tests {
		s.Assert().Equal(tt.want, s.splitter.GetLineStructure(tt.args.line))
	}
}