package webserver

import (
	"context"
	"crypto/tls"
	// "database/sql"
	"fmt"
	"net/http"

	"time"

	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"
	"HighFrequencyDNSChecker/components/web/api"
)

var (
    server *http.Server
)


func WatchForPortChanges(conf sqldb.WebServerConfiguration, done *chan bool) {
    var (
        newConf sqldb.WebServerConfiguration
        err error
    )
    for {
        time.Sleep(1 * time.Minute)

        newConf, err = sqldb.GetWebServerConfig(sqldb.AppDB)
        if err != nil {
            log.AppLog.Error("Failed to get web-server configuration from db:", err)
        }

        if newConf != conf {
            conf = newConf
            if server != nil {
                log.AppLog.Info("Web-Server configuration has been changed:", conf)
                close(*done)

                newDone := make(chan bool)
                *done = newDone

                go StartServer(conf, newDone)
            }
        }
    }
}


func StartServer(conf sqldb.WebServerConfiguration, done chan bool) {
    var (
        server *http.Server
        err error
        // resolversConf sqldb.ResolversConfiguration
        // configs sqldb.Config
    )

    if !CheckPortAvailability(conf.Port) {
        log.AppLog.Error("Port is already in use. Cannot start the web server. Port:", conf.Port)
        return
    }

    // configs, err = sqldb.GetConfgurations(sqldb.AppDB)
    // if err != nil {
    //     log.AppLog.Error("Failed to get resolvers config from db, error: ", err)
    // }

    loginHandlerWithConf := func(w http.ResponseWriter, r *http.Request) {
        loginHandler(w, r, conf)
    }

    http.HandleFunc("/login", loginHandlerWithConf)
    http.HandleFunc("/", authMiddleware(conf, homeHandler))
    http.HandleFunc("/csv", authMiddleware(conf, csvHandler))
    http.HandleFunc("/csv/upload", authMiddleware(conf, uploadCSVHandler))
    http.HandleFunc("/csv/download", authMiddleware(conf, downloadCSVHandler))
    http.HandleFunc("/csv/delete", deleteCsvRowHandler)
    http.HandleFunc("/csv/edit", editCsvRowHandler)
    http.HandleFunc("/env", authMiddleware(conf, configHandler))
	http.HandleFunc("/logout", authMiddleware(conf, logoutHandler))
    http.HandleFunc("/members", authMiddleware(conf, memberConfigurationsHandler))
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "OK")
    })
    http.HandleFunc("/api/config", api.ConfigAPIHandler)
    http.HandleFunc("/api/csv/download", api.DownloadConfigurationHandler)
    http.HandleFunc("/logs", logPageHandler)

    // http.HandleFunc("/update-config", updateConfHandler)
    //http.HandleFunc("/api/config-state", api.ConfigStateHandler)

    

    server = &http.Server{
        Addr: ":" + conf.Port,
        Handler: nil,
        TLSConfig: &tls.Config{},
    }
    go func() {
        <-done
        if err := server.Shutdown(context.Background()); err != nil {
            fmt.Println("Server Shutdown:", err)
            log.AppLog.Info("Server Shutdown:", err)
        }
    }()

    fmt.Println("Server starting on port", conf.Port)
    log.AppLog.Info("Server starting on port", conf.Port)
    err = server.ListenAndServeTLS(conf.SslCertPath, conf.SslKeyPath)
    if err != http.ErrServerClosed {
        fmt.Println("Server failed:", err)
        log.AppLog.Info("Server failed:", err)
    }
}
