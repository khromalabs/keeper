# keeper
Keeper is an application to store your personal data. Formats are described
in YAML templates. Currently available frontend is console, and currently
available backend is Sqlite. Application is designed to be flexible in both
ends so more frontends/backends could be availabe in the future.

## STORAGE COMMAND OPTIONS

Default command (no option parameters) is to create a new registry. 

    -r, --read      Read storage registries
    -R, --READ      Read storage registries alt (skipping text fields)
    -u, --update    Update storage registries
    -d, --delete    Delete storage registries

For every storage option (-r|-R|-u|-d) [filter exp] is applied if available.

## INFO COMMAND OPTIONS

    -t|--template               Lists available templates
    -t|--template <template>    List template fields

Info command options are used for autocompletion whenever package
bash-completion is available.

## FILTER EXP

Filter expresion could contain one or more field descriptions like those:

    field1:searched_word \
    field2:"quoted words to search in exact order" \
    field3:%%partial_match%%

Partial matches percent symbols can appear at the beginning and/or the end of
the searched expresion, following the SQL "LIKE" standard. 

Sample commands:

keeper note                                 Add a new note registry
keeper task -r id:5                         Read from task storage
                                            registry with id 5
keeper note -u tag:project1,title:%%note%%    Update all notes with tag "project1"
                                            containing "note" in the title
keeper url -d id:10                         Delete registry from url storage
                                            with id 10
## TEMPLATES

Format for the YAML storage templates is (* means required):

<field>:
\*type:integer/string/text/autodate/tokens
 validation:
   required:true/false (default: false)
   regex: <regex expression>
   tip: <sample expression to show when regex validation fails>

- In CLI text fields will invoke the text editor defined in $EDITOR.
  "text" fields can be leaved out of registries readings using the -R option.
- Tokens fields will create a searchable table of different words associated 
  with the registry. One obvious use case is tags.

## ENVIRONMENT DEFAULTS AND CONFIGURATION OPTIONS

Default location for the optional configuration file "keeper.yaml" is
$XDG_CONFIG_HOME for Linux or Mac, being ~/.config if that's not availabe,
and %APPDATA% for Windows. 
Location can be overriden with the enviroment option KEEPER_CONFIG_FILE.

Defalt location for the keeper data directory containing the templates and the database
is $XDG_DATA_HOME/keeper in Linux or Mac, being ~/.local/share/keeper
if that's not available. For Windows default location is %APPDATA%/keeper.
Templates location can be overriden with the enviroment variable
KEEPER_TEMPLATES_PATH and database location can be overriden with the variable
KEEPER_DB_PATH.

## KEEPER CONFIGURATION FILE

The optional YAML configuration file can contain this parameters:

storage:      Selected storage engine. Only available option by now is Sqlite.
path:
  templates:  Location directory for the templates files
  db:         Database filename (full path)
editor:       External editor that will be used with the "text" fields.
              (If not defined it will try to use the enviroment variable $EDITOR.)
pager:        Pager for read operations whenever output is bigger than the
              CLI rows
