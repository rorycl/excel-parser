package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Config struct {
	Name        string
	HeaderMatch *regexp.Regexp
	Typer       string
	Fields      []string
	positions   []int
}

// searchSheetForHeader determines if the required header can be found in the provided
// sheet [][]string, returning -1 if no match was found, or the row number starting at 0
// if it was. Each row's columns are joined by a space prior to match checking.
func (c *Config) searchSheetForHeader(allRows [][]string) int {
	for i, ar := range allRows {
		if c.HeaderMatch.MatchString(strings.Join(ar, " ")) {
			return i
		}
	}
	return -1
}

// setPositions sets the positions of each column required for reporting in a
// spreadsheet using header offset positions (from 0).
func (c *Config) setPositions(headers []string) error {
	seen := map[string]bool{}
	for _, f := range c.Fields {
		for i, h := range headers {
			if h == f {
				if _, ok := seen[f]; ok {
					continue
				}
				c.positions = append(c.positions, i)
				seen[h] = true
			}
		}
	}
	if got, want := len(c.positions), len(c.Fields); got != want {
		return fmt.Errorf("got %d positions for %d fields", got, want)
	}
	return nil
}

// filterRow filters a row of Excel data based on the config recipe.
func (c *Config) filterRow(row []string) ([]string, error) {
	if len(c.positions) < 1 {
		return nil, errors.New("filterRow called without positions being initialised")
	}
	if got, want := c.positions[len(c.positions)-1], len(row)-1; got > want {
		return nil, fmt.Errorf("last position %d outside of bounds of row %d", got, want)
	}
	filteredRow := make([]string, len(c.Fields))
	for i, ix := range c.positions {
		filteredRow[i] = row[ix]
	}
	return filteredRow, nil
}

// parseExcelFile parses an excel file to find suitable data to output to a csv file.
func parseExcelFile(config *Config, inFile, outFile string) error {

	csvFile, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("could not open csv file for writing: %v", err)
	}
	csvWriter := csv.NewWriter(csvFile)
	defer func() {
		_ = csvFile.Close()
	}()

	f, err := excelize.OpenFile(inFile)
	if err != nil {
		return fmt.Errorf("open %q error: %v", inFile, err)
	}

	var found bool
	var sheets int

	// Iterate over all the sheets in the spreadsheet, searching for the header, then
	// mapping the field names/column names to column positions, then writing out the
	// wanted
	for idx, sheetName := range f.GetSheetMap() {
		sheets++
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return fmt.Errorf("could not get rows for %q, sheet %q (%d): %v", inFile, sheetName, idx, err)
		}
		headerRow := config.searchSheetForHeader(rows)
		if headerRow == -1 {
			continue
		}
		found = true

		err = config.setPositions(rows[headerRow])
		if err != nil {
			return fmt.Errorf("file %q sheet %q (%d) header row error: %v", inFile, sheetName, idx, err)
		}

		for _, row := range rows[headerRow:] {
			fRow, err := config.filterRow(row)
			if err != nil {
				return fmt.Errorf("file %q sheet %q (%d) row %v filter error: %v", inFile, sheetName, idx, row, err)
			}
			err = csvWriter.Write(fRow)
			if err != nil {
				return fmt.Errorf("file %q sheet %q (%d) row %v write error: %v", inFile, sheetName, idx, row, err)
			}
		}
		csvWriter.Flush()
		break
	}

	if !found {
		return fmt.Errorf("no headers found in the %d sheets in %q", sheets, inFile)
	}

	return nil
}

func main() {

	config := &Config{
		Name:        "enthuse config",
		HeaderMatch: regexp.MustCompile(`Source.*Date.*Type.*Payment ID.*Payment Amount`),
		Fields:      []string{"Date", "Payment ID", "Payment Amount", "Supporter ID", "First Name", "Last Name"},
	}

	err := parseExcelFile(
		config,
		"testdata/Enthuse - 2025.12.15 - 199.15 - anonymised.xlsx",
		"/tmp/enthuse.csv",
	)
	if err != nil {
		fmt.Println("parsing error: %v\n", err)
	}

}
