package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestApp(t *testing.T) {

	// Setup test directory and log file.
	testDir := t.TempDir()
	logFile, err := os.CreateTemp(testDir, "log_*.log")
	if err != nil {
		t.Fatal(err)
	}
	logFileName := logFile.Name()

	ctx := context.Background()

	// New output file invocation.
	invocation := []string{
		"excel-parser",
		"-y",
		"config_example.yaml",
		"-c",
		"enthuse",
		"-o",
		filepath.Join(testDir, "output.csv"),
		"testdata/",
	}
	app := NewApp(logFile)
	cmd := BuildCLI(app)
	err = cmd.Run(ctx, invocation)
	if err != nil {
		t.Fatal("first run error", err)
	}

	// Force overwrite output file invocation.
	invocation = []string{
		"excel-parser",
		"-y",
		"config_example.yaml",
		"-c",
		"enthuse",
		"-f",
		"-o",
		filepath.Join(testDir, "output.csv"),
		"testdata/",
	}
	app = NewApp(logFile)
	cmd = BuildCLI(app)
	err = cmd.Run(ctx, invocation)
	if err != nil {
		t.Fatal("second run error", err)
	}

	// Check the output file.
	contents, err := os.ReadFile(filepath.Join(testDir, "output.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bytes.Split(contents, []byte("\n"))), 10; got != want {
		t.Errorf("got %d lines want %d", got, want)
	}
	// fmt.Println(string(contents))

	// Check the log file.
	err = logFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	logContents, err := os.ReadFile(logFileName)
	if err != nil {
		t.Fatal(err)
	}
	/* output is something like the following
	time=2026-02-05T18:41:03.081Z level=INFO msg="Completed processing in 7.94035ms"
	time=2026-02-05T18:41:03.081Z level=INFO msg="File count 4 error count 1"
	time=2026-02-05T18:41:03.081Z level=INFO msg="Total record count 9"
	time=2026-02-05T18:41:03.081Z level=INFO msg="Output written to /tmp/TestApp3211766467/001/output.csv"
	*/
	wanted := [][]byte{
		[]byte("File count 4 error count 1"),
		[]byte("Total record count 9"),
	}
	for _, want := range wanted {
		if !bytes.Contains(logContents, want) {
			t.Errorf("could not find %q in log contents", string(want))
		}
	}

	// fmt.Println(string(logContents))

}
