package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
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
	Lat       float64   `db:"Lat"`
	Long      float64   `db:"Long"`
	Elv       int32     `db:"Elv"`
	Seq       int32     `db:"Seq"`
	Trip      int32     `db:"Trip"`
	Odo       float64   `db:"odo"`
	SOC       float64   `db:"SOC"`
	AHr       float64   `db:"AHr"`
	BatTemp   float64   `db:"BatTemp"`
	Amb       float64   `db:"Amb"`
	Wpr       int8      `db:"Wpr"`
	PlugState int8      `db:"PlugState"`
	ChrgMode  int8      `db:"ChrgMode"`
	ChrgPwr   int32     `db:"ChrgPwr"`
	VIN       string    `db:"VIN"`
	PwrSw     int8      `db:"PwrSw"`
	Tunits    string    `db:"Tunits"`
	RPM       int32     `db:"RPM"`
	SOH       float64   `db:"SOH"`
	OdoMi     float64
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

// Trip represents a trip from database
type Trip struct {
	ID        int64
	Timestamp time.Time
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
	// This should prevent "http: Accept error: accept tcp [::]:....: accept4: too many open files; retrying in ..." errors
	var server = &http.Server{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 5 * time.Second,
		Addr:         ":" + os.Getenv("http_port"),
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
