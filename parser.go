package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/xuri/excelize/v2"
)

// Parser provides a way of checking if a sheet in an Excel file matches the expected
// headers and then provides a way of filtering the content for writing to the output.
// Parser is designed to be used across a set of associated files.
type Parser struct {
	Name                 string
	HeaderMatch          *regexp.Regexp
	originalHeaders      []string
	Fields               []string
	positions            []int    // position of the Fields in the originalHeaders
	MandatoryDataHeaders []string // headers to write at the beginning of the output
	MandatoryData        []string // data to write at the beginning of each output
	mandatoryDataTpls    []*template.Template
	headersWritten       bool
	writer               *csv.Writer
	fileCloser           io.Closer
	debug                bool
}

func NewParser(name string, headerMatch *regexp.Regexp, fields, mdh, md []string, outputFile string) (*Parser, error) {
	if name == "" {
		return nil, errors.New("newparser received an empty name argument")
	}

	if outputFile == "" {
		return nil, errors.New("newparser received an empty output file argument")
	}

	if got, want := len(mdh), len(md); got != want {
		return nil, fmt.Errorf("mandatory %d headers don't match %d data fields", got, want)
	}

	csvFile, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("could not open csv file %q for writing: %w", outputFile, err)
	}
	csvWriter := csv.NewWriter(csvFile)

	p := &Parser{
		Name:                 name,
		HeaderMatch:          headerMatch,
		Fields:               fields,
		MandatoryDataHeaders: mdh,
		MandatoryData:        md,
		writer:               csvWriter,
		fileCloser:           csvFile,
	}

	err = p.compileTemplates()
	if err != nil {
		return nil, fmt.Errorf("template compilation error: %w", err)
	}

	return p, nil
}

// searchSheetForHeader determines if the required header can be found in the provided
// sheet [][]string, returning -1 if no match was found, or the row number starting at 0
// if it was. Each row's columns are joined by a space prior to match checking.
func (p *Parser) searchSheetForHeader(allRows [][]string) int {
	if allRows == nil {
		return -1
	}
	for i, ar := range allRows {
		if p.HeaderMatch.MatchString(strings.Join(ar, " ")) {
			return i
		}
	}
	return -1
}

// setColumnPositions sets the positions of each column required for reporting in a
// spreadsheet using header offset positions (from 0). Each file regenerates the header
// positions to allow for files that have had temporary working columns inserted.
// To get to this point in the program, the Excel sheet in question has to have passed
// the HeaderMatch matcher.
func (p *Parser) setColumnPositions(headers []string) error {

	// Reset the original header.
	p.originalHeaders = headers

	// Reset positions.
	p.positions = []int{}

	seen := map[string]bool{}
	for _, f := range p.Fields {
		for i, h := range headers {
			if h == f {
				// Ignore any duplicate headers.
				if _, ok := seen[f]; ok {
					continue
				}
				p.positions = append(p.positions, i)
				seen[h] = true
			}
		}
	}
	if got, want := len(p.positions), len(p.Fields); got != want {
		return fmt.Errorf("got %d positions for %d fields", got, want)
	}
	return nil
}

// funcMapper prepares a template.FuncMap
func (p *Parser) funcMapper() template.FuncMap {
	return template.FuncMap{
		"RegexpReplace":           tplRegexpReplace,
		"TimestampParseAndFormat": tplTimestampParseAndFormat,
	}
}

// compileTemplates compiles the mandatoryDataTpls from the MandatoryData, if any.
func (p *Parser) compileTemplates() error {
	if len(p.MandatoryData) < 1 {
		return nil
	}
	p.mandatoryDataTpls = make([]*template.Template, len(p.MandatoryData))
	var err error
	for i, md := range p.MandatoryData {
		tpl := template.New(fmt.Sprintf("%d", i))
		tpl.Funcs(p.funcMapper()) // add the funcMap
		p.mandatoryDataTpls[i], err = tpl.Parse(md)
		if err != nil {
			return fmt.Errorf("could not parse template %q", md)
		}
	}
	return nil
}

// ErrHeaderAlreadyWritten indicates the header has already been written. This is a soft
// error and should be caught by the user and the row skipped for writing.
var ErrHeaderAlreadyWritten error = errors.New("header already written")

// filterRow filters a row of Excel data based on the parser configuration recipe, Note
// that data on a per-sheet basis is provided and then added into if templated
// "mandatory data" is provided to add each cell of data named by the column header.
func (p *Parser) filterRow(row []string, header bool, data map[string]any) ([]string, error) {
	if len(p.positions) < 1 {
		return nil, errors.New("filterRow called without positions being initialised")
	}
	if got, want := p.positions[len(p.positions)-1], len(row)-1; got > want {
		return nil, fmt.Errorf("last position %d outside of bounds of row %d", got, want)
	}
	if header {
		if p.headersWritten {
			return nil, ErrHeaderAlreadyWritten
		}
		p.headersWritten = true
	}
	filteredRow := make([]string, len(p.Fields))
	for i, ix := range p.positions {
		filteredRow[i] = row[ix]
	}
	if len(p.MandatoryData) > 0 {
		if header {
			return slices.Concat(p.MandatoryDataHeaders, filteredRow), nil
		}
		// Replace any templated fields after first copying and then filling the
		// template data with data from the input rows using the original headers and
		// all row data as applicable.
		// Important note: data fields need to be exported (i.e. capitalised) and have
		// any spaces removed -- as a crude requirement for a template key.
		localData := data
		for i, originalHeaderName := range p.originalHeaders {
			field := fieldNameForTemplate(originalHeaderName)
			var cell string
			if i > len(row)-1 {
				cell = ""
			} else {
				cell = row[i]
			}
			localData[field] = cell
		}
		var err error
		mdh := make([]string, len(p.mandatoryDataTpls))
		for i, tpl := range p.mandatoryDataTpls {
			var b bytes.Buffer
			err = tpl.Execute(&b, localData)
			if err != nil {
				return nil, fmt.Errorf("template execution error, offset %d, err: %v", i, err)
			}
			mdh[i] = b.String()
		}
		return slices.Concat(mdh, filteredRow), nil
	}
	return filteredRow, nil
}

// writeRow writes a csv row.
func (p *Parser) writeRow(row []string) error {
	return p.writer.Write(row)
}

// Flush flushes the csv data to disk and closes the underlying file.
func (p *Parser) Flush() {
	p.writer.Flush()
}

// Close flushes the csv and closes the underlying file.
func (p *Parser) Close() error {
	p.Flush()
	return p.fileCloser.Close()
}

// log is a debugging logger
func (p *Parser) log(s string, a ...any) {
	if !p.debug {
		return
	}
	log.Printf(s, a...)
}

// Process writes filtered data from an Excel file to the writer, if found. The method
// returns the number of records and an error, if any.
func (p *Parser) Process(fileName string) (int, error) {

	var found bool
	var sheets int
	var recordCount int

	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return recordCount, fmt.Errorf("open %q error: %v", fileName, err)
	}

	// data is for providing to templates.
	data := map[string]any{
		"Filename": fileName,
	}

	// Iterate over all the sheets in the spreadsheet, searching for the header, then
	// mapping the field names/column names to column positions, then writing out the
	// wanted data, if needed.
	for idx, sheetName := range f.GetSheetMap() {

		p.log("sheet: %d : %s", idx, sheetName)
		sheets++

		rows, err := f.GetRows(sheetName)
		if err != nil {
			return recordCount, fmt.Errorf("could not get rows for %q, sheet %q (%d): %v", fileName, sheetName, idx, err)
		}
		if len(rows) < 1 {
			continue
		}
		headerRow := p.searchSheetForHeader(rows)
		if headerRow == -1 {
			continue
		}
		found = true
		p.log("found header on %s", sheetName)

		err = p.setColumnPositions(rows[headerRow])
		if err != nil {
			return recordCount, fmt.Errorf("file %q sheet %q (%d) header row error: %v", fileName, sheetName, idx, err)
		}

		for i, row := range rows[headerRow:] {
			fRow, err := p.filterRow(row, i == 0, data)
			if err != nil {
				if errors.Is(err, ErrHeaderAlreadyWritten) { // header already written by previous file.
					continue
				}
				return recordCount, fmt.Errorf("file %q sheet %q (%d) row %v filter error: %v", fileName, sheetName, idx, row, err)
			}
			err = p.writeRow(fRow)
			if err != nil {
				return recordCount, fmt.Errorf("file %q sheet %q (%d) row %v write error: %v", fileName, sheetName, idx, row, err)
			}
			// Ignore the header for record counting.
			if i != 0 {
				recordCount++
			}
		}
		break
	}

	if !found {
		return recordCount, fmt.Errorf("no headers found in the %d sheets in %q", sheets, fileName)
	}

	return recordCount, nil
}

// fieldNameForTemplate removes spaces from a field and makes the first char of a string
// a capital.
func fieldNameForTemplate(s string) string {
	if len(s) < 1 {
		return s
	}
	s = strings.TrimSpace(s)
	s = strings.Title(s)
	return strings.ReplaceAll(s, " ", "")
}

// tplRegexpReplace is a Go text/template template function.
// Example invocation:
//
//	"<please replace me> | RegexpReplace "<(.*replace.*)>" "fixed"
//	output: "<fixed>".
func tplRegexpReplace(pattern, replacement, input string) string {
	rgp, err := regexp.Compile(pattern)
	if err != nil {
		// panic with contextual info rather than use MustCompile
		panic(fmt.Sprintf("tplRegexReplace regexp compile error for %q: %v", pattern, err))
	}
	z := rgp.ReplaceAllString(input, replacement)
	return z
}

// tplTimestampParseAndFormat is a Go text/template template function.
// Example invocation:
//
//	"2026-09-01 09:01" | TimestampParseAndFormat "2006-01-02 15:04" "02/01/2006"
//	output: "01/09/2026"
//
// The func returns the original input string if it was not possible to parse.
func tplTimestampParseAndFormat(parseLayout, outputFormat, input string) string {
	parsedTime, err := time.Parse(parseLayout, input)
	if err != nil {
		return input
	}
	return parsedTime.Format(outputFormat)
}
