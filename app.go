package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

// App brings all the elements of the application together.
type App struct {
	runStart time.Time
	log      *slog.Logger
}

// NewApp initialises a new App.
func NewApp(logTarget io.Writer) *App {
	if logTarget == nil {
		logTarget = os.Stdout
	}
	return &App{
		runStart: time.Now(),
		log:      slog.New(slog.NewTextHandler(logTarget, nil)),
	}
}

// Run meets the cli Applicator interface.
func (a *App) Run(yamlFile, converter, outputFile string, force bool, filePaths ...string) error {
	if yamlFile == "" {
		return errors.New("yaml file not provided")
	}
	if converter == "" {
		return errors.New("converter from yaml file not specified")
	}
	if outputFile == "" {
		return errors.New("outputFile file not provided")
	}
	if len(filePaths) == 0 {
		return errors.New("no files provided to process")
	}
	if s, err := os.Stat(outputFile); err == nil {
		if s.IsDir() {
			return fmt.Errorf("outputFile %q is a directory", outputFile)
		}
		if !force {
			return fmt.Errorf("outputFile file %q already exists", outputFile)
		}
	}

	// Read the configuration yaml.
	configBytes, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("could not read yaml file %q: %v", yamlFile, err)
	}
	config, err := newConfig(configBytes)
	if err != nil {
		return fmt.Errorf("could not parse yaml file %q: %v", yamlFile, err)
	}

	// Determine the converter to use.
	cvt := config.GetConverter(converter)
	if cvt == nil {
		return fmt.Errorf("converter %q could not be found in the yaml file", converter)
	}

	// Determine the files to process.
	filepaths, err := ParseFileList(filePaths...)
	if err != nil {
		return err
	}

	// Setup the parser.
	parser, err := NewParser(
		cvt.Name,
		cvt.headerMatcher,
		cvt.Columns,
		cvt.AdditionalColumns,
		cvt.AdditionalData,
		outputFile,
	)
	if err != nil {
		return fmt.Errorf("could not setup new parser: %v", err)
	}
	defer parser.Close()

	// Process each file, logging any errors.
	a.log.Info("--------------------------------")
	a.log.Info(fmt.Sprintf("Started processing: %s", a.runStart.Format(time.RFC3339)))
	a.log.Info("--------------------------------")

	var fileCount int
	var errCount int
	var totalRecordCount int

	for i, file := range *filepaths {
		if i > 0 {
			a.log.Info("--------------------------------")
		}
		a.log.Info(fmt.Sprintf("Processing file %02d: %s", i+1, file))
		fileCount++

		recordCount, err := parser.Process(file)
		if err != nil {
			a.log.Warn(fmt.Sprintf("  error: %v", err))
			errCount++
			a.log.Info("  process error")
		} else {
			totalRecordCount += recordCount
			a.log.Info(fmt.Sprintf("  %d records found", recordCount))
			a.log.Info("  processed ok")
		}
		parser.Flush()
	}

	// Flush and close the csv file.
	err = parser.Close()
	if err != nil {
		a.log.Warn(fmt.Sprintf("Output close/flush error: %v", err))
	}

	a.log.Info("--------------------------------")
	a.log.Info(fmt.Sprintf("Completed processing in %s", time.Now().Sub(a.runStart)))
	a.log.Info(fmt.Sprintf("File count %d error count %d", fileCount, errCount))
	a.log.Info(fmt.Sprintf("Total record count %d", totalRecordCount))
	a.log.Info(fmt.Sprintf("Output written to %q", outputFile))
	a.log.Info("--------------------------------")

	return nil

}
