package webserver

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	"embed"
	"html/template"
	"net/http"
	// Import other packages as needed
)

//go:embed html/*.html
var tmplFS embed.FS

func memberConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
    funcMap := template.FuncMap{
        "formatTime": FormatUnixTime,
    }
    // Load the SyncConfiguration data
    syncConfig, err := sqldb.GetSyncConfig(sqldb.AppDB)
    if err != nil {
        http.Error(w, "Failed to get sync configurations: "+err.Error(), http.StatusInternalServerError)
        return
    }
    // Parse the HTML template along with navbar and styles
    tmpl, err := template.New("SyncConfiguration.html").Funcs(funcMap).ParseFS(tmplFS, "html/SyncConfiguration.html", "html/navbar.html", "html/styles.html")
    if err != nil {
        http.Error(w, "Failed to parse template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Execute the template with the SyncConfiguration data
    err = tmpl.Execute(w, syncConfig)
    if err != nil {
        http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
    }
}