package sqldb

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)


func UpdateLastCheck(db *sql.DB, dbname string, lastCheck int64) error {
    updateSQL := `UPDATE config_main SET LastCheck = ? WHERE DBname = ?`
    statement, err := db.Prepare(updateSQL)
    if err != nil {
        return err
    }
    defer statement.Close()

    _, err = statement.Exec(lastCheck, dbname)
    return err
}

func UpsertSyncMember(db *sql.DB, syncMember MemberConfiguration) error {
    upsertSQL := `INSERT OR REPLACE INTO config_sync_members (sync_id, hostname, port, SeverLastCheck, ConfigHash, ConfigLastCheck, ConfigLastUpdate, ResolvHash, ResolvLastCheck, ResolvLastUpdate, IsLocal) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

    _, err := db.Exec(upsertSQL, syncMember.SyncID, syncMember.Hostname, syncMember.Port, syncMember.SeverLastCheck, syncMember.ConfigHash, syncMember.ConfigLastCheck, syncMember.ConfigLastUpdate, syncMember.ResolvHash, syncMember.ResolvLastCheck, syncMember.ResolvLastUpdate, syncMember.IsLocal)
    return err
}

func UpdateSyncMemberTimestamps(db *sql.DB, hostname string, conflastCheck int64, configLastUpdate int64, resolvLastCheck int64, resolvLastUpdate int64, severLastCheck int64, sync bool) error {
    var (
        err error
        updateSQL string
    )
    // CheckConnectDB()

    if conflastCheck != 0 && resolvLastCheck != 0 && severLastCheck != 0 {
        updateSQL = `UPDATE config_sync_members SET SeverLastCheck = ?, ConfigLastCheck = ?, ResolvLastCheck = ?, SyncEnable = ? WHERE hostname = ?`
        db.Exec(updateSQL, severLastCheck, conflastCheck, resolvLastCheck, sync, hostname)
    } else if conflastCheck != 0 && resolvLastCheck != 0 {
        updateSQL = `UPDATE config_sync_members SET SeverLastCheck = ?, ConfigLastCheck = ?, ResolvLastCheck = ?, SyncEnable = ? WHERE hostname = ?`
        _, err = db.Exec(updateSQL, conflastCheck, conflastCheck, resolvLastCheck, sync, hostname)
    } else if conflastCheck != 0 {
        updateSQL = `UPDATE config_sync_members SET SeverLastCheck = ?, ConfigLastCheck = ?, SyncEnable = ? WHERE hostname = ?`
        _, err = db.Exec(updateSQL, conflastCheck, conflastCheck, sync, hostname)
    } else if resolvLastCheck != 0 {
        updateSQL = `UPDATE config_sync_members SET SeverLastCheck = ?, ResolvLastCheck = ?, SyncEnable = ? WHERE hostname = ?`
        _, err = db.Exec(updateSQL, resolvLastCheck, resolvLastCheck, sync, hostname)
    }
    
    if configLastUpdate != 0 && resolvLastUpdate != 0 {
        updateSQL = `UPDATE config_sync_members SET ConfigLastUpdate = ?, ResolvLastUpdate =?, SyncEnable = ? WHERE hostname = ?`
        _, err = db.Exec(updateSQL, configLastUpdate, resolvLastUpdate, sync, hostname)
    } else if configLastUpdate != 0 {
        updateSQL = `UPDATE config_sync_members SET ConfigLastUpdate = ?, SyncEnable = ? WHERE hostname = ?`
        _, err = db.Exec(updateSQL, configLastUpdate, sync, hostname)
    } else if resolvLastUpdate != 0 {
        updateSQL = `UPDATE config_sync_members SET  ResolvLastUpdate =?, SyncEnable = ? WHERE hostname = ?`
        _, err = db.Exec(updateSQL, resolvLastUpdate, sync, hostname)
    }
    return err
}

func UpdateMainConfigTimestamps(db *sql.DB, dbname string, lastCheck int64, lastUpdate int64) error {
    var updateSQL string
    var err error

    if lastCheck != 0 && lastUpdate != 0 {
        // Update both LastCheck and LastUpdate
        updateSQL = `UPDATE config_main SET LastCheck = ?, LastUpdate = ? WHERE DBname = ?`
        _, err = db.Exec(updateSQL, lastCheck, lastUpdate, dbname)
    } else if lastCheck != 0 {
        // Update only LastCheck
        updateSQL = `UPDATE config_main SET LastCheck = ? WHERE DBname = ?`
        _, err = db.Exec(updateSQL, lastCheck, dbname)
    } else if lastUpdate != 0 {
        // Update only LastUpdate
        updateSQL = `UPDATE config_main SET LastUpdate = ? WHERE DBname = ?`
        _, err = db.Exec(updateSQL, lastUpdate, dbname)
    } else {
        // Neither LastCheck nor LastUpdate provided
        return errors.New("no valid timestamp provided for update")
    }

    return err
}

func UpdateResolversConfigTimestamps(db *sql.DB, path string, lastCheck int64, lastUpdate int64) error {
    var updateSQL string
    var err error

    if lastCheck != 0 && lastUpdate != 0 {
        // Update both LastCheck and LastUpdate
        updateSQL = `UPDATE config_resolver SET LastCheck = ?, LastUpdate = ? WHERE Path = ?`
        _, err = db.Exec(updateSQL, lastCheck, lastUpdate, path)
    } else if lastCheck != 0 {
        // Update only LastCheck
        updateSQL = `UPDATE config_resolver SET LastCheck = ? WHERE Path = ?`
        _, err = db.Exec(updateSQL, lastCheck, path)
    } else if lastUpdate != 0 {
        // Update only LastUpdate
        updateSQL = `UPDATE config_resolver SET LastUpdate = ? WHERE Path = ?`
        _, err = db.Exec(updateSQL, lastUpdate, path)
    } else {
        // Neither LastCheck nor LastUpdate provided
        return errors.New("no valid timestamp provided for update")
    }

    return err
}

func UpdateMainConfEditableFields(db *sql.DB, conf MainConfiguration) error {
	updateSQL := `UPDATE config_main SET ConfPath = ?, Sync = ?, UpdateInterval = ?`
        _, err := db.Exec(updateSQL, conf.ConfPath, conf.Sync, conf.UpdateInterval)
    return err
}

func UpdateMainConfUpdateInterval(db *sql.DB, interval int) error {
    updateSQL := `UPDATE config_main SET UpdateInterval = ?`
        _, err := db.Exec(updateSQL, interval)
    return err
}

func UpdateLogConfEditableFields(db *sql.DB, conf LogConfiguration, logType int) error {
    updateSQL := `UPDATE config_logging SET path = ?, minSeverity = ?, maxAge = ?, maxSize = ?, maxFiles = ? WHERE type = ?`
        _, err := db.Exec(updateSQL, conf.Path, conf.MinSeverity, conf.MaxAge, conf.MaxSize, conf.MaxFiles, logType)
    return err
}

func UpdateWebConfEditableFields(db *sql.DB, conf WebServerConfiguration) error {
    updateSQL := `UPDATE config_web SET Port = ?, SslIsEnable = ?, SslCertPath = ?, SslKeyPath = ?, SesionTimeout = ?, InitUsername =?, InitPassword = ?`
        _, err := db.Exec(updateSQL, conf.Port, conf.SslIsEnable, conf.SslCertPath, conf.SslKeyPath, conf.SesionTimeout, conf.InitUsername, conf.InitPassword)
    return err
}

func UpdateSyncConfEditableFields(db *sql.DB, conf SyncConfiguration) error {
    updateSQL := `UPDATE config_sync SET is_enable = ?, token = ?`
        _, err := db.Exec(updateSQL, conf.IsEnable, conf.Token)
    return err
}

func UpdatePromConfEditableFields(db *sql.DB, conf PrometheusConfiguration) error {
    updateSQL := `UPDATE config_prometheus SET Url = ?, MetricName = ?, Auth = ?, Username = ?, Password = ?, RetriesCount = ?, BuferSize = ?`
        _, err := db.Exec(updateSQL, conf.Url, conf.MetricName, conf.Auth, conf.Username, conf.Password, conf.RetriesCount, conf.BuferSize)
    return err
}

// func UpdatePromLabelsConfEditableFields(db *sql.DB, conf PrometheusConfiguration) error {
//  updateSQL := `UPDATE config_prometheus SET Url = ?, MetricName = ?, Auth = ?, Username = ?, Password = ?, RetriesCount =?`
//         _, err := db.Exec(updateSQL, conf.Url, conf.MetricName, conf.Auth, conf.Username, conf.Password, conf.RetriesCount)
//     return err
// }

func UpdateResolvConfEditableFields(db *sql.DB, conf ResolversConfiguration) error {
    updateSQL := `UPDATE config_resolver SET Path = ?, PullTimeout = ?, Delimeter = ?, ExtraDelimeter = ?`
        _, err := db.Exec(updateSQL, conf.Path, conf.PullTimeout, conf.Delimeter, conf.ExtraDelimeter)
    return err
}

func UpdateWatcherConfEditableFields(db *sql.DB, conf WatcherConfiguration) error {
    updateSQL := `UPDATE config_watcher SET Location = ?, SecurityZone = ?`
        _, err := db.Exec(updateSQL, conf.Location, conf.SecurityZone)
    return err
}

func UpdateResolvConfHashAndTimestamp(db *sql.DB, hash string, timestamp int64) error {
    updateSQL := `UPDATE config_resolver SET Hash = ?, LastCheck = ?, LastUpdate = ?`
        _, err := db.Exec(updateSQL, hash, timestamp, timestamp)
    return err
}

func UpdateMemberHashAndTimes(db *sql.DB, conf Config, hostname string, currentTime int64) error {
    updateSQL := `UPDATE config_sync_members SET SeverLastCheck = ?, ConfigHash = ?, ConfigLastCheck = ?, ConfigLastUpdate = ?, ResolvHash = ?, ResolvLastCheck = ?, ResolvLastUpdate = ?, SyncEnable = ? WHERE hostname = ?`
        _, err := db.Exec(updateSQL, currentTime, conf.General.Hash, conf.General.LastCheck, conf.General.LastUpdate, conf.Resolvers.Hash, conf.Resolvers.LastCheck, conf.Resolvers.LastUpdate, conf.General.Sync, hostname)
    return err
}