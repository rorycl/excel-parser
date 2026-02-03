package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/xuri/excelize/v2"
)

type Config struct {
	HeaderMatch *regexp.Regexp
	Typer       string
	Fields      []string
	positions   []int
}

// setPositions sets the positions of each column required for reporting in a
// spreadsheet using header offset positions (from 0).
func (c *Config) setPositions(headers []string) error {
	seen := map[string]bool{}
	for i, h := range headers {
		for _, f := range fields {
			if h == f {
				if _, ok := seen[f]; ok {
					continue
				}
				c.positions = append(c.positions[i])
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
	for _, ix := range c.positions {
		filteredRow[i] = row[i]
	}
	return filteredRow
}

func main() {

	csvFile, err := os.Create("/tmp/writer.csv")
	if err != nil {
		fmt.Println("csv file open error", err)
		os.Exit(1)
	}
	csvWriter := csv.NewWriter(csvFile)

	f, err := excelize.OpenFile("testdata/Enthuse - 2025.12.15 - 199.15 - anonymised.xlsx")
	if err != nil {
		fmt.Println("open error", err)
		os.Exit(1)
	}
	defer func() {
		_ = f.Close()
		err := csvFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	for idx, name := range f.GetSheetMap() {
		fmt.Println("sheet", idx, name)
		rows, err := f.GetRows(name)
		if err != nil {
			fmt.Println("get rows", err)
			os.Exit(1)
		}
		for _, r := range rows {
			fmt.Println(r)
			csvWriter.Write(r)
		}
	}
	csvWriter.Flush()

}
