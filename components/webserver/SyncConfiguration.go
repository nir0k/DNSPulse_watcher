package webserver

import (
	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/tools"

	"html/template"
	"net/http"
)

type TemplateData struct {
    Members  []datastore.SyncMemberStateStruct
    IsEnable bool
}

func memberConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
    funcMap := template.FuncMap{
        "formatTime": tools.FormatUnixTime,
    }
    // Load the SyncConfiguration data
    isSyncEnabled := datastore.GetConfig().Sync.SyncEnabled
    data := TemplateData{
        Members: datastore.GetSyncMembersState(),
        IsEnable: isSyncEnabled,
    }
    member := datastore.SyncMemberStateStruct{
        State: datastore.StateStruct{
            Configuration: datastore.GetConfHash(),
            Csv: datastore.GetCSVHash(),
        },
        LastCheckDate: 0,
        Hostname: datastore.GetConfig().General.Hostname,
        Port: datastore.GetConfig().WebServer.Port,
        IsLocal: true,
        
    }
    data.Members = append(data.Members, member)
    
    tmpl, err := template.New("SyncConfiguration.html").Funcs(funcMap).ParseFS(tmplFS, "html/SyncConfiguration.html", "html/navbar.html", "html/styles.html")
    if err != nil {
        http.Error(w, "Failed to parse template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Execute the template with the SyncConfiguration data
    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
    }
}