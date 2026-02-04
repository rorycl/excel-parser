package main

import (
	"context"
	"fmt"
	"testing"
)

// testApplication meets the Apper interface.
type testApplication struct {
	savedOutput string
}

// Run fulfils the Apper interface used for injecting an app into the CLI.
func (ta *testApplication) Run(yaml, converter, output string, force bool, args ...string) error {
	ta.savedOutput = fmt.Sprintf("yaml %q converter %q output file %q force %t : %v",
		yaml, converter, output, force, args,
	)
	return nil
}

func TestCLI(t *testing.T) {

	testApp := &testApplication{}

	cmd := BuildCLI(testApp)
	err := cmd.Run(context.Background(),
		[]string{"program", "-y", "config_example.yaml", "-c", "enthuse", "-o", "abc.csv", "input.xlsx"},
	)
	if err != nil {
		t.Errorf("parsing error: %v", err)
	}

	expected := `yaml "config_example.yaml" converter "enthuse" output file "abc.csv" force false : [input.xlsx]`

	if got, want := testApp.savedOutput, expected; got != want {
		t.Errorf("got : %s\nwant: %s\n", got, want)
	}

}
