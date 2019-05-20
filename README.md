# backend
A Flexible Database for storing records with mixed Columns

 Package **backend** implements an Rest API that manages a "company" database.
* This package implements a two level Key/Store Data Base:
* The Data Records are indexed by ID Number as the Key
* The Data Records are implemented internally as a Key/Store
* Data Records are made up of a variable number of TAG Pairs
* TAGS are identified by the first character which is a "|" ie |TAG
* The |TAG is followed by any number of Words

Example: |NAME John Smith |AGE 50

Since Records are Key/Stores the Record Format is completely flexible.

Each Record can have Unique Tag Pairs - The User can add special Tag to Any Record:  For Example:  

** Server is on localhost:3000** Responds to endpoint `/backend/<CMD> [ <data list> ]`

## Syntax Description
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
<CMD> <Data List>

<CMD> = Single Upper Case Text Word

<Data List> = <Tag Pair> [ <Tag Pair> ... ]

<Tag Pair> = <TAG> <Word> [ <Word> ... ]

<TAG> = <|><Single Upper Case Text Word><space>

<Word> = <space><Text><space>
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
##  Commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
     /NEW <data list>                  - Creates a new record (ID # -Automatic)
     /UPDATE <|ID Number> <data list>  - Updates specified ID
     /GET/<ID Record Number>           - Returns spedified ID
     /DELETE/<ID Record Number>        - Deletes specified ID
     /LIST/                            - Lists all records
     /EXIT/                            - Close Database and exit
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
    Example: https://localhost:3000/backend/NEW |NAME John Smith |AGE 50

    Example: https://localhost:3000/backend/UPDATE |ID Rec# |NAME Mike Browm

	 NOTES:
		- TAGs should be unique ( |NAME <string> |NAME <string> ) Is Wrong!!
		- Commands should be in UPPER CASE (For CLarity)
		- !TAG must be an exact match (TAG and tag are NOT the same)
        - TAGS should be in UPPER CASE (For CLarity)
		- There is no escape for the "|" Character is the TAG Prefix identifier
		- All <data list Tags and Words are strings! Numbers are specified as <strings>

## Example of LIST Command
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
CMD:  LIST

1   |NAME John Jones |AGE 50
2   |NAME Mark Smith |AGE 29
3   |NAME Bob Brown  |AGE 35
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 Error Messages:
     "Recoverable Errors are Display and Processing Continues"

 Unrecoverable Errors are handled by the "log.Fatal" package function.
