package sqldb

import (
	"database/sql"
	"log"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
)

func InsertLogConfig(db *sql.DB, conf LogConfiguration, logType int) error {
    // Delete all existing records
    deleteSQL := `DELETE FROM config_logging WHERE type = ?`
    _, err := db.Exec(deleteSQL, logType)
    if err != nil {
        // log.Fatal(err)
        return err
    }

    // Insert new record
    insertSQL := `INSERT INTO config_logging (path, minSeverity, maxAge, maxSize, maxFiles, type) VALUES (?, ?, ?, ?, ?, ?)`
    statement, err := db.Prepare(insertSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    _, err = statement.Exec(conf.Path, conf.MinSeverity, conf.MaxAge, conf.MaxSize, conf.MaxFiles, logType)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    return nil
}

func InsertMainConfig(db *sql.DB, conf MainConfiguration) error {
    // Delete all existing records
    deleteSQL := `DELETE FROM config_main`
    _, err := db.Exec(deleteSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }

    // Insert new record
    insertSQL := `INSERT INTO config_main (HostName, IPAddress, DBname, ConfPath, Sync, UpdateInterval, Hash, LastCheck, LastUpdate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
    statement, err := db.Prepare(insertSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    _, err = statement.Exec(conf.Hostname, conf.IPAddress, conf.DBname, conf.ConfPath, conf.Sync, conf.UpdateInterval, conf.Hash, conf.LastCheck, conf.LastUpdate)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    return nil
}

func InsertWebserverConfig(db *sql.DB, conf WebServerConfiguration) error {
    // Delete all existing records
    deleteSQL := `DELETE FROM config_web`
    _, err := db.Exec(deleteSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    // Insert new record
    insertSQL := `INSERT INTO config_web (Port, SslIsEnable, SslCertPath, SslKeyPath, SesionTimeout, InitUsername, InitPassword) VALUES (?, ?, ?, ?, ?, ?, ?)`
    statement, err := db.Prepare(insertSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    _, err = statement.Exec(conf.Port, conf.SslIsEnable, conf.SslCertPath, conf.SslKeyPath, conf.SesionTimeout, conf.InitUsername, conf.InitPassword)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    return nil
}

func InsertPrometheusConfig(db *sql.DB, conf PrometheusConfiguration) error {
    // Delete all existing records
    deleteSQL := `DELETE FROM config_prometheus`
    _, err := db.Exec(deleteSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }

    // Insert new record
    insertSQL := `INSERT INTO config_prometheus (Url, MetricName, Auth, Username, Password, RetriesCount, BuferSize) VALUES (?, ?, ?, ?, ?, ?, ?)`
    statement, err := db.Prepare(insertSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    _, err = statement.Exec(conf.Url, conf.MetricName, conf.Auth, conf.Username, conf.Password, conf.RetriesCount, conf.BuferSize)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    return nil
}

func InsertPrometeusLabelsConfig(db *sql.DB, config PrometheusLabelConfiguration) error {
    // Delete all existing records
    deleteSQL := `DELETE FROM prometheus_label_config`
    _, err := db.Exec(deleteSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    val := reflect.ValueOf(config)
    typ := val.Type()

    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        name := typ.Field(i).Name

        if field.Kind() == reflect.Bool {
            insertSQL := `INSERT INTO prometheus_label_config (label, isEnable) VALUES (?, ?)`
            statement, err := db.Prepare(insertSQL)
            if err != nil {
                return err
            }
            defer statement.Close()

            _, err = statement.Exec(name, field.Bool())
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func InsertResolversConfig(db *sql.DB, conf ResolversConfiguration) error {
    // Delete all existing records
    deleteSQL := `DELETE FROM config_resolver`
    _, err := db.Exec(deleteSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    // Insert new record
    insertSQL := `INSERT INTO config_resolver (Path, PullTimeout, Delimeter, ExtraDelimeter, Hash, LastCheck, LastUpdate) VALUES (?, ?, ?, ?, ?, ?, ?)`
    statement, err := db.Prepare(insertSQL)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    _, err = statement.Exec(conf.Path, conf.PullTimeout, conf.Delimeter, conf.ExtraDelimeter, conf.Hash, conf.LastCheck, conf.LastUpdate)
    if err != nil {
        // log.Fatal(err)
        return err
    }
    return nil
}

func InsertSyncConfig(db *sql.DB, config SyncConfiguration) error {
    // Start transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    // Delete existing records in config_sync and config_sync_members
    _, err = tx.Exec("DELETE FROM config_sync")
    if err != nil {
        tx.Rollback()
        return err
    }
    _, err = tx.Exec("DELETE FROM config_sync_members")
    if err != nil {
        tx.Rollback()
        return err
    }

    // Insert into config_sync
    res, err := tx.Exec("INSERT INTO config_sync (is_enable, token) VALUES (?, ?)", config.IsEnable, config.Token)
    if err != nil {
        tx.Rollback()
        return err
    }

    // Get the last inserted id from config_sync
    lastID, err := res.LastInsertId()
    if err != nil {
        tx.Rollback()
        return err
    }

    // Insert members into config_sync_members
    for _, member := range config.Members {
        _, err := tx.Exec("INSERT INTO config_sync_members (sync_id, hostname, port, IsLocal) VALUES (?, ?, ?, ?)",
            lastID, member.Hostname, member.Port, 0) // Replace 0 with the actual value for IsLocal if needed
        if err != nil {
            tx.Rollback()
            return err
        }
    }

    // Commit the transaction
    return tx.Commit()
}

func InsertResolver(db *sql.DB, resolver Resolver) error {
    insertSQL := `INSERT INTO Resolvers (Server, IPAddress, Domain, Location, Site, ServerSecurityZone, Prefix, Protocol, Zonename, Recursion, QueryCount, ServiceMode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    _, err := db.Exec(insertSQL, resolver.Server, resolver.IPAddress, resolver.Domain, resolver.Location, resolver.Site, resolver.ServerSecurityZone, resolver.Prefix, resolver.Protocol, resolver.Zonename, resolver.Recursion, resolver.QueryCount, resolver.ServiceMode)
    return err
}

func InsertResolvers(db *sql.DB, servers []Resolver) error {
    for _, resolver := range servers {
        err := InsertResolver(db, resolver)
        if err != nil {
            if isUniqueViolationError(err) {
                // Log and skip the record
                log.Printf("Skipping record due to unique constraint violation: %v\n", resolver)
                continue
                // return nil
            }
            // Handle the error, perhaps roll back if you're using a transaction
            return err
        }
    }
    return nil
}

func InsertWatcherConfig(db *sql.DB, conf WatcherConfiguration) error {
    // Delete all existing records (if you want to keep only the latest configuration)
    deleteSQL := `DELETE FROM config_watcher`
    _, err := db.Exec(deleteSQL)
    if err != nil {
        return err
    }

    // Insert new record
    insertSQL := `INSERT INTO config_watcher (Location, SecurityZone) VALUES (?, ?)`
    statement, err := db.Prepare(insertSQL)
    if err != nil {
        return err
    }
    defer statement.Close()

    _, err = statement.Exec(conf.Location, conf.SecurityZone)
    return err
}