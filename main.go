package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {

	parser, err := NewParser(
		"enthuse config",
		regexp.MustCompile(`Source.*Date.*Type.*Payment ID.*Payment Amount`),
		[]string{"Date", "Payment ID", "Payment Amount", "Supporter ID", "First Name", "Last Name", "Source"},
		[]string{"File"},
		[]string{"{{ .Filename }}"},
		"/tmp/enthuse.csv",
	)
	if err != nil {
		fmt.Println("enthuse parser error", err)
		os.Exit(1)
	}

	err = parser.Process("testdata/Enthuse - 2025.12.15 - 199.15 - anonymised.xlsx")
	if err != nil {
		fmt.Printf("enthuse parsing error: %v\n", err)
		os.Exit(1)
	}
	parser.Flush()

	parser, err = NewParser(
		"just giving config",
		regexp.MustCompile(`Fundraiser User Id.*Fundraiser LastName.*Event Name.*Donor User Id.*Donation Ref.*Donation Date.*Donation Payment Reference.*Donation Amount`),
		[]string{"Donation Date", "Donation Ref", "Donation Amount", "Donor User Id", "Donor FirstName", "Donor LastName", "Event Name"},
		[]string{"File"},
		[]string{"testdata/JustGiving 2025.12.15- £5105.03 anonymised.xlsx"},
		"/tmp/justgiving.csv",
	)
	if err != nil {
		fmt.Println("just giving parser error", err)
		os.Exit(1)
	}

	err = parser.Process("testdata/JustGiving 2025.12.15- £5105.03 anonymised.xlsx")
	if err != nil {
		fmt.Printf("parsing error: %v\n", err)
		os.Exit(1)
	}
	parser.Flush()

}
