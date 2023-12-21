package sqldb

import (
    "database/sql"
    "testing"

    _ "github.com/mattn/go-sqlite3"
)

// TestInitDB tests the initialization of the database and creation of tables.
func TestInitDB(t *testing.T) {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open in-memory db: %v", err)
    }
    defer db.Close()

    err = InitDB(db)
    if err != nil {
        t.Fatalf("Failed to initialize DB: %v", err)
    }

    // Define the expected schema for each table
    tables := map[string][]string{
        "config_main": {"id", "HostName", "IPAddress", "DBname", "ConfPath", "Sync", "UpdateInterval", "Hash", "LastCheck", "LastUpdate"},
        "config_logging": {"id", "path", "minSeverity", "maxAge", "maxSize", "maxFiles", "type"},
        "config_web": {"id", "Port", "SslIsEnable", "SslCertPath", "SslKeyPath", "SesionTimeout", "InitUsername", "InitPassword"},
        "config_prometheus": {"id", "Url", "MetricName", "Auth", "Username", "Password", "RetriesCount"},
        "prometheus_label_config": {"id", "label", "isEnable"},
        "config_resolver": {"id", "Path", "PullTimeout", "Delimeter", "ExtraDelimeter", "Hash", "LastCheck", "LastUpdate"},
        "config_sync": {"id", "is_enable", "token"},
        "config_sync_members": {"id", "sync_id", "hostname", "port", "SeverLastCheck", "ConfigHash", "ConfigLastCheck", "ConfigLastUpdate", "ResolvHash", "ResolvLastCheck", "ResolvLastUpdate", "IsLocal", "SyncEnable"},
        "Resolvers": {"id", "Server", "IPAddress", "Domain", "Location", "Site", "ServerSecurityZone", "Prefix", "Protocol", "Zonename", "Recursion", "QueryCount", "ServiceMode"},
        "config_watcher": {"id", "Location", "SecurityZone"},
        "memebers_view": {"sync_id", "hostname", "port", "IsLocal", "IPAddress", "SeverLastCheck", "ConfigHash", "ResolvHash", "ConfigLastCheck", "ConfigLastUpdate" , "ResolvLastCheck", "ResolvLastUpdate", "SyncEnable"},
        "members_for_sync_view": {"hostname", "port"},
    }

    for tableName, columns := range tables {
        testTableSchema(t, db, tableName, columns)
    }
}

func testTableSchema(t *testing.T, db *sql.DB, tableName string, expectedColumns []string) {
    // Query to check if table exists with the expected columns
    query := `PRAGMA table_info(` + tableName + `);`
    rows, err := db.Query(query)
    if err != nil {
        t.Errorf("Failed to query table schema for %s: %v", tableName, err)
        return
    }
    defer rows.Close()

    var actualColumns []string
    for rows.Next() {
        var colInfo struct {
            ID     int
            Name   string
            Type   string
            NotNull int
            DfltValue *string
            PK int
        }
        if err := rows.Scan(&colInfo.ID, &colInfo.Name, &colInfo.Type, &colInfo.NotNull, &colInfo.DfltValue, &colInfo.PK); err != nil {
            t.Errorf("Failed to scan schema info: %v", err)
            return
        }
        actualColumns = append(actualColumns, colInfo.Name)
    }

    // Compare actual columns with expected columns
    if len(actualColumns) != len(expectedColumns) {
        t.Errorf("Schema of %s table does not match. Expected: %v, Actual: %v", tableName, expectedColumns, actualColumns)
    }

    for i, colName := range expectedColumns {
        if colName != actualColumns[i] {
            t.Errorf("Expected column %s, got %s in table %s", colName, actualColumns[i], tableName)
        }
    }
}
