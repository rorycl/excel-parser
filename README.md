# excel-parser

A command line program for parsing Excel files and extracting the
desired columns from sheets matching a header "recipe". The recipe and
column definitions are defined in a yaml file.

The parser can extract columns from sheets with arbitrarily inserted
columns, so long as the header recipe regexp still matches.

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
