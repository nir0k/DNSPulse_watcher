package sqldb

import (
	"database/sql"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
)

func GetMainConfig(db *sql.DB) (MainConfiguration, error) {
    var config MainConfiguration
    row := db.QueryRow("SELECT HostName, IPAddress, DBname, ConfPath, Sync, UpdateInterval, Hash, LastCheck, LastUpdate FROM config_main limit 1")
    err := row.Scan(&config.Hostname, &config.IPAddress, &config.DBname, &config.ConfPath, &config.Sync, &config.UpdateInterval, &config.Hash, &config.LastCheck, &config.LastUpdate)
    if err != nil {
        if err == sql.ErrNoRows {
            return config, nil
        }
        return config, err
    }
    return config, nil
}

func GetResolverConfig(db *sql.DB) (ResolversConfiguration, error) {
    var config ResolversConfiguration
    row := db.QueryRow("SELECT Hash, LastCheck, LastUpdate, Path, PullTimeout, Delimeter, ExtraDelimeter FROM config_resolver limit 1")
    err := row.Scan(&config.Hash, &config.LastCheck, &config.LastUpdate, &config.Path, &config.PullTimeout, &config.Delimeter, &config.ExtraDelimeter)
    if err != nil {
        if err == sql.ErrNoRows {
            return config, nil
        }
        return config, err
    }
    return config, nil
}

func GetLogConfig(db *sql.DB, logType int) (LogConfiguration, error) {
    var config LogConfiguration
    row := db.QueryRow("SELECT path, minSeverity, maxAge, maxSize, maxFiles FROM config_logging WHERE type = ?", logType)
    err := row.Scan(&config.Path, &config.MinSeverity, &config.MaxAge, &config.MaxSize, &config.MaxFiles)
    if err != nil {
        if err == sql.ErrNoRows {
            return config, nil
        }
        return config, err
    }
    return config, nil
}

func GetWebServerConfig(db *sql.DB) (WebServerConfiguration, error) {
    var config WebServerConfiguration
    row := db.QueryRow("SELECT Port, SslIsEnable, SslCertPath, SslKeyPath, SesionTimeout, InitUsername, InitPassword FROM config_web LIMIT 1")
    err := row.Scan(&config.Port, &config.SslIsEnable, &config.SslCertPath, &config.SslKeyPath, &config.SesionTimeout, &config.InitUsername, &config.InitPassword)
    if err != nil {
        if err == sql.ErrNoRows {
            return config, nil
        }
        return config, err
    }
    return config, nil
}

func GetSyncConfig(db *sql.DB) (SyncConfiguration, error) {
    var config SyncConfiguration
    // Get is_enable from config_sync table
    row := db.QueryRow("SELECT is_enable, token FROM config_sync LIMIT 1")
    err := row.Scan(&config.IsEnable, &config.Token)
    if err != nil {
        if err == sql.ErrNoRows {
            return config, nil // No rows found, return empty config
        }
        return config, err
    }
    // Get member configurations from config_sync_members table
    rows, err := db.Query("SELECT sync_id, hostname, port, SeverLastCheck, ConfigHash, ConfigLastCheck, ConfigLastUpdate, ResolvHash, ResolvLastCheck, ResolvLastUpdate, IsLocal FROM memebers_view")
    
    if err != nil {
        return config, err
    }
    defer rows.Close()

    for rows.Next() {
        var member MemberConfiguration
        err := rows.Scan(&member.SyncID, &member.Hostname, &member.Port, &member.SeverLastCheck, &member.ConfigHash, &member.ConfigLastCheck, &member.ConfigLastUpdate, &member.ResolvHash, &member.ResolvLastCheck, &member.ResolvLastUpdate, &member.IsLocal)
        if err != nil {
            return config, err
        }
        config.Members = append(config.Members, member)
    }
    if err := rows.Err(); err != nil {
        return config, err
    }

    return config, nil
}


func GetMembersForSync(db *sql.DB) ([]MemberForSync, error) {
    var members []MemberForSync
    rows, err := db.Query("SELECT hostname , port FROM members_for_sync_view")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    for rows.Next() {
        var member MemberForSync
        err := rows.Scan(&member.Hostname, &member.Port)
        if err != nil {
            return nil, err
        }
        members = append(members, member)
    }
    
    if err = rows.Err(); err != nil {
        return nil, err
    }

    return members, nil
}


func GetPrometheusConfig(db *sql.DB) (PrometheusConfiguration, error) {
    var config PrometheusConfiguration
    row := db.QueryRow("SELECT Url, MetricName, Auth, Username, Password, RetriesCount, BuferSize FROM config_prometheus LIMIT 1")
    err := row.Scan(&config.Url, &config.MetricName, &config.Auth, &config.Username, &config.Password, &config.RetriesCount, &config.BuferSize)
    if err != nil {
        if err == sql.ErrNoRows {
            return config, nil
        }
        return config, err
    }
    return config, nil
}

func GetPrometheusLabelConfig(db *sql.DB) (PrometheusLabelConfiguration, error) {
    var config PrometheusLabelConfiguration
    rows, err := db.Query("SELECT label, isEnable FROM prometheus_label_config")
    if err != nil {
        return config, err
    }
    defer rows.Close()

    val := reflect.ValueOf(&config).Elem()
    typ := val.Type()

    for rows.Next() {
        var label string
        var isEnable int
        err = rows.Scan(&label, &isEnable)
        if err != nil {
            return config, err
        }

        for i := 0; i < val.NumField(); i++ {
            field := val.Field(i)
            fieldName := typ.Field(i).Name

            // Convert label name from snake_case to PascalCase or however it matches your struct field names
            camelCaseLabel := convertToCamelCase(label)
            if fieldName == camelCaseLabel {
                field.SetBool(isEnable != 0)
                break
            }
        }
    }
    return config, nil
}

func GetWatcherConfig(db *sql.DB) (WatcherConfiguration, error) {
    var config WatcherConfiguration
    // Query to retrieve the watcher configuration
    row := db.QueryRow("SELECT Location, SecurityZone FROM config_watcher LIMIT 1")
    err := row.Scan(&config.Location, &config.SecurityZone)
    if err != nil {
        if err == sql.ErrNoRows {
            // No rows found, return an empty config and no error
            return config, nil
        }
        // Return an error for any other type of error
        return config, err
    }
    return config, nil
}

// GetResolvers fetches all resolvers from the database.
func GetResolvers(db *sql.DB) ([]Resolver, error) {
    query := `SELECT Server, IPAddress, Domain, Location, Site, ServerSecurityZone, Prefix, Protocol, Zonename, Recursion, QueryCount, ServiceMode FROM Resolvers`
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var resolvers []Resolver
    for rows.Next() {
        var resolver Resolver
        err := rows.Scan(&resolver.Server, &resolver.IPAddress, &resolver.Domain, &resolver.Location, &resolver.Site, &resolver.ServerSecurityZone, &resolver.Prefix, &resolver.Protocol, &resolver.Zonename, &resolver.Recursion, &resolver.QueryCount, &resolver.ServiceMode)
        if err != nil {
            return nil, err
        }
        resolvers = append(resolvers, resolver)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return resolvers, nil
}

func GetConfgurations(db *sql.DB) ( Config, error) {
    var (
        conf Config
        err error
    )

    conf.General, err = GetMainConfig(db)
    if err != nil {
        return conf, err
    }
    conf.Log, err = GetLogConfig(db, 0)
    if err != nil {
        return conf, err
    }
    conf.Audit, err = GetLogConfig(db, 1)
    if err != nil {
        return conf, err
    }
    conf.WebServer, err = GetWebServerConfig(db)
    if err != nil {
        return conf, err
    }
    conf.Sync, err = GetSyncConfig(db)
    if err != nil {
        return conf, err
    }
    conf.Prometheus, err = GetPrometheusConfig(db)
    if err != nil {
        return conf, err
    }
    conf.PrometheusLabels, err = GetPrometheusLabelConfig(db)
    if err != nil {
        return conf, err
    }
    conf.Resolvers, err = GetResolverConfig(db)
    if err != nil {
        return conf, err
    }
    conf.Watcher, err = GetWatcherConfig(db)
    if err != nil {
        return conf, err
    }

    return conf, nil
}

func GetLogPaths(db *sql.DB) (LogsPaths, error) {
    query := `SELECT type, path FROM config_logging`
    rows, err := db.Query(query)
    if err != nil {
        return LogsPaths{}, err
    }
    defer rows.Close()

    var logs LogsPaths
    for rows.Next() {
        var (
            t int
            path string
        )
        err := rows.Scan(&t, &path)
        if err != nil {
            return LogsPaths{}, err
        }
        if t == 1 {
            logs.Log = path
        } else {
            logs.Audit = path
        }
    }
    return logs, err
}