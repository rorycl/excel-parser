package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseFiles(t *testing.T) {

	tests := []struct {
		name     string
		paths    []string
		expected []string
		Err      error
	}{
		{
			name: "2 files 1 ok directory",
			paths: []string{
				"testdata/parseFiles/okDir/",
				"testdata/parseFiles/1.xls",
				"testdata/parseFiles/3.xLsX",
			},
			expected: []string{
				"testdata/parseFiles/okDir/4.xls",
				"testdata/parseFiles/1.xls",
				"testdata/parseFiles/3.xLsX",
			},
		},
		{
			name: "2 files 1 not ok directory",
			paths: []string{
				"testdata/parseFiles/1.xls",
				"testdata/parseFiles/3.xLsX",
				"testdata/parseFiles/notOkDir/",
			},
			Err: fmt.Errorf("directory %q had no valid Excel files", "testdata/parseFiles/notOkDir/"),
		},
		{
			name: "3 files 1 invalid",
			paths: []string{
				"testdata/parseFiles/1.xls",
				"testdata/parseFiles/3.xLsX",
				"testdata/parseFiles/4.doc",
			},
			Err: fmt.Errorf("file %q does not have a valid Excel file extension", "testdata/parseFiles/4.doc"),
		},
	}

	for ii, tt := range tests {
		t.Run(fmt.Sprintf("%d_%s", ii, tt.name), func(t *testing.T) {

			paths, err := ParseFileList(tt.paths...)
			if err != nil {
				if tt.Err == nil {
					t.Fatal("unexepected error", err)
				}
				if got, want := err.Error(), tt.Err.Error(); got != want {
					t.Errorf("got %q want %q error string", got, want)
				}
				return
			}
			if err == nil && tt.Err != nil {
				t.Errorf("expected err %q", tt.Err.Error())
			}
			if diff := cmp.Diff(paths, tt.expected); diff != "" {
				t.Errorf("unexpected path results error:\n%s", diff)
			}
		})
	}
}
