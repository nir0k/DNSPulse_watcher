package webserver

import (

	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/logger"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

func configHandler(w http.ResponseWriter, r *http.Request) {
	config := datastore.GetConfigCopy()
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
					.help-button {
						position: fixed;
						bottom: 20px;
						right: 20px;
						background-color: #007bff; /* Blue color, change as needed */
						color: white;
						border-radius: 50%; /* Circular shape */
						width: 50px; /* Size of the button */
						height: 50px;
						text-align: center;
						line-height: 50px; /* Center the '?' vertically */
						font-size: 30px; /* Size of the '?' */
						text-decoration: none; /* Remove underline from link */
						box-shadow: 2px 2px 5px rgba(0, 0, 0, 0.2); /* Optional: adds a shadow */
						z-index: 1000; /* Ensure it's above other elements */
					}
					.help-button:hover {
						background-color: #0056b3; /* Darker blue on hover */
						color: white;
						text-decoration: none; /* Keep underline off on hover */
					}
                </style>
            </head>
            <body>
                <div class="navbar">
					<a href="/">Home</a>
					<a href="/csv">Edit CSV File</a>
					<a href="/config">Show/Edit Configuration</a>
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
			fmt.Fprintf(w, `<form action="/config" method="post">`)
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
			function addSyncMemberRow() {
				var table = document.getElementById("syncMembersTable");
				var row = table.insertRow(-1); // Inserts a row at the end of the table
				var cell1 = row.insertCell(0);
				var cell2 = row.insertCell(1);
				cell1.innerHTML = '<input type="text" name="syncMemberHostname[]">';
				cell2.innerHTML = '<input type="number" name="syncMemberPort[]"> <button onclick="deleteRow(this)">Delete</button>';
			}
			
			function deleteRow(btn) {
				var row = btn.parentNode.parentNode;
				row.parentNode.removeChild(row);
			}
		</script>
		`)
		fmt.Fprintln(w, `<a href="/docs" class="help-button" target="_blank" rel="noopener noreferrer">?</a>`)
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
		fmt.Printf("Group: %s", group)
		switch group {
		case "General":
			updateInterval, err := strconv.Atoi(r.FormValue("ConfigCheckInterval"))
			if err != nil {
				http.Error(w, "UpdateInterval must be an integer", http.StatusBadRequest)
				return
			}
				
			newGenConf := datastore.GeneralConfigStruct{
				ConfigCheckInterval: updateInterval,
				Location: r.FormValue("Location"),
				SecurityZone: r.FormValue("SecurityZone"),
				
			}
			fmt.Println("newGenConf: ", newGenConf)
			err = datastore.UpdateGeneralConfig(newGenConf)
			if err != nil {
				http.Error(w, "Failed to update update General section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update General section into configuration")
				return
			}
		
		case "AppLogger":
			fmt.Println("newLogConf: ")
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
			newLogConf := datastore.LogAppConfigStruct {
				Path: r.FormValue("Path"),
				MinSeverity: r.FormValue("MinSeverity"),
				MaxAge: maxAge,
				MaxSize: maxSize,
				MaxFiles: maxFiles,
			}
			fmt.Println("newLogConf: ", newLogConf)
			err = datastore.UpdateAppLoggerConfig(newLogConf)
			if err != nil {
				http.Error(w, "Failed to update update Log section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update Log section into configuration")
				return
			}

		case "AuditLogger": 
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
			newAuditConf := datastore.LogAuditConfigStruct {
				Path: r.FormValue("Path"),
				MinSeverity: r.FormValue("MinSeverity"),
				MaxAge: maxAge,
				MaxSize: maxSize,
				MaxFiles: maxFiles,
			}
			fmt.Println("newAuditConf: ", newAuditConf)
			err = datastore.UpdateAuditLoggerConfig(newAuditConf)
			if err != nil {
				http.Error(w, "Failed to update update Audit section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update Audit section into configuration")
				return
			}

		case "WebServer":
			sslIsEnable := false
			if _, ok := r.PostForm["SSLEnabled"]; ok {
				sslIsEnable = true
			}
			sesionTimeout, err := strconv.Atoi(r.FormValue("SesionTimeout"))
			if err != nil {
				http.Error(w, "SesionTimeout must be an integer", http.StatusBadRequest)
				return
			}
			port, _ := strconv.Atoi(r.FormValue("Port"))
			newWebConf := datastore.WebServerConfigStruct{
				Port: port,
				ListenAddress: r.FormValue("ListenAddress"),
				SSLEnabled: sslIsEnable,
				SSLCertPath: r.FormValue("SSLCertPath"),
				SSLKeyPath: r.FormValue("SSLKeyPath"),
				SesionTimeout: sesionTimeout,
				Username: r.FormValue("Username"),
				Password: r.FormValue("Password"),
			}
			fmt.Println("newWebConf: ", newWebConf)
			err = datastore.UpdateWebServerConfig(newWebConf)
			if err != nil {
				http.Error(w, "Failed to update update WebServer section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update WebServer section into configuration")
				return
			}

		case "Sync":
			isEnable := false
			if _, ok := r.PostForm["SyncEnabled"]; ok {
				isEnable = true
			}
			newSyncConf := datastore.SyncConfigStruct{
				SyncEnabled: isEnable,
				Token: r.FormValue("Token"),
			}
			fmt.Println("newSyncConf: ", newSyncConf)
			err := datastore.UpdateSyncConfig(newSyncConf)
			if err != nil {
				http.Error(w, "Failed to update update Sync section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update Sync section into configuration")
				return
			}

		case "SyncMembers":
			hostnames, ports := r.Form["syncMemberHostname[]"], r.Form["syncMemberPort[]"]
            var newSyncMembers []datastore.SyncMembersStruct
            for i, hostname := range hostnames {
                if i < len(ports) {
                    port, err := strconv.Atoi(ports[i])
                    if err != nil {
                        logger.Logger.Errorf("Invalid port: %v", err)
                        continue
                    }
                    newSyncMembers = append(newSyncMembers, datastore.SyncMembersStruct{Host: hostname, Port: port})
                }
            }
			fmt.Println("newSyncConf: ", newSyncMembers)
			err := datastore.UpdateSyncMembersConfig(newSyncMembers)
			if err != nil {
				http.Error(w, "Failed to update update Sync section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update Sync section into configuration")
				return
			}
		
		
		case "Prometheus":
			auth := false
			if _, ok := r.PostForm["AuthEnabled"]; ok {
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
			newPromLabelConf := datastore.PrometheusLabelConfigStruct{
				Opcode:             r.FormValue("Opcode") == "on",
				Authoritative:      r.FormValue("Authoritative") == "on",
				Truncated:          r.FormValue("Truncated") == "on",
				Rcode:              r.FormValue("Rcode") == "on",
				RecursionDesired:   r.FormValue("RecursionDesired") == "on",
				RecursionAvailable: r.FormValue("RecursionAvailable") == "on",
				AuthenticatedData:  r.FormValue("AuthenticatedData") == "on",
				CheckingDisabled:   r.FormValue("CheckingDisabled") == "on",
				PollingRate:        r.FormValue("PollingRate") == "on",
				Recursion:          r.FormValue("Recursion") == "on",
			}
			newPromConf := datastore.PrometheusConfStruct{
				URL: r.FormValue("URL"),
				MetricName: r.FormValue("MetricName"),
				AuthEnabled: auth,
				Username: r.FormValue("Username"),
				Password: r.FormValue("Password"),
				RetriesCount: retriesCount,
				BufferSize: bufferSize,
				Labels: newPromLabelConf,
			}
			fmt.Println("newPromConf: ", newPromConf)
			err = datastore.UpdatePrometheusConfig(newPromConf)
			if err != nil {
				http.Error(w, "Failed to update update Prometheus section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update Prometheus section into configuration")
				return
			}

		case "Polling":
			pullTimeout, err := strconv.Atoi(r.FormValue("PollTimeout"))
			if err != nil {
				http.Error(w, "BufferSize must be an integer", http.StatusBadRequest)
				return
			}
			newResolvConf := datastore.PollingConfigStruct{
				Path: r.FormValue("Path"),
				PollTimeout: pullTimeout,
				Delimeter: r.FormValue("Delimeter"),
				ExtraDelimeter: r.FormValue("ExtraDelimeter"),
			}
			fmt.Println("newResolvConf: ", newResolvConf)
			err = datastore.UpdatePollingConfig(newResolvConf)
			if err != nil {
				http.Error(w, "Failed to update update Polling section into configuration", http.StatusBadRequest)
				logger.Logger.Errorf("Error to update Polling section into configuration")
				return
			}
		}
		http.Redirect(w, r, "/config", http.StatusSeeOther)
	}
}

func displayReadOnlyConfig(w http.ResponseWriter, fieldValue interface{}) {
    v := reflect.ValueOf(fieldValue)
    t := v.Type()

    // Check if fieldValue is a slice
    if t.Kind() == reflect.Slice {
        for i := 0; i < v.Len(); i++ {
            item := v.Index(i).Interface()
            displaySingleStruct(w, item) // displaySingleStruct handles a single struct
        }
        return
    }

    displaySingleStruct(w, fieldValue)
}

func displaySingleStruct(w http.ResponseWriter, fieldValue interface{}) {
    v := reflect.ValueOf(fieldValue)
    t := v.Type()

    for i := 0; i < v.NumField(); i++ {
        fieldName := t.Field(i).Name
        fieldVal := v.Field(i)

        if fieldName == "LastCheck" || fieldName == "LastUpdate" || fieldName == "Hash" {
            continue
        }

        if t.Field(i).Type == reflect.TypeOf(datastore.PrometheusLabelConfigStruct{}) {
            // Special handling for PrometheusLabelConfigStruct
            fmt.Fprintf(w, "<tr><td colspan='2'>%s</td></tr>", fieldName)
            displayPrometheusLabels(w, fieldVal.Interface().(datastore.PrometheusLabelConfigStruct))
        } else if fieldVal.Kind() == reflect.Bool {
            // Convert boolean value to a more readable format
            displayValue := "Disable"
            if fieldVal.Bool() {
                displayValue = "Enable"
            }
            fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>", fieldName, displayValue)
        } else {
            // For other types of fields, display the value as is
            fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td></tr>", fieldName, fieldVal)
        }
    }
}


func displayPrometheusLabels(w http.ResponseWriter, labels datastore.PrometheusLabelConfigStruct) {
    labelVal := reflect.ValueOf(labels)
    labelType := labelVal.Type()

    for i := 0; i < labelVal.NumField(); i++ {
        labelFieldName := labelType.Field(i).Name
        labelFieldVal := labelVal.Field(i)

        displayValue := "Disable"
        if labelFieldVal.Bool() {
            displayValue = "Enable"
        }
        fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>", labelFieldName, displayValue)
    }
}

func createEditableFields(w http.ResponseWriter, fieldValue interface{}) {
    v := reflect.ValueOf(fieldValue)
    t := v.Type()

    if t.Kind() == reflect.Slice {
        if _, ok := fieldValue.([]datastore.SyncMembersStruct); ok {
			fmt.Fprintln(w, `<table id="syncMembersTable" class="config-table">`)
			for _, member := range fieldValue.([]datastore.SyncMembersStruct) {
				fmt.Fprintf(w, `<tr>
					<td><input type="text" name="syncMemberHostname[]" value="%s"></td>
					<td><input type="number" name="syncMemberPort[]" value="%d"></td>
					<td><button type="button" onclick="deleteRow(this)">Delete</button></td>
				</tr>`, member.Host, member.Port)
			}
			fmt.Fprintln(w, `</table>`)
			fmt.Fprintln(w, `<button type="button" onclick="addSyncMemberRow()">Add Sync Member</button>`)
        } else {
            for i := 0; i < v.Len(); i++ {
                item := v.Index(i).Interface()
                createEditableSingleStruct(w, item)
            }
        }
    } else {
        createEditableSingleStruct(w, fieldValue)
    }
}

func createEditableSingleStruct(w http.ResponseWriter, fieldValue interface{}) {
    v := reflect.ValueOf(fieldValue)
    t := v.Type()

    for i := 0; i < v.NumField(); i++ {
        fieldName := t.Field(i).Name
        fieldVal := v.Field(i)

        if fieldName == "LastCheck" || fieldName == "LastUpdate" || 
           fieldName == "Hash" || fieldName == "Hostname" || fieldName == "IPAddress" {
            continue
        }

        if t.Field(i).Type == reflect.TypeOf(datastore.PrometheusLabelConfigStruct{}) {
            // Special handling for PrometheusLabelConfigStruct
            fmt.Fprintf(w, "<tr><td colspan='2'>%s</td></tr>", fieldName)
            createEditablePrometheusLabels(w, fieldVal.Interface().(datastore.PrometheusLabelConfigStruct))
        } else {
            // For other types of fields, use appropriate input fields
            createFieldInput(w, fieldName, fieldVal)
        }
    }
}

func createEditablePrometheusLabels(w http.ResponseWriter, labels datastore.PrometheusLabelConfigStruct) {
    labelVal := reflect.ValueOf(labels)
    labelType := labelVal.Type()

    for i := 0; i < labelVal.NumField(); i++ {
        labelFieldName := labelType.Field(i).Name
        labelFieldVal := labelVal.Field(i)

        checked := ""
        if labelFieldVal.Bool() {
            checked = "checked"
        }
        fmt.Fprintf(w, `<tr><td>%s</td><td><input type="checkbox" name="%s" %s></td></tr>`, labelFieldName, labelFieldName, checked)
    }
}

func createFieldInput(w http.ResponseWriter, fieldName string, fieldVal reflect.Value) {
    switch fieldVal.Kind() {
    case reflect.Bool:
        checked := ""
        if fieldVal.Bool() {
            checked = "checked"
        }
        fmt.Fprintf(w, `<tr><td>%s</td><td><input type="checkbox" name="%s" %s></td></tr>`, fieldName, fieldName, checked)
    case reflect.Int:
        fmt.Fprintf(w, `<tr><td>%s</td><td><input type="number" name="%s" value="%v"></td></tr>`, fieldName, fieldName, fieldVal)
    default:
        // For other types of fields, use text input
        fmt.Fprintf(w, `<tr><td>%s</td><td><input type="text" name="%s" value="%v"></td></tr>`, fieldName, fieldName, fieldVal)
    }
}
