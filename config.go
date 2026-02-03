package main

// This file sets all the yaml configuration items.

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/goccy/go-yaml"
)

// ErrInvalidConfig reports an invalid yaml configuration file, although
// one that passed parsing.
type ErrInvalidConfig struct {
	info string
}

// Error reports the error.
func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid config : %s", e.info)
}

// A Converter is a particular type of Excel file converter, for example for Enthuse or
// Just Giving.
type Converter struct {
	Name              string `yaml:"name"`
	HeaderMatcher     string `yaml:"headerMatcher"`
	headerMatcher     *regexp.Regexp
	Columns           []string `yaml:"columns"`
	AdditionalColumns []string `yaml:"additionalColumns"`
	AdditionalData    []string `yaml:"additionalData"`
}

// Config represents the structure of a whole yaml file.
type Config struct {
	LogFile    string      `yaml:"logFile"`
	Converters []Converter `yaml:"converters"`
}

// GetConverter retrieves the named converter. If no converter is found, a nil is
// returned.
func (c *Config) GetConverter(name string) *Converter {
	for _, converter := range c.Converters {
		if converter.Name == name {
			return &converter
		}
	}
	return nil
}

// validateConfig validates the configuration.
func (c *Config) validateConfig() error {
	var err error
	if c.LogFile == "" {
		return errors.New("a logfile has not been provided")
	}
	if len(c.Converters) == 0 {
		return errors.New("no valid converters were found")
	}
	for _, conv := range c.Converters {
		if conv.Name == "" {
			return fmt.Errorf("%q name is empty", conv.Name)
		}
		if conv.HeaderMatcher == "" {
			return fmt.Errorf("%q header matcher is empty", conv.Name)
		}
		if conv.headerMatcher, err = regexp.Compile(conv.HeaderMatcher); err != nil {
			return fmt.Errorf("%q header matcher regexp error: %v", conv.Name, err)
		}
		if len(conv.Columns) < 1 {
			return fmt.Errorf("%q columns are empty", conv.Name)
		}
		if len(conv.AdditionalColumns) != len(conv.AdditionalData) {
			return fmt.Errorf("%q additional column and data lengths don't match", conv.Name)
		}
	}
	return nil
}

// newConfig creates and validates a new config from reading a yaml
// file.
func newConfig(b []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	err := c.validateConfig()
	return &c, err
}
