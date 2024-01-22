package webserver

import (
	"HighFrequencyDNSChecker/components/datastore"
	"bufio"
	"html/template"
	"net/http"
	"os"
)

func getLastLinesOfFile(filePath string, numOfLines int) ([]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var lines []string

    for scanner.Scan() {
        lines = append([]string{scanner.Text()}, lines...) // Prepend new lines
        if len(lines) > numOfLines {
            lines = lines[:numOfLines] // Keep only the last numOfLines lines
        }
    }

    return lines, scanner.Err()
}


func logPageHandler(w http.ResponseWriter, r *http.Request) {
    // Default to regular log
    logPath := datastore.GetConfig().AppLogger.Path
    auditPath := datastore.GetConfig().AuditLogger.Path

    logType := r.URL.Query().Get("logType")
    if logType == "" {
        logType = "log"
    }

    var logFilePath string
    if logType != "audit" {
        logFilePath = logPath
    } else {
        logFilePath = auditPath
    }

    logs, err := getLastLinesOfFile(logFilePath, 100)
    if err != nil {
        http.Error(w, "Unable to read log file", http.StatusInternalServerError)
        return
    }

	tmpl, err := template.ParseFS(tmplFS, "html/logview.html", "html/navbar.html", "html/styles.html")
    if err != nil {
        http.Error(w, "Unable to parse template", http.StatusInternalServerError)
        return
    }

    data := struct {
        Logs    []string
        Selected string
    }{
        Logs: logs,
        Selected: logType,
    }

    tmpl.Execute(w, data)
}
