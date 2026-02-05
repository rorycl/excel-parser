package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	ShortUsage      = "Process excel files into a summary csv file."
	LongDescription = `This cli program uses a configuration yaml file to load all of
the Excel files listed on the command line and/or in the specified
directories into the specified csv file.`
)

// Applicator provides the general interface to the program.
type Applicator interface {
	Run(yaml, converter, output string, force bool, args ...string) error
}

// func BuildCLI(app Applicator) *cli.Command {
func BuildCLI(app Applicator) *cli.Command {

	yamlFlag := &cli.StringFlag{
		Name:    "yaml",
		Aliases: []string{"y"},
		Usage:   "configuration yaml file",
	}
	converterFlag := &cli.StringFlag{
		Name:    "converter",
		Aliases: []string{"c"},
		Usage:   "converter choice from yaml",
	}
	outputFileFlag := &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "output csv file",
	}
	forceFlag := &cli.BoolFlag{
		Name:    "force",
		Aliases: []string{"f"},
		Usage:   "force overwrite of output file",
	}
	files := &cli.StringArgs{
		Name: "excelFiles",
		Min:  1,
		Max:  -1,
	}

	cmd := &cli.Command{
		Name:        "excel-parser",
		Usage:       ShortUsage,
		Description: LongDescription,
		ArgsUsage:   "ExcelFiles",
		Flags: []cli.Flag{
			yamlFlag,
			converterFlag,
			forceFlag,
			outputFileFlag,
		},
		Arguments: []cli.Argument{
			files,
		},
		// Before runs verification before "Action" is run
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {

			if _, err := os.Stat(c.String("yaml")); err != nil {
				return ctx, fmt.Errorf("yaml file %q not found", c.String("yaml"))
			}

			if c.String("converter") == "" {
				return ctx, errors.New("no converter specified")
			}

			// Only allow overwriting the output file if 'force' is in place.
			force := c.Bool("force")
			if !force {
				_, err := os.Stat(c.String("output"))
				if err == nil {
					return ctx, fmt.Errorf("output file %q already exists -- use 'force' to overwrite", c.String("output"))
				}
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return app.Run(
				c.String("yaml"),
				c.String("converter"),
				c.String("output"),
				c.Bool("force"),
				// c.Args(),
				c.StringArgs("excelFiles")...,
			)
		},
	}

	// custom help template.
	// cmd.CustomRootCommandHelpTemplate = cmdHelpTemplate

	return cmd
}

var cmdHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [command] [options]

DESCRIPTION:
   {{.Description}}

COMMANDS:
{{range .Commands}}   {{.Name}}{{ "\t"}}{{.Usage}}
{{end}}
Run '{{.Name}} [command] --help' for more information on a command.
`
