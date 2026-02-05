package main

import (
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var testParserOutput = `
File,Date,Payment ID,Payment Amount,Supporter ID,First Name,Last Name,Source
testdata/enthuse_bad_first_sheet.xlsx,04/11/2025 00:23,py_abcdef2SZANN6oA613bvgL6ja,5,9157399,xxx,yyy,Fundraising & Donations
testdata/enthuse_bad_first_sheet.xlsx,04/11/2025 20:08,py_abcdef2Saion6OM60yICk10Qb,10,9157398,xxx,yyy,Fundraising & Donations
testdata/enthuse_bad_first_sheet.xlsx,05/11/2025 23:55,ch_abcdef2Sb8q26OA60KAgOHD7c,10,9848690,Anonymous,,Fundraising & Donations
`

func TestParser(t *testing.T) {

	testDir := t.TempDir()

	tempFile, err := os.CreateTemp(testDir, "test_parser_*")
	if err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(
		"enthuse config",
		regexp.MustCompile(`Source.*Date.*Type.*Payment ID.*Payment Amount`),
		[]string{"Date", "Payment ID", "Payment Amount", "Supporter ID", "First Name", "Last Name", "Source"},
		[]string{"File"},
		[]string{"{{ .Filename }}"},
		tempFile.Name(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Enable debug for debugging in parser.
	// parser.debug = true

	inFile := "testdata/enthuse_bad_first_sheet.xlsx"

	recordCount, err := parser.Process(inFile)
	if err != nil {
		t.Errorf("processing error: %v", err)
	}

	// Close and flush the parser.
	err = parser.Close()
	if err != nil {
		t.Fatal(err)
	}

	// check the number of records written.
	if got, want := recordCount, 3; got != want {
		t.Errorf("expected %d got %d records", got, want)
	}

	// check the output
	b, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("could not read file %q: err %v", tempFile.Name(), err)
	}
	if diff := cmp.Diff(string(b), testParserOutput[1:]); diff != "" {
		t.Errorf("got unexpected diff:\n%s", diff)
	}

	// Uncomment to see what was written to disk.
	// fmt.Println(string(b))

}
