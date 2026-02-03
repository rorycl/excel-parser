package main

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {

	configBytes, err := os.ReadFile("config_example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	config, err := newConfig(configBytes)
	if err != nil {
		t.Fatalf("new config error: %v", err)
	}

	if got, want := config.LogFile, "output.log"; got != want {
		t.Errorf("logfile got %s want %s", got, want)
	}

	if enthuse := config.GetConverter("enthuse"); enthuse == nil {
		t.Errorf("got nil converter for 'enthuse'")
	}

}
