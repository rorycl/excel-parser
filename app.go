package main

import (
	"errors"
	"fmt"
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
func NewApp() *App {
	return &App{
		runStart: time.Now(),
		log:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
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

	// Process each file, logging any errors.
	a.log.Info("--------------------------------")
	a.log.Info(fmt.Sprintf("Started processing: %s", a.runStart.Format(time.RFC3339)))
	a.log.Info("--------------------------------")

	var count int
	var errCount int
	for i, file := range *filepaths {
		if i > 0 {
			a.log.Info("--------------------------------")
		}
		a.log.Info(fmt.Sprintf("Processing file %02d: %s", i+1, file))
		count++

		err := parser.Process(file)
		if err != nil {
			a.log.Warn(fmt.Sprintf("  error: %v", err))
			errCount++
		}
		a.log.Info("  processed ok")
		parser.Flush()
	}
	a.log.Info("--------------------------------")
	a.log.Info(fmt.Sprintf("Completed processing in %s", time.Now().Sub(a.runStart)))
	a.log.Info(fmt.Sprintf("File count %d error count %d", count, errCount))
	a.log.Info("--------------------------------")

	return nil

}
