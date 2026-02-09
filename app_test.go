package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestApp(t *testing.T) {

	var ephemeralTestDir = true

	// Setup test directory and log file.
	var testDir string
	var err error
	if ephemeralTestDir {
		testDir = t.TempDir()
	} else {
		testDir, err = os.MkdirTemp("/tmp/", "testapp_*")
		if err != nil {
			t.Fatal(err)
		}
	}
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
	if got, want := len(bytes.Split(contents, []byte("\n"))), 11; got != want {
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
	time=2026-02-09T11:43:13.830Z level=INFO msg=--------------------------------
	time=2026-02-09T11:43:13.830Z level=INFO msg="Completed processing in 7.434675ms"
	time=2026-02-09T11:43:13.830Z level=INFO msg="File count 5 error count 2"
	time=2026-02-09T11:43:13.830Z level=INFO msg="Total record count 9"
	time=2026-02-09T11:43:13.830Z level=INFO msg="Output written to \"/tmp/testapp_375922856/output.csv\""
	time=2026-02-09T11:43:13.830Z level=INFO msg=--------------------------------
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
