package main

import (
	"reflect"
	"testing"
)

//func TestGetConfig(t *testing.T) {
//	tests := []struct {
//		name    string
//		want    Config
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := GetConfig()
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetConfig() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestGetLineStructure(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLineStructure(tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLineStructure() = %v, want %v", got, tt.want)
			}
		})
	}
}
//
//func TestWrite(t *testing.T) {
//	type args struct {
//		data string
//		name string
//		path []string
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//		})
//	}
//}
//
//func TestXMLSplitter_GetScanner(t *testing.T) {
//	type fields struct {
//		path string
//		conf Config
//	}
//	type args struct {
//		target string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *bufio.Scanner
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &XMLSplitter{
//				path: tt.fields.path,
//				conf: tt.fields.conf,
//			}
//			got, err := s.GetScanner(tt.args.target)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetScanner() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetScanner() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestXMLSplitter_ProcessFile(t *testing.T) {
//	type fields struct {
//		path string
//		conf Config
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   int
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &XMLSplitter{
//				path: tt.fields.path,
//				conf: tt.fields.conf,
//			}
//			if got := s.ProcessFile(); got != tt.want {
//				t.Errorf("ProcessFile() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_handleError(t *testing.T) {
//	type args struct {
//		err error
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//		})
//	}
//}