package api

import (
    "encoding/json"
    "net/http"
    sqldb "HighFrequencyDNSChecker/components/db"
	// log "HighFrequencyDNSChecker/components/log"
)

// Handler for the API endpoint
func ConfigAPIHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
        return
    }

    config, err := sqldb.GetConfgurations(sqldb.AppDB) // Implement this function in your sqldb package
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Serialize the configuration to JSON
    jsonResponse, err := json.Marshal(config)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Set Content-Type header and write the JSON response
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonResponse)
}
