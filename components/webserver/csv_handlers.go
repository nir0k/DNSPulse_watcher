package webserver

import (
	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/logger"
	"HighFrequencyDNSChecker/components/tools"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)


func csvHandler(w http.ResponseWriter, r *http.Request) {
	conf := datastore.GetConfig().Polling
    csvFile := conf.Path
    var editingRow = -1

    switch r.Method {
    case "GET":
        file, err := os.Open(csvFile)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer file.Close()

        records, err := csv.NewReader(file).ReadAll()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        fmt.Fprintf(w, `
    <html>
    <head>
        <title>Edit CSV File</title>
        <style>
            body {
                margin: 0;
                padding: 0;
                background-color: #f0f0f0;
            }
            .navbar { 
                position: fixed;
                top: 0;
                width: 100%%;
                height: 60px;
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
            .title-container {
                display: flex;
                justify-content: center;
                align-items: center;
            }
            .navbar a.logout {
                float: right;
            }
            .main-content {
                margin-top: 60px;
            }
            tr:nth-child(odd) {
                background-color: #f2f2f2;
            }
            tr:nth-child(even) {
                background-color: #ffffff;
            }
            h1 {
                color: #333;
            }
            table {
                table-layout: auto;
                width: 100%%;
                border-collapse: collapse;
            }
            td {
                border: 1px solid black;
                padding: 8px;
                text-align: left;
                word-wrap: break-word;
                min-width: 50px;
            }
            input[type="text"] {
                width: 100%%;
                box-sizing: border-box;
            }
            th {
                background-color: lightblue !important;
                position: sticky;
                padding: 8px;
                top: 0;
                background-color: white;
                z-index: 1;
            }
            .help-button {
                position: fixed;
                bottom: 20px;
                right: 20px;
                background-color: #007bff; /* Blue color, change as needed */
                color: white;
                border-radius: 50%%; /* Circular shape */
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
        <script>
        // Your AJAX function here
        function ajaxRequest(method, url, data, callback) {
            var xhr = new XMLHttpRequest();
            xhr.open(method, url, true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.onreadystatechange = function() {
                if (xhr.readyState == 4 && xhr.status == 200) {
                    callback(xhr.responseText);
                }
            };
            xhr.send(JSON.stringify(data));
        }
        function escapeHTML(str) {
            return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
        }
    
        function unescapeHTML(str) {
            var textArea = document.createElement('textarea');
            textArea.innerHTML = str;
            return textArea.value;
        }
        // Additional JavaScript functions for delete, edit, and add
        // Function to delete a row
        function deleteRow(rowIndex, event) {
            event.preventDefault();
            ajaxRequest('POST', '/csv/delete', { index: rowIndex }, function(response) {
                // Assuming the deletion was successful
                var row = document.getElementById('row-' + rowIndex);
                if (row) {
                    row.parentNode.removeChild(row);
                }
            });
        }  
        // Function to edit a row      
        function editRow(rowIndex, event) {
            event.preventDefault();
        
            var row = document.getElementById('row-' + rowIndex);
            if (!row) return;
        
            var cells = row.getElementsByTagName('td');
            for (var i = 0; i < cells.length - 1; i++) {
                var cell = cells[i];
                var cellValue = unescapeHTML(cell.innerHTML); // Unescape the HTML here
                var input = document.createElement('input');
                input.type = 'text';
                input.value = cellValue;
                cell.innerHTML = '';
                cell.appendChild(input);
            }
        
            document.getElementById('edit-btn-' + rowIndex).style.display = 'none';
            document.getElementById('delete-btn-' + rowIndex).style.display = 'none';
            document.getElementById('save-btn-' + rowIndex).style.display = 'inline-block';
            document.getElementById('cancel-btn-' + rowIndex).style.display = 'inline-block';
        }              
        function saveRow(rowIndex, event) {
            event.preventDefault();
        
            var row = document.getElementById('row-' + rowIndex);
            if (!row) return;
        
            var newData = [];
            var inputs = row.getElementsByTagName('input');
            for (var i = 0; i < inputs.length; i++) {
                newData.push(inputs[i].value);
            }
        
            ajaxRequest('POST', '/csv/edit', { index: rowIndex, data: newData }, function(response) {
                for (var i = 0; i < newData.length; i++) {
                    row.cells[i].innerHTML = escapeHTML(newData[i]); // Escape HTML entities here
                }
        
                document.getElementById('edit-btn-' + rowIndex).style.display = 'inline-block';
                document.getElementById('delete-btn-' + rowIndex).style.display = 'inline-block';
                document.getElementById('save-btn-' + rowIndex).style.display = 'none';
                document.getElementById('cancel-btn-' + rowIndex).style.display = 'none';
            });
        }                             
        </script>    
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
            <div class="title-container">
                <h1>Edit CSV File</h1>
            </div>
    `)
        
        fmt.Fprintf(w, "<div>")
        fmt.Fprintf(w, "<a href='/csv/download'><button type='button'>Download CSV</button></a>")
        fmt.Fprintf(w, "<form method='POST' action='/csv/upload' enctype='multipart/form-data'>")
        fmt.Fprintf(w, "<input type='file' name='csvfile'>")
        fmt.Fprintf(w, "<input type='submit' value='Upload'>")
        fmt.Fprintf(w, "</form>")
        fmt.Fprintf(w, "</div>")

        fmt.Fprintf(w, "<form method='POST' action='/csv'>") 
        fmt.Fprintf(w, "<table border='1'><tr>")

        for _, header := range records[0] {
            fmt.Fprintf(w, "<th>%s</th>", header)
        }
        fmt.Fprintf(w, "<th>Action</th></tr>")

        for i, record := range records {
            if i == 0 { 
                continue
            }

            fmt.Fprintf(w, "<tr id='row-%d'>", i)
            for j, field := range record {
                escapedField := escapeHTML(field)
                fmt.Fprintf(w, "<td id='cell-%d-%d'>%s</td>", i, j, escapedField)
            }
            if i == editingRow {
                
                for _, field := range record {
                    fmt.Fprintf(w, "<td><input type='text' name='row%d[]' value='%s'></td>", i, field)
                    
                }
                fmt.Fprintf(w, "<td><button type='submit' name='save' value='%d'>Save</button>", i)
                fmt.Fprintf(w, "<a href='/csv'><button type='button'>Cancel</button></a></td>")
            } else {
                fmt.Fprintf(w, "<td style='display: flex; justify-content: space-around;'>")
                fmt.Fprintf(w, "<button id='edit-btn-%d' onclick='editRow(%d, event)'>Edit</button>", i, i)
                fmt.Fprintf(w, "<button id='delete-btn-%d' onclick='deleteRow(%d, event)'>Delete</button>", i, i)
                fmt.Fprintf(w, "<button id='save-btn-%d' onclick='saveRow(%d, event)' style='display: none;'>Save</button>", i, i)
                fmt.Fprintf(w, "<button id='cancel-btn-%d' onclick='cancelEdit(%d, event)' style='display: none;'>Cancel</button>", i, i)
                fmt.Fprintf(w, "</td>")
            }
            fmt.Fprintf(w, "</tr>")
        }

        fmt.Fprintf(w, "<tr>")
        for range records[0] {
            fmt.Fprintf(w, "<td><input type='text' name='newrow[]'></td>")
        }
        fmt.Fprintf(w, "<td><button type='submit' name='add'>Add</button></td></tr>")
        fmt.Fprintf(w, "</table>")

        
        // fmt.Fprintf(w, "</form>")
        fmt.Fprintf(w, "<form method='POST' action='/csv'>")
        fmt.Fprintf(w, "<table>")
        // Add your form fields here
        fmt.Fprintf(w, "</table>")
        fmt.Fprintf(w, "</form>")
        fmt.Fprintf(w, "</div>")
        fmt.Fprintf(w, `<a href="/docs" class="help-button" target="_blank" rel="noopener noreferrer">?</a>`)
        fmt.Fprintf(w, "</body></html>")

    case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	
		file, err := os.Open(csvFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
	
		records, err := csv.NewReader(file).ReadAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	
		
		if saveIndex, ok := r.Form["save"]; ok {
			index, err := strconv.Atoi(saveIndex[0])
			if err == nil && index > 0 && index < len(records) {
				editedRow := r.Form[fmt.Sprintf("row%d[]", index)]
				if len(editedRow) == len(records[0]) {
					records[index] = editedRow 
				}
			}
	
		
		} else if deleteIndex, ok := r.Form["delete"]; ok {
			index, err := strconv.Atoi(deleteIndex[0])
			if err == nil && index > 0 && index < len(records) {
				records = append(records[:index], records[index+1:]...)
			}
	
		
		} else if _, ok := r.Form["add"]; ok {
			newRow := r.Form["newrow[]"]
			if len(newRow) == len(records[0]) {
				records = append(records, newRow)

                // log.AuthLog.WithFields(logrus.Fields{
                //     "action": "csv_add_row",
                //     "new_row": newRow, // Be cautious about logging sensitive data
                //     "ip":     GetClientIP(r),
                //     "token":  ExtractToken(r),
                // }).Info("CSV row added")
			}
		}
	
		
		outputFile, err := os.Create(csvFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer outputFile.Close()
	
		writer := csv.NewWriter(outputFile)
		if err := writer.WriteAll(records); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	
		
		http.Redirect(w, r, "/csv", http.StatusFound)	
    }
}

func uploadCSVHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    conf := datastore.GetConfig().Polling
    csvFile := conf.Path

    // csvFile := watcher.Dns_param.Path
    if r.Method != http.MethodPost {
        return
    }
        
    file, _, err := r.FormFile("csvfile")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    
    dst, err := os.Create(csvFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    _, err = io.Copy(dst, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    
    fmt.Println("File updated successfully!")

    
    http.Redirect(w, r, "/csv", http.StatusFound)
    
    logger.Audit.WithFields(logrus.Fields{
        "action": "csv_upload",
        "ip":     tools.GetClientIP(r),
        "token":  tools.ExtractToken(r),
    }).Info("CSV file uploaded")
}

func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    conf := datastore.GetConfig().Polling
    csvFile := conf.Path

    file, err := os.Open(csvFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer file.Close()

    
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename="+csvFile)

    
    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    logger.Audit.WithFields(logrus.Fields{
        "action": "csv_download",
        "ip":     tools.GetClientIP(r),
        "token":  tools.ExtractToken(r),
    }).Info("CSV file downloaded")
}


// New handler for AJAX delete
func deleteCsvRowHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    conf := datastore.GetConfig().Polling
    
    if r.Method != "POST" {
        http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
        return
    }

    // Structure to parse the JSON request body
    type DeleteRequest struct {
        Index int `json:"index"`
    }

    var req DeleteRequest
    // Parse the JSON body
    err = json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    csvFile := conf.Path

    // Open and read the CSV file
    file, err := os.Open(csvFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer file.Close()

    records, err := csv.NewReader(file).ReadAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Check if the index is valid
    if req.Index <= 0 || req.Index >= len(records) {
        http.Error(w, "Invalid index", http.StatusBadRequest)
        return
    }
    deletedRowData := records[req.Index]
    // Remove the row from the CSV data
    records = append(records[:req.Index], records[req.Index+1:]...)

    // Write the modified data back to the CSV file
    file, err = os.Create(csvFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    if err := writer.WriteAll(records); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    logger.Audit.WithFields(logrus.Fields{
        "action": "csv_delete_row",
        "ip":     tools.GetClientIP(r),
        "token":  tools.ExtractToken(r),
        "deleted_data": deletedRowData,
    }).Info("CSV row deleted")
    // Send a success response back
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "Row deleted successfully")
}

func editCsvRowHandler(w http.ResponseWriter, r *http.Request) {
    var err error
	conf := datastore.GetConfig().Polling

    if r.Method != "POST" {
        http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
        return
    }

    // Structure to parse the JSON request body
    type EditRequest struct {
        Index int      `json:"index"`
        Data  []string `json:"data"`
    }

    var req EditRequest
    // Parse the JSON body
    err = json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    csvFile := conf.Path

    // Open and read the CSV file
    file, err := os.Open(csvFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer file.Close()

    records, err := csv.NewReader(file).ReadAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Validate the index and update the record
    if req.Index > 0 && req.Index < len(records) {
        if len(req.Data) == len(records[0]) {
            records[req.Index] = req.Data
        } else {
            http.Error(w, "Invalid data length", http.StatusBadRequest)
            return
        }
    } else {
        http.Error(w, "Invalid index", http.StatusBadRequest)
        return
    }

    // Write the modified data back to the CSV file
    outputFile, err := os.Create(csvFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer outputFile.Close()

    writer := csv.NewWriter(outputFile)
    if err := writer.WriteAll(records); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    logger.Audit.WithFields(logrus.Fields{
        "action": "csv_edit_row",
        "ip":     tools.GetClientIP(r),
        "token":  tools.ExtractToken(r),
        "new_data":  req.Data,
    }).Info("CSV row edited")
    // Send a success response back
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "Row edited successfully")
}

func escapeHTML(s string) string {
    return strings.ReplaceAll(s, "&", "&amp;")
}