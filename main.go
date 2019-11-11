package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Env struct {
	db *sql.DB
}

type LeafSpyData struct {
}

func (env *Env) updateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())

	stmt, err := env.db.Prepare(`INSERT INTO data VALUES(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		w.Write([]byte(`"status":"1"`))
		return
	}

	_, err = stmt.Exec(r.FormValue("DevBat"), r.FormValue("Gids"), r.FormValue("Lat"), r.FormValue("Long"), r.FormValue("Elv"), r.FormValue("Seq"), r.FormValue("Trip"), r.FormValue("Odo"),
		r.FormValue("SOC"), r.FormValue("AHr"), r.FormValue("BatTemp"), r.FormValue("Amb"), r.FormValue("Wpr"), r.FormValue("PlugState"), r.FormValue("ChgrMode"), r.FormValue("ChrgPwr"),
		r.FormValue("VIN"), r.FormValue("PwrSw"), r.FormValue("Tunits"), r.FormValue("RPM"), r.FormValue("SOH"))

	if err != nil {
		w.Write([]byte(`"status":"1"`))
		return
	}

	w.Write([]byte(`"status":"0"`))

}

func main() {
	// Initialize .env file
	err := godotenv.Load("config.env")
	if err != nil {
		log.Panic(err)
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
	http.ListenAndServe(":80", nil)
}
