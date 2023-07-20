# Keeper
Keeper is a GPLv3 application designed to privately store your personal info using
custom formats described in YAML templates.
Currently available frontend is console, and currently available backend is Sqlite.
Application is designed to be flexible in both ends so more frontends/backends
could be availabe in the future.

## USAGE SYNOPSIS
```
keeper <template> [-r|read|-R|--READ|-u|--update|-d|--delete] [filter exp]
keeper -t|--template
keeper -t|--template <template>
```

## STORAGE COMMAND OPTIONS

Storage commands always require to specify first a template which will be
used to create a new storage support, following the same name, if not already
available.

(Default command without option parameters is to create a new registry).

    -r, --read      Read storage registries
    -R, --READ      Read storage registries alt (skipping text fields)
    -u, --update    Update storage registries
    -d, --delete    Delete storage registries

For every storage option `-r|-R|-u|-d` `[filter exp]` is applied if available.

## FILTER EXP

Filter expresion could contain one or more field descriptions like those:

    field:"exact match whole field"
    field:%partial_match%
    field:%partial_match_at_end

Partial matches percent symbols can appear at the beginning and/or the end of
the searched expresion, following the SQL "LIKE" standard. 

## INFO COMMAND OPTIONS

    -t|--template               Lists available templates
    -t|--template <template>    List template fields

Info command options are used for autocompletion whenever package
bash-completion is available.

## SAMPLE COMMANDS

```
keeper note                                 Add a new note registry
keeper task -r id:5                         Read from task storage registry with id 5
keeper note -u tag:project1,title:%note%    Update all notes with tag "project1" containing
                                            the word "note" in the field "title"
keeper url -d id:10                         Delete registry from url storage with id 10
```

## TEMPLATE FORMAT

Format for the YAML storage templates per field is:

```
<field>:
 *type: <integer|string|text|autodate|tokens>
  validation:
    required: <true|FALSE>
    regex:    <regex expression>
    tip:      <sample expression to show when regex validation fails>
```
(`*` means required)

- In CLI text fields will invoke the text editor defined in `$EDITOR`.
  `text` multi-line fields can be ignored in readings using the `-R` option.
- `tokens` fields will create a searchable table of different words associated 
  with the registry. One obvious use case is tags.

## ENVIRONMENT DEFAULTS AND CONFIGURATION OPTIONS

*Note: The application has been designed with Windows compatibility in mind but not
tested there yet so please by now take the Windows compatibility with a pinch
of salt.*

Location for the keeper data directory containing the templates and the database
is `$XDG_DATA_HOME/keeper` in Linux or Mac, if `$XDG_DATA_HOME` is undefined the
default is `~/.local/share/keeper`.
For Windows default location is `%APPDATA%/keeper`.
Location can be overriden with the enviroment option `KEEPER_CONFIG_FILE` or 
in the configuration file (see below).

Templates location can be overriden with the enviroment variable
`KEEPER_TEMPLATES_PATH` and database location can be overriden with the variable
`KEEPER_DB_PATH` or in the configuration file (see below).

## CONFIGURATION FILE

Default location for the optional configuration file `keeper.yaml` is
`$XDG_CONFIG_HOME` for Linux or Mac, being `~/.config` if that's not available,
and `%APPDATA%` for Windows. 
Location can be overriden with the enviroment option `KEEPER_CONFIG_FILE`.

The optional YAML configuration file can contain these parameters:

```
storage:      Selected storage engine. Only available option by now is `sqlite`.
path:
  templates:  Location directory for the templates files
  db:         Database filename (full path)
editor:       External editor that will be used with the `text` fields.
              (If not defined it will try to use the enviroment variable `$EDITOR`.)
pager:        Pager for read operations whenever output is bigger than the
              CLI rows (default `$PAGER` or `less`)
```

## TODO

- Add scriptable actions to templates. The idea is to allow automatic processing
  of some fields eg hook some content to an AI to allow automatic tags generation.
- Right now any change in a template format, once the storage already exists would mean 
  that the program refuses to perform any operation. Right now only valid option
  would be to rename the template and use a new storage or perform manually the changes
  in the storage. Adding some support to allow template modifications would be nice.
