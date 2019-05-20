// Package backend implements an Rest API that manages a "company" database.
//
// Server is on localhost:3000
// Responds to endpoint "/backend/"
//
// Specify the record <Command String>:
//     /NEW/<data string>   			- Creates a new record (next id avaiable)
//     /UPDATE/<ID string><data string>			- Updates specified record (id > 0)
//			  <ID string> <ID REC #>
//	   /GET/<ID string>						- Returns spedified record (id > 0)
//     /DELETE/<ID string>					- Deletes specified record (id > 0)
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
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/prologic/bitcask"
)

// Constants
const fileName = "./db"

// Global Variables
// 	- db variable holds the reference to the data base handle.
//  - dbi variable holds Record Index for writing records
//  - err Error variables
//  - Top Index detected from Fold Command
var db *bitcask.Bitcask
var dbi int = 0
var err error
var topIndex = 0
var fieldZero []string

// Types -
// TagPair structure holds the data for each "tag pair" in the <data string>
//
type TagPair struct {
	tagName  string
	tagValue string
}

//
// Init() Function - Open Database and Set Database Handle "db".
//
func init() {
	db, err = bitcask.Open(fileName)
	//	defer db.Close()
	logFatal("Database Open Failure: ", err)
}

// logFatal Function - Handlers Fatal Error.
//
func logFatal(errMsg string, err error) {
	if err != nil {
		log.Fatal(errMsg, err)
	}
}

//
// recIndex Function - Returns the Next Record Index.
//
func recIndex() int {
	dbi += 1
	return dbi
}

func top(key string) error {
	key = strings.TrimSpace(key)
	newKey, keyErr := strconv.Atoi(key)
	if keyErr != nil {
		fmt.Println("Key Conversion to integer error: ", keyErr)
		return keyErr
	}
	if newKey > topIndex {
		topIndex = newKey
	}
	return nil
}

func formOut(value string) []TagPair {
	value = value + "|"
	var vb int32 = 0x007C // Vertcal Bar
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != vb
	}
	//	fmt.Println("formOut Update: ", value)
	value = fieldZero[0] + " " + fieldZero[1] + " " + value
	value = strings.Replace(value, "|", " | ", -1) //Insure '|' is bounded by spaces

	//	fmt.Println("New Value: ", value)
	fields := strings.FieldsFunc(value, f)
	//	fmt.Println("Fields: ", fields)
	_, tagList := parseTags(fields, false)
	//	fmt.Println("CMD: ", cmd, "  tagList: ", tagList)
	return tagList
}

// acquireInput - Acquire URL data from HTTP Resource.
//		- Input all the data from URL.patn
//		- Insure that the TAG precursor '|' is always bound by space character.
//		- Add a " |' to the end of <data string>
//		- Crack Inout into fields and return as a slice
//
func acquireInput(w http.ResponseWriter, r *http.Request) []string {
	total := r.URL.Path
	//	fmt.Println("total: ", total)
	//	fmt.Println(total)                             //debug
	total = strings.Replace(total, "|", " | ", -1) //Insure '|' is bounded by spaces
	if !strings.HasSuffix(total, "|") {
		total += " |" // Insure that <data string> ends with a Vertical Bar
		//		fmt.Println(total)
	}
	var vb int32 = 0x007C // Vertcal Bar
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != vb
	}
	//	fmt.Println(total) //debug
	fields := strings.FieldsFunc(total, f)

	//	fmt.Println("Fields: ", fields)
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
	//	fmt.Println("parseVertBar: Fields: ", fields)
	vtList := make([]int, 0)
	for i, v := range fields {
		vt := strings.TrimSpace(v)
		if vt == "|" {
			vtList = append(vtList, i)
		}
	}
	if len(vtList) > 1 {
		//		fmt.Println("vtList Pair Check: ", len(vtList)-2, "  ", len(vtList)-1)
		if vtList[len(vtList)-2] == vtList[len(vtList)-1]-1 {
			vtList = vtList[:len(vtList)-1]
		}
	}
	return vtList
}

// parseTags - Reads the FIELD Slice and isolates each TagPair in the
//			<data string<> and creates a TagPair Struct <TAG><Value>,
//			These values are stored in the "tagList" slice.
func parseTags(fields []string, flag bool) (string, []TagPair) {
	//	fmt.Println("parseTags Input: Fields: ", fields)
	if flag {
		fieldZero = fields
		//		fmt.Println("fieldZero: ", fieldZero)
	}
	var tagList []TagPair
	//	fmt.Println("Enter ParseVertBar", fields)
	vtList := parseVertBar(fields)
	//	fmt.Println("vtList: ", vtList)
	for i := 0; i < len(vtList)-1; i++ {
		j := i + 1
		vi := vtList[i]
		vj := vtList[j]

		tn := fields[vi+1]
		//		fmt.Println(" tn: ", tn)
		tv := ""
		for tvi := vi + 2; tvi < vj; tvi++ {
			//fmt.Println("tvi: ", tvi)
			tv += fields[tvi] + " "
		}
		tagList = append(tagList, TagPair{tn, tv})
	}
	//	fmt.Println("parseTags: ", fields[1], " ", tagList)
	return fields[1], tagList
}

func CommandLoop(w http.ResponseWriter, r *http.Request) error {

	fields := acquireInput(w, r)
	cmd, tagList := parseTags(fields, true)
	fmt.Println("CMD: ", cmd, "  tagList: ", tagList)

	switch cmd {

	case "NEW":
		out := ""
		for i := 0; i < len(tagList); i++ {
			out += "|" + tagList[i].tagName
			out += " " + tagList[i].tagValue
		}
		dbi := strconv.Itoa(recIndex())
		fmt.Println("Output Record: ", dbi, " ", out)
		db.Put(dbi, []byte(out))

	case "LIST":
		topIndex = 0
		var value []byte
		err = db.Fold(func(key string) error {
			value, err = db.Get(key)
			if err != nil {
				return err
			}

			keyIdx, err := strconv.Atoi(key)
			//			fmt.Println("keyIdx", keyIdx)
			if err == nil {
				if keyIdx > topIndex {
					topIndex = keyIdx
				}
			}
			return nil
		})
		fmt.Println("topIndex: ", topIndex)

		for i := 1; i < topIndex+1; i++ {
			dbi := strconv.Itoa(i)
			value, _ := db.Get(dbi)
			if len(value) > 1 {
				fmt.Println(dbi, " ", string(value))
			}
		}

	case "DEL":
		if tagList[0].tagName == "ID" {
			dbi := tagList[0].tagValue
			dbi = strings.TrimSpace(dbi)
			if len(tagList[0].tagValue) > 1 {
				err = db.Delete(dbi)
				if err != nil {
					fmt.Println("Delete Error: ", err)
				}
			} else {
				fmt.Println("Delete Record: ", dbi)
			}
		} else {
			fmt.Println("DEL Did not contain ID Tag Pair")
		}

	case "GET":
		if tagList[0].tagName == "ID" {
			dbi := tagList[0].tagValue
			dbi = strings.TrimSpace(dbi)
			if len(tagList) < 1 {
				fmt.Println("Invalid ID Record Number!")
			} else {
				value, _ := db.Get(dbi)
				if len(value) < 1 {
					fmt.Println("Requested Record ", dbi, " has been deleted!")
				} else {
					fmt.Println(dbi, " ", string(value))
				}
			}
		}

	case "UPDATE":
		//fmt.Println("Update Processing")
		if tagList[0].tagName == "ID" {
			dbi := tagList[0].tagValue
			dbi = strings.TrimSpace(dbi)
			if len(tagList) < 1 {
				fmt.Println("Invalid ID Record Number!")
			} else {
				value, _ := db.Get(dbi)
				if len(value) < 1 {
					fmt.Println("Requested Record ", dbi, " has been deleted!")
				} else {
					//fmt.Println("Update Value: ", dbi, " ", string(value))
					// Update Record Processing
					tagList = formOut(string(value))
					fmt.Println("formOut: ", tagList)
				}
			}
		}

	case "EXIT":
		db.Close()
		os.Exit(0)

	default:
		fmt.Println("Bad Command Value: ")
	}

	return nil // Currently no errors to report other than logFatal
}

// backendHandler - This function controls the programs processing.
func backendHandler(w http.ResponseWriter, r *http.Request) {
	err := CommandLoop(w, r)
	logFatal("Command Loop Crashed: ", err)
}

//
// main - Sets up handler and waits on traffic.
//
func main() {
	http.HandleFunc("/backend/", backendHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
	for {
		http.HandleFunc("/backend/", backendHandler)
	}
}
