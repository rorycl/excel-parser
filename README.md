# excel-parser

A command line program for parsing Excel files and extracting the
desired columns from sheets matching a header "recipe". The recipe and
column definitions are defined in a yaml file. Additional columns
can be added.

The program is best suited to extracting and transforming data from
rigid format Excel files, such as those recording payment data on
platforms such as JustGiving or Enthuse. The anonymised examples in
`testdata` are from these providers.

Several different converters can be defined in the same yaml file. The
example `config_example.yaml` provides two, one each for JustGiving and
Enthuse.

## Usage

The parser can extract columns from sheets with arbitrarily inserted
columns, so long as the header recipe regexp still matches the header
line.

Columns may be added using `additionalColumns` and `additionalData` for
any particular converter in  the yaml file. The `additionalData` fields
can be defined using either verbatim strings or Go templates, the latter
utilising either the special field `{{ .Filename }}` or the name of a
column header in the original data. Note that column header names have
spaces removed and are Title cased, so the header ` donation date `
becomes referrable as `{{.DonationDate}}` in Go template syntax.

Two template funcs are provided for the additional data fields:

* TimestampParseAndFormat  
  Example:  
  `ABC-{{ "2026-09-01 09:01" | TimestampParseAndFormat "2006-01-02 15:04" "02/01/2006" }}`  
  returns:  
  `ABC-01/09/2026`

* RegexpReplace  
  Example:  
  `{{ "<please replace me>" | RegexReplace "(replace.*me)" "fix" }}`  
  returns:  
  `<please fix>`

## Run

```
NAME:
   excel-parser - Process excel files into a summary csv file.

USAGE:
   excel-parser [global options] ExcelFiles

DESCRIPTION:
   This cli program uses a configuration yaml file to load all of
   the Excel files listed on the command line and/or in the specified
   directories into the specified csv file.

GLOBAL OPTIONS:
   --yaml string, -y string       configuration yaml file
   --converter string, -c string  converter choice from yaml
   --force, -f                    force overwrite of output file
   --output string, -o string     output csv file
   --help, -h                     show help
```

## Licence

This project is licensed under the [MIT Licence](LICENCE).
