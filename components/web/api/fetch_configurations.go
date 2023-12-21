package api

import (
	config "HighFrequencyDNSChecker/components/configurations"
	sqldb "HighFrequencyDNSChecker/components/db"
	"HighFrequencyDNSChecker/components/log"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type stateConf struct {
    General            bool
    Log                bool
    Audit              bool
    WebServer          bool
    Sync               bool
    Prometheus         bool
    PrometheusLabels   bool
    Resolvers          bool
    Watcher            bool
}


func CheckConfFromMembers() {
    var (
        currentConfig sqldb.Config
        newConfig sqldb.Config
        newResolvers sqldb.Config
        members []sqldb.MemberForSync
        err error
    )
    currentConfig, err = sqldb.GetConfgurations(sqldb.AppDB)
    if err != nil {
        log.AppLog.Error("Failed to get local configurations. error:", err)
        return
    }
    if !currentConfig.General.Sync {
        log.AppLog.Debug("Check new configuration skipped. Sycn disabled in configuration")
        return
    }
	members, err = sqldb.GetMembersForSync(sqldb.AppDB)
    if err != nil {
        log.AppLog.Error("Failed to get members for sync. error:", err)
        return
    }
    if members == nil {
        log.AppLog.Debug("No servers to sync")
        return
    }
    newConfig = currentConfig
    newResolvers = currentConfig
    for _, m := range members {
        var (
            all bool
            memberConfig sqldb.Config
            err error
        )
        url := fmt.Sprintf("https://%s:%s/api/config", m.Hostname, m.Port)
        memberConfig, err = fetchConfig(url)
        currentTime := time.Now().Unix()
        if err !=nil {
            log.AppLog.Error("error fetching data from '", m, "': ", err)
            continue
        }
        
        err = sqldb.UpdateMemberHashAndTimes(sqldb.AppDB, memberConfig, m.Hostname, currentTime)
        if err != nil {
            log.AppLog.Warn("Failed to update lastchek timestamps for '", m, "' error: ", err)
        }
        all, _ = compareConfig(newConfig, memberConfig)
        if !all {
            continue
        }
        newConfig = memberConfig
        if !compareResolversList(newResolvers.Resolvers, memberConfig.Resolvers) {
            newResolvers = memberConfig
        }
    }
    result, state := compareConfig(currentConfig, newConfig)
    if !result {
        log.AppLog.Info("Configuration is changed. fetch new configuration from ", newConfig.General.Hostname)
        updateConfigurations(&newConfig, state)
    }
    if currentConfig.Resolvers.Hash != newResolvers.Resolvers.Hash {
        err = updateResolversList(newResolvers, currentConfig.Resolvers.Path)
        if err != nil {
            log.AppLog.Error("Failed update resolvers lists, error: ", err)
        }
        log.AppLog.Info("Resolvers list has been changed")    
    }
    if result {
        log.AppLog.Debug("Configuration isn't changed")
    }
}


func fetchConfig(url string) (sqldb.Config, error) {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

    resp, err := client.Get(url)
    if err != nil {
        return sqldb.Config{}, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return sqldb.Config{}, err
    }

    var config sqldb.Config
    if err := json.Unmarshal(body, &config); err != nil {
        return sqldb.Config{}, err
    }

    return config, nil
}

func compareConfig(first sqldb.Config, second sqldb.Config) (bool, stateConf) {
    var (
        state stateConf
        all bool
    )
    if first.General.LastUpdate < second.General.LastUpdate {
        
        state.General = first.General.UpdateInterval == second.General.UpdateInterval
        // state.Log = first.Log != second.Log
        // state.Audit = first.Audit != second.Audit
        // state.WebServer = first.WebServer != second.WebServer
        // state.Sync = first.Sync.IsEnable != second.Sync.IsEnable
        state.Prometheus = first.Prometheus != second.Prometheus
        state.PrometheusLabels = first.PrometheusLabels != second.PrometheusLabels
        state.Resolvers = first.Resolvers != second.Resolvers
        // state.Watcher = first.Watcher != second.Watcher
    }
    all = state.General ||
        // state.Log ||
        // state.Audit ||
        // state.WebServer ||
        // state.Sync ||
        state.Prometheus ||
        state.PrometheusLabels ||
        state.Resolvers // ||
        // state.Watcher 

    return all, state
}

func compareResolversList(first sqldb.ResolversConfiguration, second sqldb.ResolversConfiguration) bool {
    if first.LastUpdate < second.LastUpdate {
        return first.Hash == second.Hash
    }
    return true
}

func updateConfigurations(conf *sqldb.Config, status stateConf) {
    var (
        err error
    )
    if status.General {
        err = sqldb.UpdateMainConfUpdateInterval(sqldb.AppDB, conf.General.UpdateInterval)
        if err != nil {
            log.AppLog.Error("Failed update Main configuration update interval, error:", err)
        }
    }
    // if status.Log {
    //     err = sqldb.UpdateLogConfEditableFields(sqldb.AppDB, conf.Log, 0)
    //     if err != nil {
    //         log.AppLog.Error("Failed update Log configuration, error:", err)
    //     }
    // }
    // if status.Audit {
    //     err = sqldb.UpdateLogConfEditableFields(sqldb.AppDB, conf.Log, 1)
    //     if err != nil {
    //         log.AppLog.Error("Failed update audit log configuration, error:", err)
    //     }
    // }
    // if status.WebServer {
    //     err = sqldb.UpdateWebConfEditableFields(sqldb.AppDB, conf.WebServer)
    //     if err != nil {
    //         log.AppLog.Error("Failed update Web-Server configuration, error:", err)
    //     }
    // }
    // if status.Sync {
    //     err = sqldb.InsertSyncConfig(sqldb.AppDB, conf.Sync)
    //     if err != nil {
    //         log.AppLog.Error("Failed update Web-Server configuration, error:", err)
    //     }

    // }
    if status.Prometheus {
        err = sqldb.UpdatePromConfEditableFields(sqldb.AppDB, conf.Prometheus)
        if err != nil {
            log.AppLog.Error("Failed update Prometheus configuration, error:", err)
        }
    }
    if status.PrometheusLabels {
        err = sqldb.InsertPrometeusLabelsConfig(sqldb.AppDB, conf.PrometheusLabels)
        if err != nil {
            log.AppLog.Error("Failed update Prometheus labels configuration, error:", err)
        }
    }
    if status.Resolvers {
        err = sqldb.UpdateResolvConfEditableFields(sqldb.AppDB, conf.Resolvers)
        if err != nil {
            log.AppLog.Error("Failed update resolvers configuration, error:", err)
        }
    }
    // if status.Watcher {
    //     err = sqldb.UpdateWatcherConfEditableFields(sqldb.AppDB, conf.Watcher)
    //     if err != nil {
    //         log.AppLog.Error("Failed update Prometheus labels configuration, error:", err)
    //     }
    // }
}

func updateResolversList(conf sqldb.Config, path string) error {
    var (
        err error
        resolvers []sqldb.Resolver
        hash string
    )
    url := fmt.Sprintf("https://%s:%s/api/csv/download", conf.General.Hostname, conf.WebServer.Port)
    err = DownloadFile(url, path)
    if err != nil {
       return err
    }
    resolvers, err = config.ReadResolversFromCSV(conf.Resolvers)
    if err != nil {
        return err
    }
    err = sqldb.InsertResolvers(sqldb.AppDB, resolvers)
    if err != nil {
        return err
    }
    hash, err = config.CalculateHash(conf.Resolvers.Path, md5.New)
    if err != nil {
        return err
    }
    err = sqldb.UpdateResolvConfHashAndTimestamp(sqldb.AppDB, hash, time.Now().Unix())
    return err
}

func DownloadFile(url string, filepath string) error {
    // Send a GET request to the API endpoint
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Check that the server actually sent compressed data
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("bad status: %s", resp.Status)
    }

    // Create a file to save the response content
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Copy the response body to the file
    _, err = io.Copy(out, resp.Body)
    return err
}
