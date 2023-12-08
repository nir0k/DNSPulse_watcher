package webserver

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/nir0k/HighFrequencyDNSChecker/components/watcher"
    "github.com/nir0k/HighFrequencyDNSChecker/components/log"
    "github.com/sirupsen/logrus"
)


func csvHandler(w http.ResponseWriter, r *http.Request) {
    csvFile := watcher.Dns_param.Dns_servers_path
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
            .navbar { background-color: #333; overflow: hidden; }
            .navbar a { float: left; display: block; color: white; text-align: center; padding: 14px 16px; text-decoration: none; }
            .navbar a.logout { float: right; }

            /* Add styles for the table */
            table {
                // background-color: lightblue !important;
                table-layout: auto;
                width: 100%%;
                border-collapse: collapse;
            }
            td {
                border: 1px solid black;
                padding: 8px;
                text-align: left;
                word-wrap: break-word; /* Enable word wrapping */
                min-width: 50px; /* Set a minimum width */
            }
            input[type="text"] {
                width: 100%%; /* Make input fields take up available space */
                box-sizing: border-box; /* Include padding and borders in the width */
            }
            th {
                background-color: lightblue !important;
                position: sticky;
                padding: 8px;
                top: 0;
                background-color: white;
                z-index: 1;
            }
            tr:nth-child(even) {
                background-color: #f2f2f2; /* Light color for alternate rows */
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
                var cellValue = cell.textContent; // Use textContent instead of innerHTML
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
                    row.cells[i].innerHTML = newData[i];
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
            <a href="/env">Show/Edit Configuration</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
        <h1>Edit CSV File</h1>
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
                fmt.Fprintf(w, "<td id='cell-%d-%d'>%s</td>", i, j, field) // Modify each cell like this
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

                log.AuthLog.WithFields(logrus.Fields{
                    "action": "csv_add_row",
                    "new_row": newRow, // Be cautious about logging sensitive data
                    "ip":     GetClientIP(r),
                    "token":  ExtractToken(r),
                }).Info("CSV row added")
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
    csvFile := watcher.Dns_param.Dns_servers_path
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
    
    log.AuthLog.WithFields(logrus.Fields{
        "action": "csv_upload",
        "ip":     GetClientIP(r),
        "token":  ExtractToken(r), // Extract token from the request
    }).Info("CSV file uploaded")
}


func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
    csvFile := watcher.Dns_param.Dns_servers_path

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
    log.AuthLog.WithFields(logrus.Fields{
        "action": "csv_download",
        "ip":     GetClientIP(r),
        "token":  ExtractToken(r),
    }).Info("CSV file downloaded")
}


// New handler for AJAX delete
func deleteCsvRowHandler(w http.ResponseWriter, r *http.Request) {
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
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    csvFile := watcher.Dns_param.Dns_servers_path // or your CSV file path

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
    log.AuthLog.WithFields(logrus.Fields{
        "action": "csv_delete_row",
        "ip":     GetClientIP(r),
        "token":  ExtractToken(r),
        "deleted_data": deletedRowData,
    }).Info("CSV row deleted")
    // Send a success response back
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "Row deleted successfully")
}


func editCsvRowHandler(w http.ResponseWriter, r *http.Request) {
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
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    csvFile := watcher.Dns_param.Dns_servers_path // or your CSV file path

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
    log.AuthLog.WithFields(logrus.Fields{
        "action": "csv_edit_row",
        "ip":     GetClientIP(r),
        "token":  ExtractToken(r),
        "new_data":  req.Data,
    }).Info("CSV row edited")
    // Send a success response back
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "Row edited successfully")
}
