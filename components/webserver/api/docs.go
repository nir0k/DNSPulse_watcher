package api

import (
	"HighFrequencyDNSChecker/components/logger"
	"HighFrequencyDNSChecker/components/tools"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

func ServeAPISpec(staticFS fs.FS, port int) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Adjust the path to the spec file if necessary
        specFile := "api-spec.yaml"
    
        // Open the file from the embedded file system
        fileData, err := fs.ReadFile(staticFS, specFile)
        if err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }
        specContent := string(fileData)
        
        serverURL := fmt.Sprintf("https://%s:%d", tools.GetLocalIP(), port)
        specContent = strings.Replace(specContent, "SERVER_URL_PLACEHOLDER", serverURL, -1)

        // Set appropriate content type (assuming YAML)
        w.Header().Set("Content-Type", "application/yaml")
    
        // Serve the file content using bytes.NewReader to provide an io.ReadSeeker
        w.Write([]byte(specContent))
    }
}

func ServeDocs(tmplFS fs.FS) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        htmlFile, err := fs.ReadFile(tmplFS, "html/redoc.html")
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            logger.Logger.Errorf("Error reading redoc.html: %s", err)
            return
        }

        // Set the content type to HTML
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write(htmlFile)
        logger.Logger.Infof("Redoc HTML Content: %s", string(htmlFile))
    }
}
