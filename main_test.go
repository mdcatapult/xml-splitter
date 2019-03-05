package main

import (
	"fmt"
	"testing"
)

func TestAddLinesWithTrailingRegex(t *testing.T) {

	config.split = "</Entry>"
	var input = "<ArbitraryNode></ArbitraryNode></Entry><Entry>"

	s := XMLSplitter{}
	result := s.GetLines(input)

	if len(result) != 3 {
		t.Error(fmt.Sprintf("len(result) = %d, expected 3 from %s", len(result), result))
	} else {
		if result[1] != "" {
			t.Error(fmt.Sprintf("expected empty string at position 1: recieved %s", result[1]))
		}
		if result[0] != "<ArbitraryNode></ArbitraryNode></Entry>" {
			t.Error(fmt.Sprintf("%s != <ArbitraryNode></ArbitraryNode></Entry>", result[1]))
		}

		if result[2] != "<Entry>" {
			t.Error(fmt.Sprintf("%s != <Entry>", result[2]))
		}
	}

}

func TestAddLinesWithTrailingRegexAndMultipleDocsInOneLine(t *testing.T) {

	config.split = "</Entry>"
	var input = "<Id>1</Id></Entry><Entry><Id>2</Id></Entry><Entry><Id>3</Id></Entry><Entry>"

	s := XMLSplitter{}
	result := s.GetLines(input)

	if len(result) != 7 {
		t.Error(fmt.Sprintf("len(result) = %d, expected 7 from %s", len(result), result))
	} else {
		if result[1] != "" || result[3] != "" || result[5] != "" {
			t.Error(fmt.Sprintf("expected empty string at positions 1, 3 & 5: recieved %s", result[2]))
		}

		if result[0] != "<Id>1</Id></Entry>" {
			t.Error(fmt.Sprintf("recieved %s, expected <Id>1</Id></Entry>", result[0]))
		}
		if result[2] != "<Entry><Id>2</Id></Entry>" {
			t.Error(fmt.Sprintf("recieved %s, expected  <Entry><Id>2</Id></Entry>", result[1]))
		}
		if result[4] != "<Entry><Id>3</Id></Entry>" {
			t.Error(fmt.Sprintf("recieved %s, expected  <Entry><Id>3</Id></Entry>", result[2]))
		}
		if result[6] != "<Entry>" {
			t.Error(fmt.Sprintf("recieved %s, expected  <Entry><Id>3</Id></Entry>", result[2]))
		}
	}
}

//type MockSplitter struct {
//	*XMLSplitter
//}
//
//func (s *MockSplitter) GetScanner(target string) (*bufio.Scanner, error) {
//	fmt.Println("MOCK SCANNER")
//	input := "<Id>1</Id></Entry><Entry><Id>2</Id></Entry><Entry><Id>3</Id></Entry><Entry>"
//	r := strings.NewReader(input)
//	return bufio.NewScanner(r), nil
//}
//
//func TestProcessFileWithMultipleDocsInOneLine(t *testing.T) {
//	config.split = "</Entry>"
//	s := XMLSplitter{path: "/a/dummy/path/to/file.xml"}
//	filesCreated := s.ProcessFile()
//	fmt.Println(filesCreated)
//}
