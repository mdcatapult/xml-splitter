package main

import (
	"reflect"
	"testing"
)

func TestXMLSplitter_GetLineStructure(t *testing.T) {
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
		s := XMLSplitter{}
		t.Run(tt.name, func(t *testing.T) {
			if got := s.GetLineStructure(tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLineStructure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tabs(t *testing.T) {
	type args struct {
		depth int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy path",
			args: args {
				depth: 1,
			},
			want: "  ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tabs(tt.args.depth); got != tt.want {
				t.Errorf("tabs() = %v, want %v", got, tt.want)
			}
		})
	}
}