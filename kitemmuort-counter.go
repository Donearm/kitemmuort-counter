/*
Author: Gianluca Fiore <forod.g@gmail.com> © 2013-2020
*/

package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"time"
)

var usageString = `
kitemmuort-count (-c|-s) [-d <date>]

Arguments:
	-count|-c
		Show the kitemmuort count for a date (default is today)
	-set|-s
		Set the amount of kitemmuort for a date (default is today)
	-date|-d
		Operate (count/set) on a specific date instead than today
		Use YYYY-MM-DD (example: 2012-10-01)

`
var countArg bool	// show the kitemmuort number or not
var dateArg string	// the date string argument
var setArg int		// the kitemmuort count argument to be set for today

// parse flags
func flagsInit() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageString)
	}

	const (
		fDate	= ""
		fCount = false
		fSet	= 0
	)

	flag.StringVar(&dateArg, "date", fDate, "")
	flag.StringVar(&dateArg, "d", fDate, "")
	flag.BoolVar(&countArg, "count", fCount, "")
	flag.BoolVar(&countArg, "c", fCount, "")
	flag.IntVar(&setArg, "set", fSet, "")
	flag.IntVar(&setArg, "s", fSet, "")

	flag.Parse()

	if countArg != false && setArg != 0 {
		fmt.Fprintf(os.Stderr, "Either use -count or -set, not both\n")
	}
	if countArg == false && setArg == 0 {
		// no argument given, default to show kitemmuort count for today
		countArg = true
	}
}

// checkDbExist checks that the db exists and that the default table has been created
func checkDbExist(db *sql.DB) bool {
	rows, err := db.Query("SELECT * FROM kitemmuorts")
	if err != nil {
		return false
	}
	defer rows.Close()
	if rows != nil {
		return true
	}

	return false
}

// createTable creates the default table to host the application's data
func createTable(db *sql.DB) {
	cStmt, err := db.Prepare("CREATE TABLE kitemmuorts(date text, count int, UNIQUE (date));")
	if err != nil {
		log.Fatal(err)
	}

	exResult, err := cStmt.Exec()
	// check that table doesn't already exist
	if err != nil {
		log.Println(err)
		fmt.Println(exResult)
	}
}

// returnHomeDir returns the path of the current user's home directory
func returnHomeDir() string {
	if homedir := os.Getenv("HOMEPATH"); homedir != "" {
		return homedir
	} else if homedir := os.Getenv("HOME"); homedir != "" {
		return homedir
	} else {
		return ""
	}
}

// formatDateString formats the dateString flag to the appropriate format for 
// the SQLite database
func formatDateString(d string) string {
	const dateLayout = "2006-01-02" // the date format layout, to match
	// how date is stored in the sqlite db

	if d == "" {
		// empty date, use today's
		t := time.Now()
		return t.Format(dateLayout)
	}
	t, err := time.Parse(dateLayout, d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing the date, is it in YYYY-MM-DD format?\n")
		log.Fatal(err)
	}
	return t.Format(dateLayout)
}

func main() {
	var dbReady bool	// is the db ready?
	var dbFile string	// the database filename

	flagsInit()

	homedir := returnHomeDir()

	dbFile = filepath.Join(homedir, ".kitemmuort.db")

	// check the db file exists
	if _, err := os.Stat(dbFile); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(dbFile)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// file exists but perhaps it can't be accessed?
			log.Fatal(err)
		}
	}
	// open connection to the db
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbReady = checkDbExist(db)
	if dbReady != true {
		createTable(db)
	}

	// set the date to query on
	t := formatDateString(dateArg)

	if countArg {
		var date string
		var count int
		err := db.QueryRow("SELECT date, count FROM kitemmuorts WHERE date = ?", t).Scan(&date, &count)
		if count == 0 {
			// nothing has been set, yet, for date, exit
			fmt.Fprintf(os.Stdout, "\nNo kitemmuorts set for %s yet. Want to add some?\n", t)
			os.Exit(0)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "\nKitemmuort count for %s is %d\n", date, count)
	}

	if setArg != 0 {
		stmt, err := db.Prepare("INSERT OR REPLACE INTO kitemmuorts(date, count) VALUES(?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		_, execErr := stmt.Exec(t, setArg)
		if err != nil {
			log.Fatal(execErr)
		}
		fmt.Fprintf(os.Stdout, "\n%d kitemmuorts set for %s\n", setArg, t)
	}
}
