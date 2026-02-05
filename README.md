# excel-parser

A command line program for parsing Excel files and extracting the
desired columns to a csv file.

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
