# backend
Job Test Progam

Package **backend** implements an Rest API that manages a "company" database.

 Server is on localhost:3000

 Responds to endpoint "/backend/"

 Specify the record <Command String>:
     /NEW/<data string>   			- Creates a new record (next id avaiable)
     /UPDATE/id/<data string>			- Updates specified record (id > 0)
	   /GET/id/							- Returns spedified record (id > 0)
     /DELETE/id/						- Deletes specified record (id > 0)
	   /LIST/							- Lists all records in database
	   /FIND/<Find string>		    - Lists records matching <Find Specification>
         - <Find String>: |TAG <Search string> |TAG <Search String> ...
	       		- May contains logic Tags:  |AND and/or |OR  - (|OR is assumed between pairs}
			 <Search String>: Matchs if <TAG Value> "contains" the <search string>

 Example: https://localhost:3000/backend/NEW |TAG <string> |TAG <string> < | or End of Line>

	 - End Point /backend/<Command String><ID> String><Data String>
         	- <ID String> only present for <Command Strings> = GET, DELETE and UPDATE
			- <Command String>: NEW, UPDATE, GET, DELETE, LIST, or FIND
				- <Commands Strings> are always Upper CASE

   		- <Data string>: - Composed of <!TAG><String> Pairs
						     - Terminated by | or end of line

			. <Data String> are not allowed for Commands GET and DELETE
			- Command LIST does not have <ID String>  or <Data String>

	 NOTES:
		- TAGs should be unique (If #TAG <string> #TAG <string> (last tag value used))
		- Commands may be any case (NEW same as new)
		- #TAG must be an exact match (TAG and tag are NOT the same)
		- Escaping the | Character in <string>:  || == |  (Not allowed in TAGS)
		- All <strings> "are" strings! Numbers are specified as <strings>
			- for example: "|NAME John Smith |AGE 20"  (Data Strings Implicitly strings)

 Result of examples:



 Error Messages:  If a viable stock name is not found:
    "Error! The requested stock(s) could not be found."

 
