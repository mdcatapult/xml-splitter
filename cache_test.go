package main

import "testing"

func Test_processCache_newDirectory(t *testing.T) {
	type fields struct {
		depth            int
		currentDirectory []string
		directoryCounter map[string]int
		fileCounter      map[string]int
		totalFiles       int
		innerText        string
	}
	type args struct {
		tag Tag
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "XML root",
			fields: fields{
				currentDirectory: []string{"output-folder", "target-folder"},
				directoryCounter: make(map[string]int),
			},
			args: args{
				tag: Tag{Name: "xml-tag"},
			},
			want: "output-folder/target-folder/xml-tag/0",
		},
		{
			name: "repeated tag",
			fields: fields{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				directoryCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				tag: Tag{Name: "repeated-tag"},
			},
			want: "output-folder/target-folder/xml-tag/0/repeated-tag/2",
		},
		{
			name: "new-tag",
			fields: fields{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				directoryCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				tag: Tag{Name: "new-tag"},
			},
			want: "output-folder/target-folder/xml-tag/0/new-tag/0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &processCache{
				depth:            tt.fields.depth,
				currentDirectory: tt.fields.currentDirectory,
				directoryCounter: tt.fields.directoryCounter,
				fileCounter:      tt.fields.fileCounter,
				totalFiles:       tt.fields.totalFiles,
				innerText:        tt.fields.innerText,
			}
			if got := p.newDirectory(tt.args.tag); got != tt.want {
				t.Errorf("newDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processCache_newFile(t *testing.T) {
	type fields struct {
		depth            int
		currentDirectory []string
		directoryCounter map[string]int
		fileCounter      map[string]int
		totalFiles       int
		innerText        string
	}
	type args struct {
		tag Tag
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "repeated tag",
			fields: fields{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				fileCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				tag: Tag{Name: "repeated-tag"},
			},
			want: "output-folder/target-folder/xml-tag/0/repeated-tag.2.xml",
		},
		{
			name: "new-tag",
			fields: fields{
				currentDirectory: []string{"output-folder", "target-folder", "xml-tag", "0"},
				fileCounter: map[string]int{
					"output-folder/target-folder/xml-tag/0/repeated-tag": 1,
				},
			},
			args: args{
				tag: Tag{Name: "new-tag"},
			},
			want: "output-folder/target-folder/xml-tag/0/new-tag.0.xml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &processCache{
				depth:            tt.fields.depth,
				currentDirectory: tt.fields.currentDirectory,
				directoryCounter: tt.fields.directoryCounter,
				fileCounter:      tt.fields.fileCounter,
				totalFiles:       tt.fields.totalFiles,
				innerText:        tt.fields.innerText,
			}
			if got := p.openFile(tt.args.tag); got != tt.want {
				t.Errorf("openFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
