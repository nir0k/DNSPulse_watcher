package webserver

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"
	// "time"

	// "database/sql"
	"fmt"
	"net/http"
	"strconv"

	// "path/filepath"
	"reflect"
)

func configHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		config sqldb.Config
	)
	config, err = sqldb.GetConfgurations(sqldb.AppDB)
		if err != nil {
			http.Error(w, "Failed to get configurations", http.StatusNoContent)
			log.AppLog.Error("Failed to get  configurations, error: ", err)
		}
	w.Header().Set("Content-Type", "text/html")

	switch r.Method {
	case "GET":
		fmt.Fprintln(w, `
            <html>
            <head>
                <title>Show/Edit Configuration</title>
                <style>
                    body {
                        margin: 0;
                        padding: 0;
                        background-color: #f0f0f0;
                    }
                    .navbar {
                        position: fixed;
                        top: 0;
                        width: 100%;
                        background-color: #333;
                        overflow: hidden;
                    }
                    .navbar a {
                        float: left;
                        display: block;
                        color: white;
                        text-align: center;
                        padding: 14px 16px;
                        text-decoration: none;
                    }
                    .navbar a.logout {
                        float: right;
                    }
                    .main-content {
                        margin-top: 60px; /* Adjust this margin to avoid overlapping with the navbar */
                    }
                    .config-container {
                        padding: 20px;
                        background-color: #fff;
                        margin: 20px;
                        border-radius: 5px;
                    }
                    .group-header {
                        background-color: #333;
                        color: white;
                        padding: 10px;
                        margin: 10px 0;
                        border-radius: 5px;
                    }
                    .group-content {
                        padding: 10px;
                        border: 1px solid #ddd;
                        border-radius: 5px;
                        background-color: #fff;
                        margin: 10px 0;
                    }
                    .config-table {
                        width: 100%;
                        border-collapse: collapse;
                    }
                    .config-table tr:nth-child(odd) {
                        background-color: #f2f2f2;
                    }
                    .config-table tr:nth-child(even) {
                        background-color: #ffffff;
                    }
                    .config-table th, .config-table td {
                        border: 1px solid #ddd;
                        padding: 8px;
                    }
                </style>
            </head>
            <body>
                <div class="navbar">
					<a href="/">Home</a>
					<a href="/csv">Edit CSV File</a>
					<a href="/env">Show/Edit Configuration</a>
					<a href="/members">Sync Status</a>
					<a href="/logs">LogView</a>
					<a href="/logout" class="logout">Logout</a>
                </div>
                <div class="main-content">
                    <h2>Show/Edit Configuration</h2>
        `)

		v := reflect.ValueOf(&config).Elem()
		for i := 0; i < v.NumField(); i++ {
			fieldName := v.Type().Field(i).Name
			fieldValue := v.Field(i).Interface()
			groupID := fmt.Sprintf("group%d", i)
	
			fmt.Fprintf(w, `<div class="config-container">`)
	
			// Group header
			fmt.Fprintf(w, `
				<div class="group-header" onclick="toggleGroup('%s')">
					%s
				</div>
			`, groupID, fieldName)
	
			// Group content (initially hidden)
			fmt.Fprintf(w, `<div id="%s-content" class="group-content" style="display: none;">`, groupID)
			fmt.Fprintln(w, `<table class="config-table">`)
			displayReadOnlyConfig(w, fieldValue)
			fmt.Fprintln(w, "</table>")
			fmt.Fprintf(w, `<button type="button" onclick="toggleEdit('%s')">Edit</button>`, groupID)
			fmt.Fprintf(w, `</div>`)
	
			// Editable form (hidden initially)
			fmt.Fprintf(w, `<div id="%s-edit" class="group-content" style="display: none;">`, groupID)
			fmt.Fprintf(w, `<form action="/env" method="post">`)
			fmt.Fprintf(w, `<input type="hidden" name="group" value="%s">`, fieldName)
			fmt.Fprintln(w, `<table class="config-table">`)
			createEditableFields(w, fieldValue)
			fmt.Fprintln(w, "</table>")
			fmt.Fprintf(w, `<button type="submit">Save</button> <button type="button" onclick="toggleEdit('%s')">Cancel</button>`, groupID)
			fmt.Fprintf(w, "</form>")
			fmt.Fprintf(w, `</div>`)
	
			fmt.Fprintf(w, `</div>`) // End of config-container
		}
	
		// JavaScript for collapsible groups and edit toggle
		fmt.Fprintln(w, `
		<script>
			function toggleGroup(groupID) {
				var content = document.getElementById(groupID + '-content');
				var edit = document.getElementById(groupID + '-edit');
				if (content.style.display === 'block') {
					content.style.display = 'none';
				} else {
					content.style.display = 'block';
					edit.style.display = 'none'; // Hide the edit view when collapsing the group
				}
			}
		
			function toggleEdit(groupID) {
				var content = document.getElementById(groupID + '-content');
				var edit = document.getElementById(groupID + '-edit');
				if (edit.style.display === 'none') {
					edit.style.display = 'block';
					content.style.display = 'none'; // Hide the content view when showing the edit view
				} else {
					edit.style.display = 'none';
					content.style.display = 'block'; // Show the content view when hiding the edit view
				}
			}
		</script>
		`)

		fmt.Fprintln(w, `</div></body></html>`)

	case "POST":
		// Parse the form data
		if err := r.ParseForm(); err != nil {
			// Handle error
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Identify which group is being updated
		group := r.FormValue("group")
		switch group {
		case "General":
			syncValue := false
			if _, ok := r.PostForm["Sync"]; ok {
				syncValue = true
			}
			updateInterval, err := strconv.Atoi(r.FormValue("UpdateInterval"))
			if err != nil {
				http.Error(w, "UpdateInterval must be an integer", http.StatusBadRequest)
				return
			}

			newGenConf := sqldb.MainConfiguration{
				ConfPath: r.FormValue("ConfPath"),
				Sync: syncValue,
				UpdateInterval: updateInterval,
			}
			fmt.Println("newGenConf: ", newGenConf)
			err = sqldb.UpdateMainConfEditableFields(sqldb.AppDB, newGenConf)
			if err != nil {
				http.Error(w, "Failed to update General configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update General configurations, error: ", err)
				return
			}
		
		case "Log":
			maxAge, err := strconv.Atoi(r.FormValue("MaxAge"))
			if err != nil {
				http.Error(w, "MaxAge must be an integer", http.StatusBadRequest)
				return
			}
			maxSize, err := strconv.Atoi(r.FormValue("MaxSize"))
			if err != nil {
				http.Error(w, "MaxSize must be an integer", http.StatusBadRequest)
				return
			}
			maxFiles, err := strconv.Atoi(r.FormValue("MaxFiles"))
			if err != nil {
				http.Error(w, "MaxFiles must be an integer", http.StatusBadRequest)
				return
			}
			newLogConf := sqldb.LogConfiguration {
				Path: r.FormValue("Path"),
				MinSeverity: r.FormValue("MinSeverity"),
				MaxAge: maxAge,
				MaxSize: maxSize,
				MaxFiles: maxFiles,
			}
			fmt.Println("newLogConf: ", newLogConf)
			err = sqldb.UpdateLogConfEditableFields(sqldb.AppDB, newLogConf, 0)
			if err != nil {
				http.Error(w, "Failed to update Log configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update Log configurations, error: ", err)
				return
			}

		case "Audit": 
			maxAge, err := strconv.Atoi(r.FormValue("MaxAge"))
			if err != nil {
				http.Error(w, "MaxAge must be an integer", http.StatusBadRequest)
				return
			}
			maxSize, err := strconv.Atoi(r.FormValue("MaxSize"))
			if err != nil {
				http.Error(w, "MaxSize must be an integer", http.StatusBadRequest)
				return
			}
			maxFiles, err := strconv.Atoi(r.FormValue("MaxFiles"))
			if err != nil {
				http.Error(w, "MaxFiles must be an integer", http.StatusBadRequest)
				return
			}
			newAuditConf := sqldb.LogConfiguration {
				Path: r.FormValue("Path"),
				MinSeverity: r.FormValue("MinSeverity"),
				MaxAge: maxAge,
				MaxSize: maxSize,
				MaxFiles: maxFiles,
			}
			fmt.Println("newAuditConf: ", newAuditConf)
			err = sqldb.UpdateLogConfEditableFields(sqldb.AppDB, newAuditConf, 1)
			if err != nil {
				http.Error(w, "Failed to update Audit Log configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update Audit Log configurations, error: ", err)
				return
			}

		case "WebServer":
			sslIsEnable := false
			if _, ok := r.PostForm["SslIsEnable"]; ok {
				sslIsEnable = true
			}
			sesionTimeout, err := strconv.Atoi(r.FormValue("SesionTimeout"))
			if err != nil {
				http.Error(w, "SesionTimeout must be an integer", http.StatusBadRequest)
				return
			}

			newWebConf := sqldb.WebServerConfiguration{
				Port: r.FormValue("Port"),
				SslIsEnable: sslIsEnable,
				SslCertPath: r.FormValue("SslCertPath"),
				SslKeyPath: r.FormValue("SslKeyPath"),
				SesionTimeout: sesionTimeout,
				InitUsername: r.FormValue("InitUsername"),
				InitPassword: r.FormValue("InitPassword"),
			}
			fmt.Println("newWebConf: ", newWebConf)
			err = sqldb.UpdateWebConfEditableFields(sqldb.AppDB, newWebConf)
			if err != nil {
				http.Error(w, "Failed to update Web-server configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update Web-server configurations, error: ", err)
				return
			}

		case "Sync":
			isEnable := false
			if _, ok := r.PostForm["SslIsEnable"]; ok {
				isEnable = true
			}
			newSyncConf := sqldb.SyncConfiguration{
				IsEnable: isEnable,
			}
			fmt.Println("newSyncConf: ", newSyncConf)
			err := sqldb.UpdateSyncConfEditableFields(sqldb.AppDB, newSyncConf)
			if err != nil {
				http.Error(w, "Failed to update sync configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update sync configurations, error: ", err)
				return
			}

		case "Prometheus":
			auth := false
			if _, ok := r.PostForm["Auth"]; ok {
				auth = true
			}
			retriesCount, err := strconv.Atoi(r.FormValue("RetriesCount"))
			if err != nil {
				http.Error(w, "RetriesCount must be an integer", http.StatusBadRequest)
				return
			}
			bufferSize, err := strconv.Atoi(r.FormValue("BufferSize"))
			if err != nil {
				http.Error(w, "BufferSize must be an integer", http.StatusBadRequest)
				return
			}

			newPromConf := sqldb.PrometheusConfiguration{
				Url: r.FormValue("Url"),
				MetricName: r.FormValue("MetricName"),
				Auth: auth,
				Username: r.FormValue("Username"),
				Password: r.FormValue("Password"),
				RetriesCount: retriesCount,
				BufferSize: bufferSize,
			}
			fmt.Println("newPromConf: ", newPromConf)
			err = sqldb.UpdatePromConfEditableFields(sqldb.AppDB, newPromConf)
			if err != nil {
				http.Error(w, "Failed to update prometheus configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update prometheus configurations, error: ", err)
				return
			}

		case "PrometheusLabels":
			opcode := false
			if _, ok := r.PostForm["Opcode"]; ok {
				opcode = true
			}
			authoritative := false
			if _, ok := r.PostForm["Authoritative"]; ok {
				authoritative = true
			}
			truncated := false
			if _, ok := r.PostForm["Truncated"]; ok {
				truncated = true
			}
			rcode := false
			if _, ok := r.PostForm["Rcode"]; ok {
				rcode = true
			}
			recursionDesired := false
			if _, ok := r.PostForm["RecursionDesired"]; ok {
				recursionDesired = true
			}
			recursionAvailable := false
			if _, ok := r.PostForm["RecursionAvailable"]; ok {
				recursionAvailable = true
			}
			authenticatedData := false
			if _, ok := r.PostForm["AuthenticatedData"]; ok {
				authenticatedData = true
			}
			checkingDisabled := false
			if _, ok := r.PostForm["CheckingDisabled"]; ok {
				checkingDisabled = true
			}
			pollingRate := false
			if _, ok := r.PostForm["PollingRate"]; ok {
				pollingRate = true
			}
			recursion := false
			if _, ok := r.PostForm["Recursion"]; ok {
				recursion = true
			}
			newPromLabelConf := sqldb.PrometheusLabelConfiguration{
				Opcode: opcode,
				Authoritative: authoritative,
				Truncated: truncated,
				Rcode: rcode,
				RecursionDesired: recursionDesired,
				RecursionAvailable: recursionAvailable,
				AuthenticatedData: authenticatedData,
				CheckingDisabled: checkingDisabled,
				PollingRate: pollingRate,
				Recursion: recursion,
			}
			fmt.Println("newPromLabelConf: ", newPromLabelConf)
			err := sqldb.InsertPrometeusLabelsConfig(sqldb.AppDB, newPromLabelConf)
			if err != nil {
				http.Error(w, "Failed to update prometheus labels configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update prometheus labels configurations, error: ", err)
				return
			}

		case "Resolvers":
			pullTimeout, err := strconv.Atoi(r.FormValue("PullTimeout"))
			if err != nil {
				http.Error(w, "BufferSize must be an integer", http.StatusBadRequest)
				return
			}
			newResolvConf := sqldb.ResolversConfiguration{
				Path: r.FormValue("Path"),
				PullTimeout: pullTimeout,
				Delimeter: r.FormValue("Delimeter"),
				ExtraDelimeter: r.FormValue("ExtraDelimeter"),
			}
			fmt.Println("newResolvConf: ", newResolvConf)
			err = sqldb.UpdateResolvConfEditableFields(sqldb.AppDB, newResolvConf)
			if err != nil {
				http.Error(w, "Failed to update resolver configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update resolver configurations, error: ", err)
				return
			}

		case "Watcher":
			newWatcherConf := sqldb.WatcherConfiguration{
				Location: r.FormValue("Path"),
				SecurityZone: r.FormValue("Delimeter"),
			}
			fmt.Println("newWatcherConf: ", newWatcherConf)
			err := sqldb.UpdateWatcherConfEditableFields(sqldb.AppDB, newWatcherConf)
			if err != nil {
				http.Error(w, "Failed to update wathcer configurations", http.StatusBadRequest)
				log.AppLog.Error("Failed to update watcher configurations, error: ", err)
				return
			}
		}
		http.Redirect(w, r, "/env", http.StatusSeeOther)
		// http.Redirect(w, r, "/env?timestamp=" + strconv.FormatInt(time.Now().UnixNano(), 10), http.StatusSeeOther)


	}
}

// displayReadOnlyConfig displays the configuration fields in a read-only format
func displayReadOnlyConfig(w http.ResponseWriter, fieldValue interface{}) {
	v := reflect.ValueOf(fieldValue)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		if fieldName == "LastCheck" || fieldName == "LastUpdate" || fieldName == "Hash" || fieldName == "Members" {
            continue
		}
		fieldVal := v.Field(i).Interface()
		if fieldName == "Sync" ||
			fieldName == "SslIsEnable" || 
			fieldName == "IsEnable" || 
			fieldName == "Auth" ||
			fieldName == "Opcode" ||
			fieldName == "Authoritative" ||
			fieldName == "Truncated" ||
			fieldName == "Rcode" ||
			fieldName == "RecursionDesired" ||
			fieldName == "RecursionAvailable" ||
			fieldName == "AuthenticatedData" ||
			fieldName == "CheckingDisabled" ||
			fieldName == "PollingRate" ||
			fieldName == "Recursion" {
            // Convert the boolean value to a more readable format
            displayValue := "Disable"
            if fieldVal.(bool) { // Check if the boolean value is true
                displayValue = "Enable"
            }
            fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>", fieldName, displayValue)
        } else {
            // For other types of fields, display the value as is
            fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td></tr>", fieldName, fieldVal)
        }
	}
}

// createEditableFields generates input fields for each field in the configuration
func createEditableFields(w http.ResponseWriter, fieldValue interface{}) {
	v := reflect.ValueOf(fieldValue)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		if fieldName == "LastCheck" || 
			fieldName == "LastUpdate" || 
			fieldName == "DBname" || 
			fieldName == "Hash" || 
			fieldName == "Hostname" || 
			fieldName == "IPAddress" || 
			fieldName == "Members"  {
            continue
		}
		fieldVal := v.Field(i).Interface()
		if fieldName == "Sync" ||
			fieldName == "SslIsEnable" || 
			fieldName == "IsEnable" || 
			fieldName == "Auth" ||
			fieldName == "Opcode" ||
			fieldName == "Authoritative" ||
			fieldName == "Truncated" ||
			fieldName == "Rcode" ||
			fieldName == "RecursionDesired" ||
			fieldName == "RecursionAvailable" ||
			fieldName == "AuthenticatedData" ||
			fieldName == "CheckingDisabled" ||
			fieldName == "PollingRate" ||
			fieldName == "Recursion" {
            // Assuming 'sync' is a boolean field
            checked := ""
            if fieldVal.(bool) { // Check if the boolean value is true
                checked = "checked"
            }
            fmt.Fprintf(w, `<tr><td>%s</td><td><input type="checkbox" name="%s" %s></td></tr>`, fieldName, fieldName, checked)
        } else if 
			fieldName == "UpdateInterval" || 
			fieldName == "MaxAge" || 
			fieldName == "MaxSize" || 
			fieldName == "MaxFiles" || 
			fieldName == "SesionTimeout" || 
			fieldName == "RetriesCount" || 
			fieldName == "BufferSize" ||
			fieldName == "PullTimeout" {
			fmt.Fprintf(w, `<tr><td>%s</td><td><input type="number" name="%s" value="%v"></td></tr>`, fieldName, fieldName, fieldVal)
		}else {
            // For other types of fields, use text input
            fmt.Fprintf(w, `<tr><td>%s</td><td><input type="text" name="%s" value="%v"></td></tr>`, fieldName, fieldName, fieldVal)
        }

	

		// // Creating a text input for each field. Adjust the type of input based on the actual field data type.
		// fmt.Fprintf(w, `<tr><td>%s</td><td><input type="text" name="%s" value="%v"></td></tr>`, fieldName, fieldName, fieldVal)
	}
}
