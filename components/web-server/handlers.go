package webserver

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)


func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, `
    <html>
    <head>
        <title>Config Editor</title>
        <style>
            .navbar { background-color: #333; overflow: hidden; }
            .navbar a { float: left; display: block; color: white; text-align: center; padding: 14px 16px; text-decoration: none; }
            .navbar a.logout { float: right; }
        </style>
    </head>
    <body>
        <div class="navbar">
            <a href="/">Home</a>
            <a href="/csv">Edit CSV File</a>
            <a href="/env">Show/Edit Configuration</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
        <h1>Welcome to the Config Editor</h1>
    </body>
    </html>
    `)
}


func csvHandler(w http.ResponseWriter, r *http.Request) {
    csvFile := "dns_servers.csv"
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

        
        if editIndex, ok := r.URL.Query()["edit"]; ok {
            editingRow, _ = strconv.Atoi(editIndex[0]) 
        }

        
        fmt.Fprintf(w, `
    <html>
    <head>
        <title>Edit CSV File</title>
        <style>
            .navbar { background-color: #333; overflow: hidden; }
            .navbar a { float: left; display: block; color: white; text-align: center; padding: 14px 16px; text-decoration: none; }
            .navbar a.logout { float: right; }
        </style>
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

            fmt.Fprintf(w, "<tr>")
            if i == editingRow {
                
                for _, field := range record {
                    fmt.Fprintf(w, "<td><input type='text' name='row%d[]' value='%s'></td>", i, field)
                }
                fmt.Fprintf(w, "<td><button type='submit' name='save' value='%d'>Save</button>", i)
                fmt.Fprintf(w, "<a href='/csv'><button type='button'>Cancel</button></a></td>")
            } else {
                
                for _, field := range record {
                    fmt.Fprintf(w, "<td>%s</td>", field)
                }
                fmt.Fprintf(w, "<td><a href='/csv?edit=%d'><button type='button'>Edit</button>", i)
                fmt.Fprintf(w, "<button type='submit' name='delete' value='%d'>Delete</button></a></td>", i)
            }
            fmt.Fprintf(w, "</tr>")
        }

        
        fmt.Fprintf(w, "<tr>")
        for range records[0] {
            fmt.Fprintf(w, "<td><input type='text' name='newrow[]'></td>")
        }
        fmt.Fprintf(w, "<td><button type='submit' name='add'>Add</button></td></tr>")
        fmt.Fprintf(w, "</table>")

        
        fmt.Fprintf(w, "</form>")
        
        fmt.Fprintf(w, "<a href='/'>Go to Main Page</a>")
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

func envHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
    if !ok || !checkCredentials(username, password) {
        w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    envFile := ".env" 

    w.Header().Set("Content-Type", "text/html") 
	editMode := r.URL.Query().Get("edit") == "true"

    switch r.Method {
    case "GET":
        content, err := os.ReadFile(envFile)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
		fmt.Fprintf(w, `
    <html>
    <head>
        <title>Show/Edit Configuration</title>
        <style>
            .navbar { background-color: #333; overflow: hidden; }
            .navbar a { float: left; display: block; color: white; text-align: center; padding: 14px 16px; text-decoration: none; }
            .navbar a.logout { float: right; }
			.row { display: flex; align-items: center; margin-bottom: 5px; padding: 5px; }
			.label { width: 300px; }
			input[type='text'], .value { width: 500px; }
			.row:nth-child(odd) { background-color: #e8e8e8; }
        </style>
    </head>
    <body>
        <div class="navbar">
            <a href="/">Home</a>
            <a href="/csv">Edit CSV File</a>
            <a href="/env">Show/Edit Configuration</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
        <h2>Show/Edit Configuration</h2>
    `)

        if editMode {
            fmt.Fprintf(w, "<form method='POST' action='/env'>")
        }

        lines := strings.Split(string(content), "\n")
        var currentGroup, currentComment string
        groupHeaderDetected := false

        for i, line := range lines {
            line = strings.TrimSpace(line)
            if groupHeaderDetected {
                if strings.HasPrefix(line, "#") && !strings.Contains(line, "=") {
                    currentGroup = strings.Trim(line, "# ")
                    groupHeaderDetected = false
                }
                continue
            }

            if strings.HasPrefix(line, "#") {
                if strings.Contains(line, "=") {
                    groupHeaderDetected = true
                    continue
                } else {
                    
                    currentComment = strings.TrimPrefix(line, "# ")
                    continue
                }
            }

            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 {
                tooltip := currentComment
                currentComment = "" 

                if editMode {
                    if currentGroup != "" {
                        fmt.Fprintf(w, "<h3>%s</h3>", currentGroup)
                        currentGroup = "" 
                    }
                    fmt.Fprintf(w, "<div class='row' title='%s'><div class='label'><label for='line%d'>%s</label></div><input type='text' id='line%d' name='%s' value='%s'></div>",
                        tooltip, i, parts[0], i, parts[0], parts[1])
                } else {
                    if currentGroup != "" {
                        fmt.Fprintf(w, "<h3>%s</h3>", currentGroup)
                        currentGroup = "" 
                    }
                    fmt.Fprintf(w, "<div class='row' title='%s'><div class='label'>%s:</div><div class='value'>%s</div></div>",
                        tooltip, parts[0], parts[1])
                }
            }
        }

        if editMode {
            
            fmt.Fprintf(w, "<input type='submit' value='Save Changes'>")
            fmt.Fprintf(w, "<input type='button' value='Cancel' onclick='window.history.back()'>")
            fmt.Fprintf(w, "<a href='/'><button type='button'>Return to Main Page</button></a>")
            fmt.Fprintf(w, "</form>")
        } else {
            
            fmt.Fprintf(w, "<a href='/env?edit=true'><button type='button'>Edit Configuration</button></a>")
            fmt.Fprintf(w, "<a href='/'><button type='button'>Return to Main Page</button></a>")
        }

        fmt.Fprintf(w, "</body></html>")

    case "POST":
        if err := r.ParseForm(); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        
        content, err := os.ReadFile(envFile)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        lines := strings.Split(string(content), "\n")
        for i := range lines {
            line := lines[i]
            if line != "" && !strings.HasPrefix(line, "#") {
                parts := strings.SplitN(line, "=", 2)
                if len(parts) == 2 {
                    if newValue, ok := r.Form[parts[0]]; ok {
                        lines[i] = parts[0] + "=" + newValue[0] 
                    }
                }
            }
        }

        
        updatedContent := strings.Join(lines, "\n")
        if err := os.WriteFile(envFile, []byte(updatedContent), 0644); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        fmt.Fprint(w, "<p>.env updated successfully!</p><p><a href='/env'>Go Back</a></p>")

    }
}


func uploadCSVHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        
        file, _, err := r.FormFile("csvfile")
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        defer file.Close()

        
        dst, err := os.Create("dns_servers.csv")
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
    } 
    // else {
        
    // }
}


func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
    
    csvFile := "dns_servers.csv"

    
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
}
