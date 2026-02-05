package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/xuri/excelize/v2"
)

// Parser provides a way of checking if a sheet in an Excel file matches the expected
// headers and then provides a way of filtering the content for writing to the output.
// Parser is designed to be used across a set of associated files.
type Parser struct {
	Name                 string
	HeaderMatch          *regexp.Regexp
	Fields               []string
	MandatoryDataHeaders []string // headers to write at the beginning of the output
	MandatoryData        []string // data to write at the beginning of each output
	mandatoryDataTpls    []*template.Template
	positions            []int
	headersWritten       bool
	firstFileProcessed   bool
	writer               *csv.Writer
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
// spreadsheet using header offset positions (from 0). Only the first file in a set is
// used for the column position setting; the assumption is that all the files have the
// same structure.
func (p *Parser) setColumnPositions(headers []string) error {
	if p.firstFileProcessed {
		return nil
	}
	p.firstFileProcessed = true
	seen := map[string]bool{}

	for _, f := range p.Fields {
		for i, h := range headers {
			if h == f {
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

// compileTemplates compiles the mandatoryDataTpls from the MandatoryData, if any.
func (p *Parser) compileTemplates() error {
	if len(p.MandatoryData) < 1 {
		return nil
	}
	p.mandatoryDataTpls = make([]*template.Template, len(p.MandatoryData))
	var err error
	for i, md := range p.MandatoryData {
		p.mandatoryDataTpls[i], err = template.New(fmt.Sprintf("%d", i)).Parse(md)
		if err != nil {
			return fmt.Errorf("could not parse template %q", md)
		}
	}
	return nil
}

// ErrHeaderAlreadyWritten indicates the header has already been written. This is a soft
// error and should be caught by the user and the row skipped for writing.
var ErrHeaderAlreadyWritten error = errors.New("header already written")

// filterRow filters a row of Excel data based on the parser configuration recipe,
func (p *Parser) filterRow(row []string, header bool, data any) ([]string, error) {
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
		// Replace any templated
		var err error
		mdh := make([]string, len(p.mandatoryDataTpls))
		for i, tpl := range p.mandatoryDataTpls {
			var b bytes.Buffer
			err = tpl.Execute(&b, data)
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

// Flush flushes the csv data to disk.
func (p *Parser) Flush() {
	p.writer.Flush()
	// _ = p.writer.Close()
}

// log is a debugging logger
func (p *Parser) log(s string, a ...any) {
	if !p.debug {
		return
	}
	log.Printf(s, a...)
}

// Process writes filtered data from an Excel file to the writer, if found.
func (p *Parser) Process(fileName string) error {

	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return fmt.Errorf("open %q error: %v", fileName, err)
	}

	var found bool
	var sheets int
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
			return fmt.Errorf("could not get rows for %q, sheet %q (%d): %v", fileName, sheetName, idx, err)
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
			return fmt.Errorf("file %q sheet %q (%d) header row error: %v", fileName, sheetName, idx, err)
		}

		for i, row := range rows[headerRow:] {
			fRow, err := p.filterRow(row, i == 0, data)
			if err != nil {
				if errors.Is(err, ErrHeaderAlreadyWritten) { // header already written by previous file.
					continue
				}
				return fmt.Errorf("file %q sheet %q (%d) row %v filter error: %v", fileName, sheetName, idx, row, err)
			}
			err = p.writeRow(fRow)
			if err != nil {
				return fmt.Errorf("file %q sheet %q (%d) row %v write error: %v", fileName, sheetName, idx, row, err)
			}
		}
		break
	}

	if !found {
		return fmt.Errorf("no headers found in the %d sheets in %q", sheets, fileName)
	}

	return nil
}
