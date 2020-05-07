package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func updateLeafStatusCom(wg *sync.WaitGroup, leafstatusURL string) {
	defer wg.Done()

	netClient := &http.Client{
		Timeout: time.Second * 10, // Prevents default golang http client to cause issues
	}

	_, err := netClient.Get(leafstatusURL)

	if err != nil {
		log.Panic(err)
	}
}

// updateHandler handles the /update part of the webserver
func (env *Env) updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Reroute data to leaf-status.com with credentials
	leafstatusURL := strings.Replace(r.URL.RawQuery, "user=", "user="+url.QueryEscape(os.Getenv("leafstatus_user")), 1)
	leafstatusURL = strings.Replace(leafstatusURL, "pass=", "pass="+url.QueryEscape(os.Getenv("leafstatus_pass")), 1)
	leafstatusURL = "https://leaf-status.com/api/vehicle/update?" + leafstatusURL
	var wg sync.WaitGroup
	wg.Add(1)
	updateLeafStatusCom(&wg, leafstatusURL)

	// Prepare a statement
	stmt, err := env.db.Prepare(`INSERT INTO data VALUES(NULL,NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		log.Panic(err)
		w.Write([]byte(`"status":"1"`))
		wg.Wait()
		return
	}

	// Execute query
	_, err = stmt.Exec(r.FormValue("DevBat"), r.FormValue("Gids"), r.FormValue("Lat"), r.FormValue("Long"), r.FormValue("Elv"), r.FormValue("Seq"), r.FormValue("Trip"), r.FormValue("Odo"),
		r.FormValue("SOC"), r.FormValue("AHr"), r.FormValue("BatTemp"), r.FormValue("Amb"), r.FormValue("Wpr"), r.FormValue("PlugState"), r.FormValue("ChgrMode"), r.FormValue("ChrgPwr"),
		r.FormValue("VIN"), r.FormValue("PwrSw"), r.FormValue("Tunits"), r.FormValue("RPM"), r.FormValue("SOH"))

	if err != nil {
		log.Panic(err)
		w.Write([]byte(`"status":"1"`))
		wg.Wait()
		return
	}

	// This sends feedback to LeafSpy that operation was successful
	w.Write([]byte(`"status":"0"`))

	// Wait for leaf-status.com goroutine
	wg.Wait()
}

// baseHandler handles the base (header and footer) of the page
func (env *Env) baseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Parse template
	t, err := template.ParseFiles("./templates/base.html")

	if err != nil {
		log.Panic("Template error: " + err.Error())
	}

	t.Execute(w, struct {
		GMapsAPIKey string
	}{
		GMapsAPIKey: os.Getenv("gmaps_apikey"),
	})
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

	t.Execute(w, struct {
		DataRow     LeafDataRow
		GMapsAPIKey string
	}{
		DataRow:     row,
		GMapsAPIKey: os.Getenv("gmaps_apikey"),
	})
}

// tripsHandler handles the trips page
func (env *Env) tripsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Check if ID is passed
	id := r.URL.Query().Get("id")

	if id != "" {
		// ID passed, show a single trip

		// Prepare statement
		stmt, err := env.db.Prepare("SELECT * FROM data WHERE Trip = ? ORDER BY time ASC")

		if err != nil {
			log.Panic("Database error: " + err.Error())
		}

		rows, err := stmt.Query(id)

		if err != nil {
			log.Panic("Database error: " + err.Error())
		}

		dataRows := []LeafDataRow{}
		i := 0
		for rows.Next() {
			dataRow := LeafDataRow{}

			err := rows.Scan(&dataRow.ID, &dataRow.Timestamp, &dataRow.DevBat, &dataRow.Gids, &dataRow.Lat, &dataRow.Long, &dataRow.Elv, &dataRow.Seq, &dataRow.Trip, &dataRow.Odo, &dataRow.SOC, &dataRow.AHr, &dataRow.BatTemp, &dataRow.Amb, &dataRow.Wpr, &dataRow.PlugState, &dataRow.ChrgMode, &dataRow.ChrgPwr, &dataRow.VIN, &dataRow.PwrSw, &dataRow.Tunits, &dataRow.RPM, &dataRow.SOH)

			if err != nil {
				log.Panic("Database error: " + err.Error())
			}

			dataRow.OdoMi = dataRow.Odo / 1.609
			i++

			dataRows = append(dataRows, dataRow)
		}

		// Parse template
		t, err := template.ParseFiles("./templates/trip.html")

		if err != nil {
			log.Panic("Template error: " + err.Error())
		}

		// Display template with data
		t.Execute(w, struct {
			ID           string
			DataRows     []LeafDataRow
			GMapsAPIKey  string
			TripStart    time.Time
			TripEnd      time.Time
			TripDuration time.Duration
		}{
			id,
			dataRows,
			os.Getenv("gmaps_apikey"),
			dataRows[0].Timestamp,
			dataRows[len(dataRows)-1].Timestamp,
			dataRows[len(dataRows)-1].Timestamp.Sub(dataRows[0].Timestamp),
		})

	} else {
		// ID not passed, show list of trips
		q := ""
		order := ""

		if r.URL.Query().Get("order") == "asc" {
			q = "SELECT Trip, MIN(time) FROM data GROUP BY Trip ORDER BY MIN(time) ASC"
			order = "desc"
		} else {
			q = "SELECT Trip, MIN(time) FROM data GROUP BY Trip ORDER BY MIN(time) DESC"
			order = "asc"
		}

		// Prepare a statement
		rows, err := env.db.Query(q)

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

		t.Execute(w, struct {
			Trips []Trip
			Order string
		}{
			trips,
			order,
		})
	}
}
