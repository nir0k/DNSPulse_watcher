package webserver

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)


func envHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
    if !ok || !CheckCredentials(username, password) {
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
            .spacer {
                width: 20px; /* Adjust the width as needed */
            }
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
                
            fmt.Fprintf(w, "<input type='submit' value='Save Changes'>")
            fmt.Fprintf(w, "<input type='button' value='Cancel' onclick='window.history.back()'>")
            fmt.Fprintf(w, "</form>")
        } else {
            
            fmt.Fprintf(w, "<a href='/env?edit=true'><button type='button'>Edit Configuration</button></a>")
        }
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
                    fmt.Fprintf(w, "<div class='row' title='%s'><div class='label'><label for='line%d'>%s: </label></div><div class='spacer'></div><input type='text' id='line%d' name='%s' value='%s'></div>",
                        tooltip, i, parts[0], i, parts[0], parts[1])
                } else {
                    if currentGroup != "" {
                        fmt.Fprintf(w, "<h3>%s</h3>", currentGroup)
                        currentGroup = "" 
                    }
                    fmt.Fprintf(w, "<div class='row' title='%s'><div class='label'>%s: </div><div class='spacer'></div><div class='value'>%s</div></div>",
                        tooltip, parts[0], parts[1])
                }
            }
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