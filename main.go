package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Env struct {
	db *sql.DB
}

func (env *Env) updateHandler(w http.ResponseWriter, r *http.Request) {
	// Prepare a statement
	stmt, err := env.db.Prepare(`INSERT INTO data VALUES(NULL,NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		log.Panic("Statement prepare error")
		w.Write([]byte(`"status":"1"`))
		return
	}

	// Execute query
	_, err = stmt.Exec(r.FormValue("DevBat"), r.FormValue("Gids"), r.FormValue("Lat"), r.FormValue("Long"), r.FormValue("Elv"), r.FormValue("Seq"), r.FormValue("Trip"), r.FormValue("Odo"),
		r.FormValue("SOC"), r.FormValue("AHr"), r.FormValue("BatTemp"), r.FormValue("Amb"), r.FormValue("Wpr"), r.FormValue("PlugState"), r.FormValue("ChgrMode"), r.FormValue("ChrgPwr"),
		r.FormValue("VIN"), r.FormValue("PwrSw"), r.FormValue("Tunits"), r.FormValue("RPM"), r.FormValue("SOH"))

	if err != nil {
		log.Panic("Query execution error")
		w.Write([]byte(`"status":"1"`))
		return
	}

	// This sends feedback to LeafSpy that operation was successful
	w.Write([]byte(`"status":"0"`))

	// Reroute data to leaf-status.com with credentials
	leafstatusURL := strings.Replace(r.URL.RawQuery, "user=", "user="+url.QueryEscape(os.Getenv("leafstatus_user")), 1)
	leafstatusURL = strings.Replace(leafstatusURL, "pass=", "pass="+url.QueryEscape(os.Getenv("leafstatus_pass")), 1)
	leafstatusURL = "https://leaf-status.com/api/vehicle/update?" + leafstatusURL

	_, err = http.Get(leafstatusURL)

	if err != nil {
		log.Panic("Leaf-status.com error")
	}

}

func (env *Env) indexHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// Initialize .env file
	err := godotenv.Load("config.env")
	if err != nil {
		log.Panic(err)
	}

	// Check if error log file is set
	if os.Getenv("log_file") != "" {
		// Open log file
		f, err := os.OpenFile(os.Getenv("log_file"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

		if err != nil {
			log.Println(err)
		}

		defer f.Close()

		// Set log to output
		log.SetOutput(f)

		// Log that we started the program
		log.Println("Program started")
	}

	// Open database connection
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s)/%s", os.Getenv("db_user"), os.Getenv("db_pass"), os.Getenv("db_host"), os.Getenv("db_schema"))

	db, err := sql.Open(os.Getenv("db_type"), dbDSN)
	if err != nil {
		log.Panic(err)
	}

	env := &Env{db: db}

	defer env.db.Close()

	// Set up web server
	http.HandleFunc("/update", env.updateHandler)
	http.HandleFunc("/", env.indexHandler)
	http.ListenAndServe(":"+os.Getenv("http_port"), nil)
}
