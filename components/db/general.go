package sqldb

import (
	"database/sql"
	// "fmt"
	// "log"

	_ "github.com/mattn/go-sqlite3"
)

var (
    AppDB *sql.DB
    DBName string
)


func InitDB(db *sql.DB) error {
    // db, err := sql.Open("sqlite3", dbName)
    // if err != nil {
    //     log.Fatal(err)
    // }
    var err error

    createTableSQL := `CREATE TABLE IF NOT EXISTS config_main (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        HostName varchar(255),
        IPAddress varchar(40),
        DBname varchar(255),
        ConfPath varchar(255),
        Sync INTEGER,
        UpdateInterval INTEGER,
        Hash varchar(255),
        LastCheck INTEGER,
        LastUpdate INTEGER
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS config_logging (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        path varchar(255),
        minSeverity varchar(50),
        maxAge INTEGER,
        maxSize INTEGER,
        maxFiles INTEGER,
        type INTEGER 
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS config_web (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        Port varchar(10),
        SslIsEnable INTEGER,
        SslCertPath varchar(255),
        SslKeyPath varchar(255),
        SesionTimeout INTEGER,
        InitUsername varchar(255),
        InitPassword varchar(255)
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS config_prometheus (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        Url varchar(255),
        MetricName varchar(50),
        Auth INTEGER,
        Username varchar(255),
        Password varchar(255),
        RetriesCount INTEGER,
        BuferSize INTEGER
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS prometheus_label_config (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        label varchar(50),
        isEnable INTEGER
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS config_resolver (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        Path varchar(255),
        PullTimeout INTEGER,
        Delimeter varchar(1),
        ExtraDelimeter varchar(1),
        Hash varchar(255),
        LastCheck INTEGER,
        LastUpdate INTEGER
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS config_sync (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        is_enable INTEGER,
        token varchar(255)
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS config_sync_members (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sync_id INTEGER,
        hostname varchar(255) UNIQUE,
        port varchar(10),
        SeverLastCheck INTEGER,
        ConfigHash varchar(255),
        ConfigLastCheck INTEGER,
        ConfigLastUpdate INTEGER,
        ResolvHash varchar(255),
        ResolvLastCheck INTEGER,
        ResolvLastUpdate INTEGER,
        IsLocal INTEGER,
        SyncEnable INTEGER,
        FOREIGN KEY (sync_id) REFERENCES config_sync(id)
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS Resolvers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        Server varchar(255),
        IPAddress varchar(40),
        Domain varchar(255),
        Location varchar(100),
        Site varchar(100),
        ServerSecurityZone varchar(100),
        Prefix varchar(100),
        Protocol varchar(5),
        Zonename varchar(100),
        Recursion INTEGER,
        QueryCount INTEGER,
        ServiceMode INTEGER,
        UNIQUE (Server, IPAddress, Site, ServerSecurityZone, Zonename, Recursion)
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createTableSQL = `CREATE TABLE IF NOT EXISTS config_watcher (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        Location varchar(100),
        SecurityZone varchar(100)
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }

    createViewSQL := `CREATE VIEW IF NOT EXISTS memebers_view AS
    SELECT csm.sync_id  
        , csm.hostname
        , csm.port
        , COALESCE(csm.IsLocal, 0) as IsLocal
        , COALESCE(cm.IPAddress, '-') as IPAddress
        , COALESCE(csm.SeverLastCheck, 0) as SeverLastCheck
        , COALESCE(csm.ConfigHash, '-') as ConfigHash
        , COALESCE(csm.ResolvHash, '-') as ResolvHash
        , COALESCE(csm.ConfigLastCheck, 0) as ConfigLastCheck
        , COALESCE(csm.ConfigLastUpdate, 0) as ConfigLastUpdate
        , COALESCE(csm.ResolvLastCheck, 0) as ResolvLastCheck
        , COALESCE(csm.ResolvLastUpdate, 0) as ResolvLastUpdate
        , COALESCE(csm.SyncEnable, 1) as SyncEnable
    FROM config_sync_members csm
    LEFT JOIN config_main cm ON cm.HostName = csm.hostname;`

    _, err = db.Exec(createViewSQL)
    if err != nil {
        return err
    }

    createViewSQL = `CREATE VIEW IF NOT EXISTS members_for_sync_view AS
    SELECT hostname 
        , port
    FROM config_sync_members csm
    WHERE IsLocal != 1`

    _, err = db.Exec(createViewSQL)
    if err != nil {
        return err
    }

    return nil
}
