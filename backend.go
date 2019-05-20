// Package backend implements an Rest API that manages a "company" database.
//		This package implements a two level Key/Store Data Base:
//		The Data Records are indexed by ID Number ID/RECORD DATA
//		The Data Records are implemented internally as a Key/Store
//			Data Records are made up of a variable number of TAG Pairs
//.				TAGS are identified by the first character which is a "|" ie |TAG
//					The |TAG is followed by any number of Words
//						Example: |NAME John Smith |AGE 50
//		Since Records are Key/Stores the Record Format is Flexiable.
//			Each Record can have any Tag Pairs - The User can a special Tags
//			to Any Record:  For Example:  Some Records might be Flagged in some way.
//
// Server is on localhost:3000
// Responds to endpoint "/backend/<cmd> <data list>"
//
// Command Format:
// 		<data list> is made up a variable number of <tag pairs>
//			<tag pairs> consist of a <|TAG> and variable number of <text words>
// For Examples:
//			/NEW |NAME John Smith |AGE 56
//			/UPDATE |ID 10 |NAME Bob Lay  - Changes Name Value of ID 10
//
//  Commands:
//
//     /NEW <data list>   		- Creates a new record (Using next avaiable ID Value)
//     /UPDATE <|ID Number> <data list>	 - Updates specified ID
//	   /GET/<ID string>						 - Returns spedified ID
//     /DELETE/<ID Number>					 - Deletes specified ID
//	   /LIST/								 - Lists all records in database
//	   /EXIT/		 						 - Close Database and exit
//
// Example: https://localhost:3000/backend/NEW |NEW John Smith |AGE 50
//
//	 NOTES:
//		- TAGs should be unique (If |NAME <string> |NAME <string> (Both will be present?)
//		- Commands should be in UPPER CASE (For CLarity)
//		- !TAG must be an exact match (TAG and tag are NOT the same)
//		- There is no escape for the "|" Character as it is the TAG Prefix identifier
//		- All <data list Tags and Words are strings! Numbers are specified as <strings>
//
//
// Result List Example:
//CMD:  LIST   tagList:  []
//1   |NAME John jones |AGE 50
//2   |NAME John jones |AGE 50
//3   |NAME John jones |AGE 50
//
// Error Messages:
//     "Recoverable Errors are Display and Processing Continues"
//
// Unrecoverable Errors are handled by the "log" package.
//
package main

import (
	"fmt"
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
//  - Special Buffer for Update Mode
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
	logFatal("Database Open Failure: ", err)
}

//
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

//
// top Function - Finds and Stores the Highest Index Key in the database
//
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

//
// formOut Function - Converts a data string into a tagList
//
func formOut(value string) []TagPair {
	value = value + "|"
	var vb int32 = 0x007C // Vertcal Bar
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != vb
	}
	value = fieldZero[0] + " " + fieldZero[1] + " " + value
	value = strings.Replace(value, "|", " | ", -1) //Insure '|' is bounded by spaces

	fields := strings.FieldsFunc(value, f)
	_, tagList := parseTags(fields, false)
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
	total = strings.Replace(total, "|", " | ", -1) //Insure '|' is bounded by spaces
	if !strings.HasSuffix(total, "|") {
		total += " |" // Insure that <data string> ends with a Vertical Bar
	}
	var vb int32 = 0x007C // Vertcal Bar
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != vb
	}
	fields := strings.FieldsFunc(total, f)
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
	vtList := make([]int, 0)
	for i, v := range fields {
		vt := strings.TrimSpace(v)
		if vt == "|" {
			vtList = append(vtList, i)
		}
	}
	if len(vtList) > 1 {
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
	if flag {
		fieldZero = fields
	}
	var tagList []TagPair
	vtList := parseVertBar(fields)
	for i := 0; i < len(vtList)-1; i++ {
		j := i + 1
		vi := vtList[i]
		vj := vtList[j]

		tn := fields[vi+1]
		tv := ""
		for tvi := vi + 2; tvi < vj; tvi++ {
			tv += fields[tvi] + " "
		}
		tagList = append(tagList, TagPair{tn, tv})
	}
	return fields[1], tagList
}

//
// Command Processing Function - The Processing for each Command is here
//
func CommandLoop(w http.ResponseWriter, r *http.Request) error {

	fields := acquireInput(w, r)
	cmd, tagList := parseTags(fields, true)
	if len(tagList) < 1 {
		fmt.Println("CMD: ", cmd)
	} else {
		fmt.Println("CMD: ", cmd, "  tagList: ", tagList)
	}

	switch cmd {

	case "NEW":
		out := ""
		for i := 0; i < len(tagList); i++ {
			out += "|" + tagList[i].tagName
			out += " " + tagList[i].tagValue
		}
		dbi := strconv.Itoa(recIndex())
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
			if err == nil {
				if keyIdx > topIndex {
					topIndex = keyIdx
				}
			}
			return nil
		})

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
		if tagList[0].tagName == "ID" {
			dbi := tagList[0].tagValue
			dbi = strings.TrimSpace(dbi)
			if len(tagList) < 1 {
				fmt.Println("Invalid ID Record Number!")
			} else {
				out := ""
				for i := 0; i < len(tagList); i++ {
					out += "|" + tagList[i].tagName
					out += " " + tagList[i].tagValue
				}
				newList := formOut(string(out))
				value, _ := db.Get(dbi)
				if len(value) < 1 {
					fmt.Println("Requested Record ", dbi, " has been deleted!")
				} else {
					tagList = formOut(string(value))
					for ti := 0; ti < len(tagList); ti++ {
						for ni := 0; ni < len(newList); ni++ {
							if tagList[ti].tagName == newList[ni].tagName {
								tagList[ti].tagValue = newList[ni].tagValue
								break
							}
						}
					}
					out := ""
					for i := 0; i < len(tagList); i++ {
						out += "|" + tagList[i].tagName
						out += " " + tagList[i].tagValue
					}
					db.Put(dbi, []byte(out))
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
