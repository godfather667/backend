// Package backend implements an Rest API that manages a "company" database.
//
// Server is on localhost:3000
// Responds to endpoint "/backend/"
//
// Specify the record <Command String>:
//     /NEW/<data string>   			- Creates a new record (next id avaiable)
//     /UPDATE/id/<data string>			- Updates specified record (id > 0)
//	   /GET/id/							- Returns spedified record (id > 0)
//     /DELETE/id/						- Deletes specified record (id > 0)
//	   /LIST/							- Lists all records in database
//	   /FIND/<Find string>		    - Lists records matching <Find Specification>
//         - <Find String>: |TAG <Search string> |TAG <Search String> ...
//	       		- May contains logic Tags:  |AND and/or |OR  - (|OR is assumed between pairs}
//			 <Search String>: Matchs if <TAG Value> "contains" the <search string>
//	   /EXIT/							- Close Database and exit
//
// Example: https://localhost:3000/backend/NEW |TAG <string> |TAG <string> < | or End of Line>
//
//	 - End Point /backend/<Command String><ID> String><Data String>
//         	- <ID String> only present for <Command Strings> = GET, DELETE and UPDATE
//			- <Command String>: NEW, UPDATE, GET, DELETE, LIST, or FIND
//				- <Commands Strings> are always Upper CASE

//   		- <Data string>: - Composed of <!TAG><String> Pairs
//						     - Terminated by | or end of line
//
//			. <Data String> are not allowed for Commands GET and DELETE
//			- Command LIST does not have <ID String>  or <Data String>
//
//	 NOTES:
//		- TAGs should be unique (If #TAG <string> #TAG <string> (last tag value used))
//		- Commands may be any case (NEW same as new)
//		- #TAG must be an exact match (TAG and tag are NOT the same)
//		- Escaping the | Character in <string>:  || == |  (Not allowed in TAGS)
//		- All <strings> "are" strings! Numbers are specified as <strings>
//			- for example: "|NAME John Smith |AGE 20"  (Data Strings Implicitly strings)
//
// Result of examples:
//
//
//
// Error Messages:  If a viable stock name is not found:
//     "Error! The requested stock(s) could not be found."
//
// Unrecoverable Errors are handled by the "log" package.
//
// This program uses the minimum standard libraries to accomplish its goals!
// "fmt"  - For displaying text.
// "io/ioutil" - For reading the response body.
// "net/http" - For REST Operations.
// "strings" - For manipulating strings.
// "unicode" - For determining text classes (ie string, number..).
// "log" - For reporting Fatal Errors!
//
package main

import (
	"fmt"
	_ "io/ioutil"
	"log"
	"net/http"
	"strings"
	"unicode"

	_ "github.com/prologic/bitcask"
)

// Constants
const fileName = "$HOME/db"

// Global Variables
// 	- db variable holds the reference to the data base handle.

// var db Bitcask.Bitcask

// Types
// TagPair structure holds the data for each "tag pair" in the <data string>
type TagPair struct {
	tagName  string
	tagValue string
}

/*
func init() {
	db, err = bitcask.Open("/tmp/db")
	defer db.Close()

}
*/

// acquireInput - Acquire URL data from HTTP Resource.
//		- Input all the data from URL.patn
//		- Insure that the TAG precursor '|' is always bound by space character.
//		- Add a " |' to the end of <data string>
//		- Crack Inout into fields and return as a slice
//
func acquireInput(w http.ResponseWriter, r *http.Request) []string {
	total := r.URL.Path                            //debug
	fmt.Println(total)                             //debug
	total = strings.Replace(total, "|", " | ", -1) //Insure '|' is bounded by spaces
	if !strings.HasSuffix(total, "|") {
		total += " |" // Insure that <data string> ends with a Vertical Bar
		fmt.Println(total)
	}
	var vb int32 = 0x007C // Vertcal Bar
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != vb
	}
	fmt.Println(total) //debug
	fields := strings.FieldsFunc(total, f)

	fmt.Println("Exiting acquire:")
	return fields
}

// parseVertBar - The function processing FIELDS slice finding the location
//			of the Vertical Bar characters (Data Field delimiters). It
//			creates a slice containing the location of Vertical Bars in the
//			Field Array,
// Additionally the function detects whether the last two items in the
//			Field Array both contain Vertical Bars, Since the acquireInput
//			function adds a " |" to the end of the the <Data string> and
//			insures that Vertica Bars are surronded by spaces. It is neccessary
//			to check if <data string> ends in two Verticals. If it does the
//		    last one is removed.
func parseVertBar(fields []string) []int {
	fmt.Println("parseVertBar: Fields: ", fields)
	vtList := make([]int, 0)
	for i, v := range fields {
		vt := strings.TrimSpace(v)
		if vt == "|" {
			vtList = append(vtList, i)
		}
	}
	if len(vtList) > 1 {
		fmt.Println("vtList Pair Check: ", len(vtList)-2, "  ", len(vtList)-1)
		if vtList[len(vtList)-2] == vtList[len(vtList)-1]-1 {
			vtList = vtList[:len(vtList)-1]
		}
	}
	return vtList
}

// parseTags - Reads the FIELD Slice and isolates each TagPair in the
//			<data string<> and creates a TagPair Struct <TAG><Value>,
//			These values are stored in the "tagList" slice.
func parseTags(fields []string) []TagPair {
	fmt.Println("parseTags: Fields: ", fields)
	var tagList []TagPair
	vtList := parseVertBar(fields)
	fmt.Println("vtList: ", vtList)
	for i := 0; i < len(vtList)-1; i++ {
		j := i + 1
		vi := vtList[i]
		vj := vtList[j]
		span := vj - vi
		fmt.Println("Span: ", span, "vi: ", vi, "vj: ", vj)
		if span < 2 {
			log.Fatal("Empty Tag Pair!")
		}
		tn := fields[vi+1]
		fmt.Println(tn)
		tv := ""
		for tvi := vi + 2; tvi < vj; tvi++ {
			fmt.Println("tvi: ", tvi)
			tv += fields[tvi] + " "
		}
		tagList = append(tagList, TagPair{tn, tv})
	}
	return tagList
}

// backendHandler - This function controls the programs processing.
func backendHandler(w http.ResponseWriter, r *http.Request) {
	fields := acquireInput(w, r)
	fmt.Println(len(fields))
	for v, i := range fields { //debug
		fmt.Println("Field[", i, "] = ", v) //debug
	}
	tagList := parseTags(fields)
	fmt.Println("tagList: ", tagList)

}

//
// main - Sets up handler and waits on traffic.
//
func main() {
	http.HandleFunc("/backend/", backendHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
