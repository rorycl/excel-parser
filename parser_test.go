package main

import (
	"bytes"
	"os"
	"regexp"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
)

var testEnthuseParserOutput = `
File,Reference,Date,Payment ID,Payment Amount,Supporter ID,First Name,Last Name,Source
testdata/enthuse_bad_first_sheet.xlsx,ENTH-20251104,04/11/2025 00:23,py_abcdef2SZANN6oA613bvgL6ja,5,9157399,xxx,yyy,Fundraising & Donations
testdata/enthuse_bad_first_sheet.xlsx,ENTH-20251104,04/11/2025 20:08,py_abcdef2Saion6OM60yICk10Qb,10,9157398,xxx,yyy,Fundraising & Donations
testdata/enthuse_bad_first_sheet.xlsx,ENTH-20251105,05/11/2025 23:55,ch_abcdef2Sb8q26OA60KAgOHD7c,10,9848690,Anonymous,,Fundraising & Donations
`

var testJustGivingParserOutput = `
File,Reference,Donation Date,Donation Ref,Donation Amount,Donor User Id,Donor FirstName,Donor LastName,Event Name
testdata/justgiving_example.xlsx,JG-20251104,04/11/2025,989777889,10,990868212,Anonymous,Anonymous,Just Fundraising
testdata/justgiving_example.xlsx,JG-20251108,08/11/2025,989777890,10,990868223,Anonymous,Anonymous,Christmas Event
`

// runParser runs a parser for an input Excel file, with output to tmpFile. The results
// are compared to expectedRecords and expectedOutput.
func runParser(t *testing.T, inputExcelFile, tmpFile string, parser *Parser, expectedRecords int, expectedOutput string) {

	t.Helper()

	// Enable debug for debugging in parser.
	// parser.debug = true

	recordCount, err := parser.Process(inputExcelFile)
	if err != nil {
		t.Errorf("processing error: %v", err)
	}

	// Close and flush the parser.
	err = parser.Close()
	if err != nil {
		t.Fatal(err)
	}

	// check the number of records written.
	if got, want := recordCount, expectedRecords; got != want {
		t.Errorf("expected %d got %d records", got, want)
	}

	// check the output
	b, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("could not read file %q: err %v", tmpFile, err)
	}
	// slice off the leading "\n" of the expectedOutput.
	if diff := cmp.Diff(string(b), expectedOutput[1:]); diff != "" {
		t.Errorf("got unexpected diff:\n%s", diff)
	}

	// Uncomment to see what was written to disk.
	// fmt.Println(string(b))

}

func TestParserForEnthuse(t *testing.T) {
	testDir := t.TempDir()

	tempFile, err := os.CreateTemp(testDir, "test_parser_*")
	if err != nil {
		t.Fatal(err)
	}
	parser, err := NewParser(
		"enthuse config",
		regexp.MustCompile(`Source.*Date.*Type.*Payment ID.*Payment Amount`),
		[]string{"Date", "Payment ID", "Payment Amount", "Supporter ID", "First Name", "Last Name", "Source"},
		[]string{"File", "Reference"},
		[]string{"{{ .Filename }}", `ENTH-{{ .Date | TimestampParseAndFormat "02/01/2006 15:04" "20060102" }}`},
		tempFile.Name(),
	)
	if err != nil {
		t.Fatal(err)
	}
	runParser(t, "testdata/enthuse_bad_first_sheet.xlsx", tempFile.Name(), parser, 3, testEnthuseParserOutput)
}

func TestParserForJustGiving(t *testing.T) {
	testDir := t.TempDir()

	tempFile, err := os.CreateTemp(testDir, "test_parser_*")
	if err != nil {
		t.Fatal(err)
	}
	parser, err := NewParser(
		"just giving config",
		regexp.MustCompile(`Fundraiser User Id.*Fundraiser LastName.*Event Name.*Donor User Id.*Donation Ref.*Donation Date.*Donation Payment Reference.*Donation Amount`),
		[]string{"Donation Date", "Donation Ref", "Donation Amount", "Donor User Id", "Donor FirstName", "Donor LastName", "Event Name"},
		[]string{"File", "Reference"},
		[]string{"{{ .Filename }}", `JG-{{ .DonationDate | TimestampParseAndFormat "02/01/2006" "20060102" }}`},
		tempFile.Name(),
	)
	if err != nil {
		t.Fatal(err)
	}
	runParser(t, "testdata/justgiving_example.xlsx", tempFile.Name(), parser, 2, testJustGivingParserOutput)
}

// TestFieldNameForTemplate tests field name capitalisation for use as template keys.
func TestFieldNameForTemplate(t *testing.T) {
	if got, want := fieldNameForTemplate(" donation date "), "DonationDate"; got != want {
		t.Errorf("test field name got %q want %q", got, want)
	}
}

// TestTplFuncRegexpReplace tests the template function tplRegexpReplace.
func TestTplFuncRegexpReplace(t *testing.T) {
	tpl := template.New("test")
	tpl.Funcs(template.FuncMap{"RegexReplace": tplRegexpReplace})
	tpl, err := tpl.Parse(`Dialogue: {{ . | RegexReplace "^(hello).*" "${1} young chap" }}`)
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	err = tpl.Execute(&b, "hello old man")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b.String(), "Dialogue: hello young chap"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
