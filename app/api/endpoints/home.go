package endpoints

import (
	"correlator/db"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		response []byte
		amount   int       = -1 // How much alerts will be returned
		Since    time.Time      // Return alerts since that time
		To       time.Time      // To that time
	)
	w.Header().Set("Content-Type", "application/json")

	if v := r.URL.Query().Get("amount"); v != "" {
		if a, err := strconv.Atoi(v); err == nil && a > 0 {
			amount = a
		}
	}
	if v := r.URL.Query().Get("since"); v != "" {
		layout := time.RFC3339
		if parsedTime, err := time.Parse(layout, v); err == nil && parsedTime.Before(time.Now()) {
			Since = parsedTime
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		layout := time.RFC3339
		if parsedTime, err := time.Parse(layout, v); err == nil && parsedTime.After(Since) {
			To = parsedTime
		}
	}

	alerts := []db.Alert{}
	if tx := db.DB.Order("ID DESC").Limit(amount).Find(&alerts); tx.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	var data []map[string]interface{}

	if !Since.IsZero() && !To.IsZero() {
		for _, a := range alerts {
			if a.Timestamp.After(Since) && a.Timestamp.Before(To) {
				d := map[string]interface{}{}
				json.Unmarshal(a.Text, &d)
				data = append(data, d)
			}
		}
	} else if Since.IsZero() && To.IsZero() {
		for _, a := range alerts {
			d := map[string]interface{}{}
			json.Unmarshal(a.Text, &d)
			data = append(data, d)
		}
	} else {
		if !Since.IsZero() {
			for _, a := range alerts {
				if a.Timestamp.After(Since) {
					d := map[string]interface{}{}
					json.Unmarshal(a.Text, &d)
					data = append(data, d)
				}
			}
		}
		if !To.IsZero() {
			for _, a := range alerts {
				if a.Timestamp.Before(To) {
					d := map[string]interface{}{}
					json.Unmarshal(a.Text, &d)
					data = append(data, d)
				}
			}
		}
	}

	if response, err = json.Marshal(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else {
		w.Write(response)
	}
}

func HomePage(w http.ResponseWriter, r *http.Request) {

}
