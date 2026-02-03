package main

import (
	"os"
	"testing"
)

func TestParser(t *testing.T) {

	tempFile, err := os.CreateTemp("", "test_parser_*")
	if err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(
		"enthuse config",
		`Source.*Date.*Type.*Payment ID.*Payment Amount`,
		"enthuse",
		[]string{"Date", "Payment ID", "Payment Amount", "Supporter ID", "First Name", "Last Name", "Source"},
		[]string{"File"},
		[]string{"{{ .Filename }}"},
		tempFile.Name(),
	)
	if err != nil {
		t.Fatal(err)
	}

	inFile := "testdata/enthuse_bad_first_sheet.xlsx"

	err = parser.Process(inFile)
	if err != nil {
		t.Errorf("processing error: %v", err)
	}

}
