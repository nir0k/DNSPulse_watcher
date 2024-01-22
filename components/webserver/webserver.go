package webserver

import (
	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/logger"
	"HighFrequencyDNSChecker/components/tools"
	"HighFrequencyDNSChecker/components/webserver/api"
	"context"
	"crypto/tls"
	"embed"
	"fmt"

	"io/fs"
	"net/http"
	"strconv"
)

//go:embed html/*.html static/*
var tmplFS embed.FS

func Webserver() {
	done := make(chan bool)
    go startServer(done)
	<-done
}


func startServer(done chan bool) {
    var (
        server *http.Server
        err error
    )
	conf := datastore.GetConfig().WebServer

    if !tools.CheckPortAvailability(conf.Port) {
		logger.Logger.Errorf("Port is already in use. Cannot start the web server. Port: %d\n", conf.Port)
        return
    }

    staticFS, err := fs.Sub(tmplFS, "static")
    if err != nil {
        panic(err)
    }
    
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/", authMiddleware(homeHandler))
    http.HandleFunc("/csv", authMiddleware(csvHandler))
    http.HandleFunc("/csv/upload", authMiddleware(uploadCSVHandler))
    http.HandleFunc("/csv/download", authMiddleware(downloadCSVHandler))
    http.HandleFunc("/csv/delete", deleteCsvRowHandler)
    http.HandleFunc("/csv/edit", editCsvRowHandler)
	http.HandleFunc("/config", authMiddleware(configHandler))
	http.HandleFunc("/logout", authMiddleware(logoutHandler))
    http.HandleFunc("/members", authMiddleware(memberConfigurationsHandler))
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "OK")
    })
    http.HandleFunc("/logs", authMiddleware(logPageHandler))
    http.HandleFunc("/api/conf/state", authMiddleware(api.ConfStateHandler))
    http.HandleFunc("/api/conf/download/config", authMiddleware(api.ConfigDownloadHandler))
    http.HandleFunc("/api/conf/download/csv", authMiddleware(api.CsvDownloadHandler))
    http.Handle("/api/api-spec", corsMiddleware(http.HandlerFunc(api.ServeAPISpec(staticFS, conf.Port))))
    http.HandleFunc("/docs", authMiddleware(api.ServeDocs(tmplFS)))
    http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/docs", http.StatusFound)
    })
    http.Handle("/api/", http.StripPrefix("/api/", http.FileServer(http.FS(staticFS))))
    
    server = &http.Server{
        Addr: ":" + strconv.Itoa(conf.Port),
        Handler: nil,
        TLSConfig: &tls.Config{},
    }
    go func() {
        <-done
        if err := server.Shutdown(context.Background()); err != nil {
            fmt.Println("Server Shutdown:", err)
			logger.Logger.Infof("Server Shutdown: %s\n", err)
        }
    }()

    fmt.Println("Server starting on port", conf.Port)
    logger.Logger.Infof("Server starting on port %d", conf.Port)
    err = server.ListenAndServeTLS(conf.SSLCertPath, conf.SSLKeyPath)
    if err != http.ErrServerClosed {
        fmt.Println("Server failed:", err)
        logger.Logger.Infof("Server failed: %s", err)
    }
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        next.ServeHTTP(w, r)
    })
}