package api

import (
	"HighFrequencyDNSChecker/components/datastore"
	"encoding/json"
	"net/http"
)

func ConfStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
        http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
        return
    }
	var state datastore.StateStruct

	state.Configuration = datastore.GetConfHash()
	state.Csv = datastore.GetCSVHash()
	
	w.Header().Set("Content-Type", "application/json")

	prettyJSON, err := json.MarshalIndent(state, "", "    ")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	w.Write(prettyJSON)
}