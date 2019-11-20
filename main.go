package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	_ "text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Env is used for global variables (e.g. database handles)
type Env struct {
	db *sql.DB
}

// LeafDataRow represents a row of data received by LeafSpyPro
type LeafDataRow struct {
	ID        int32     `db:"id"`
	Timestamp time.Time `db:"time"`
	DevBat    int8      `db:"DevBat"`
	Gids      int16     `db:"Gids"`
	Lat       float32   `db:"Lat"`
	Long      float32   `db:"Long"`
	Elv       int32     `db:"Elv"`
	Seq       int32     `db:"Seq"`
	Trip      int32     `db:"Trip"`
	Odo       float32   `db:"odo"`
	SOC       float32   `db:"SOC"`
	AHr       float32   `db:"AHr"`
	BatTemp   float32   `db:"BatTemp"`
	Amb       float32   `db:"Amb"`
	Wpr       int8      `db:"Wpr"`
	PlugState int8      `db:"PlugState"`
	ChrgMode  int8      `db:"ChrgMode"`
	ChrgPwr   int32     `db:"ChrgPwr"`
	VIN       string    `db:"VIN"`
	PwrSw     int8      `db:"PwrSw"`
	Tunits    string    `db:"Tunits"`
	RPM       int32     `db:"RPM"`
	SOH       float32   `db:"SOH"`
	OdoMi     float32
}

// retrieveLastRow retrieves latest row of data in the database
func (dataRow *LeafDataRow) retrieveLastRow(env *Env) {
	// Prepare statement
	row := env.db.QueryRow("SELECT * FROM data ORDER BY time DESC LIMIT 1")

	err := row.Scan(&dataRow.ID, &dataRow.Timestamp, &dataRow.DevBat, &dataRow.Gids, &dataRow.Lat, &dataRow.Long, &dataRow.Elv, &dataRow.Seq, &dataRow.Trip, &dataRow.Odo, &dataRow.SOC, &dataRow.AHr, &dataRow.BatTemp, &dataRow.Amb, &dataRow.Wpr, &dataRow.PlugState, &dataRow.ChrgMode, &dataRow.ChrgPwr, &dataRow.VIN, &dataRow.PwrSw, &dataRow.Tunits, &dataRow.RPM, &dataRow.SOH)

	if err != nil {
		log.Panic("Error retrieving last data")
	}

	// Convert data
	dataRow.OdoMi = dataRow.Odo / 1.609

}

// updateHandler handles the /update part of the webserver
func (env *Env) updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Prepare a statement
	stmt, err := env.db.Prepare(`INSERT INTO data VALUES(NULL,NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		log.Panic("Statement prepare error:")
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

// BasePage contains data for / template
type BasePage struct {
	DataRow     LeafDataRow
	GMapsAPIKey string
}

// baseHandler handles the base (header and footer) of the page
func (env *Env) baseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Parse template
	t, err := template.ParseFiles("./templates/base.html")

	if err != nil {
		log.Panic("Template error: " + err.Error())
	}

	t.ExecuteTemplate(w, "base", nil)
}

// indexHandler handles the index page, which shows last row of the data
func (env *Env) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	row := LeafDataRow{}
	row.retrieveLastRow(env)

	// Parse template
	t, err := template.ParseFiles("./templates/index.html")

	if err != nil {
		log.Panic("Template error: " + err.Error())
	}

	p := BasePage{
		DataRow:     row,
		GMapsAPIKey: os.Getenv("gmaps_apikey"),
	}

	t.Execute(w, p)
}

// Trip represents a trip from database
type Trip struct {
	ID        int64
	Timestamp time.Time
}

// tripsHandler handles the trips page
func (env *Env) tripsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Check if ID is passed
	id := r.URL.Query().Get("id")

	if id != "" {
		// ID passed, show a single trip

		// Prepare statement
		stmt, err := env.db.Prepare("SELECT * FROM data WHERE Trip = ?")

		if err != nil {
			log.Panic("Database error: " + err.Error())
		}

		rows, err := stmt.Query(id)

		if err != nil {
			log.Panic("Database error: " + err.Error())
		}

		dataRows := []LeafDataRow{}

		for rows.Next() {
			dataRow := LeafDataRow{}

			err := rows.Scan(&dataRow.ID, &dataRow.Timestamp, &dataRow.DevBat, &dataRow.Gids, &dataRow.Lat, &dataRow.Long, &dataRow.Elv, &dataRow.Seq, &dataRow.Trip, &dataRow.Odo, &dataRow.SOC, &dataRow.AHr, &dataRow.BatTemp, &dataRow.Amb, &dataRow.Wpr, &dataRow.PlugState, &dataRow.ChrgMode, &dataRow.ChrgPwr, &dataRow.VIN, &dataRow.PwrSw, &dataRow.Tunits, &dataRow.RPM, &dataRow.SOH)

			if err != nil {
				log.Panic("Database error: " + err.Error())
			}

			dataRow.OdoMi = dataRow.Odo / 1.609

			dataRows = append(dataRows, dataRow)
		}

		// Parse template
		t, err := template.ParseFiles("./templates/trip.html")

		if err != nil {
			log.Panic("Template error: " + err.Error())
		}

		// Display template with data
		t.Execute(w, struct {
			ID       string
			DataRows []LeafDataRow
		}{
			id,
			dataRows,
		})

	} else {
		// ID not passed, show list of trips

		// Prepare a statement
		rows, err := env.db.Query("SELECT Trip, MIN(time) FROM data GROUP BY Trip")

		if err != nil {
			log.Panic("Database error: " + err.Error())
		}

		// Read all trips and save them into an array
		trips := []Trip{}
		for rows.Next() {
			var id int64
			var timestamp time.Time
			var trip Trip
			rows.Scan(&id, &timestamp)
			trip.ID = id
			trip.Timestamp = timestamp

			trips = append(trips, trip)

		}

		// Parse template
		t, err := template.ParseFiles("./templates/trips.html")

		if err != nil {
			log.Panic("Template error: " + err.Error())
		}

		t.Execute(w, trips)
	}
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
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", os.Getenv("db_user"), os.Getenv("db_pass"), os.Getenv("db_host"), os.Getenv("db_schema"))

	db, err := sql.Open(os.Getenv("db_type"), dbDSN)
	if err != nil {
		log.Panic(err)
	}

	env := &Env{db: db}

	defer env.db.Close()

	// Set up web server
	// This prevents "http: Accept error: accept tcp [::]:....: accept4: too many open files; retrying in ..." errors
	var server *http.Server

	if os.Getenv("use_ssl") == "1" {
		server = &http.Server{
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 5 * time.Second,
			Addr:         ":" + os.Getenv("http_port"),
		}
	} else {
		server = &http.Server{
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 5 * time.Second,
			Addr:         ":" + os.Getenv("http_port"),
		}
	}

	// Function handlers
	http.HandleFunc("/update", env.updateHandler)
	http.HandleFunc("/trips/", env.tripsHandler)
	http.HandleFunc("/index/", env.indexHandler)
	http.HandleFunc("/", env.baseHandler)

	// Handle static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the webserver
	if os.Getenv("use_ssl") == "1" {
		server.ListenAndServeTLS("./certs/server.crt", "./certs/server.key")
	} else {
		server.ListenAndServe()
	}

}
