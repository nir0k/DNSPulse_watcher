package sqldb

import (
	"database/sql"
	// "log"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

)

func setupInMemoryDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open in-memory db: %v", err)
    }
    if err = InitDB(db); err != nil {
        t.Fatalf("Failed to initialize db: %v", err)
    }
    return db
}

func TestInsertMainConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    conf := MainConfiguration{
        DBname:         "testDB",
        Hostname:       "testHost",
        IPAddress:      "192.168.1.1",
        ConfPath:       "/test/path",
        Sync:           true,
        UpdateInterval: 30,
        Hash:           "testhash",
        LastCheck:      1625076000,
        LastUpdate:     1625076000,
    }

    if err := InsertMainConfig(db, conf); err != nil {
        t.Errorf("InsertMainConfig failed: %v", err)
    }

    var result MainConfiguration
    err := db.QueryRow("SELECT DBname, Hostname, IPAddress, ConfPath, Sync, UpdateInterval, Hash, LastCheck, LastUpdate FROM config_main").Scan(&result.DBname, &result.Hostname, &result.IPAddress, &result.ConfPath, &result.Sync, &result.UpdateInterval, &result.Hash, &result.LastCheck, &result.LastUpdate)
    if err != nil {
        t.Errorf("Failed to get main config: %v", err)
    }

    if result != conf {
        t.Errorf("Retrieved config does not match inserted config. Got %+v, want %+v", result, conf)
    }
}


func TestInsertLogConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    logConf := LogConfiguration{
        Path:        "test.log",
        MinSeverity: "INFO",
        MaxAge:      7,
        MaxSize:     100,
        MaxFiles:    5,
    }

    if err := InsertLogConfig(db, logConf, 0); err != nil {
        t.Errorf("InsertLogConfig failed: %v", err)
    }

    var result LogConfiguration
    err := db.QueryRow("SELECT path, minSeverity, maxAge, maxSize, maxFiles FROM config_logging WHERE type = ?", 0).Scan(&result.Path, &result.MinSeverity, &result.MaxAge, &result.MaxSize, &result.MaxFiles)
    if err != nil {
        t.Errorf("Failed to get log config: %v", err)
    }

    if result.Path != logConf.Path || result.MinSeverity != logConf.MinSeverity || result.MaxAge != logConf.MaxAge || result.MaxSize != logConf.MaxSize || result.MaxFiles != logConf.MaxFiles {
        t.Errorf("Retrieved log config does not match inserted config. Got %+v, want %+v", result, logConf)
    }
}

func TestInsertWebserverConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    webServerConf := WebServerConfiguration{
        Port:         "8080",
        SslIsEnable:  true,
        SslCertPath:  "/ssl/cert/path",
        SslKeyPath:   "/ssl/key/path",
        SesionTimeout: 30,
        InitUsername: "admin",
        InitPassword: "password",
    }

    if err := InsertWebserverConfig(db, webServerConf); err != nil {
        t.Errorf("InsertWebserverConfig failed: %v", err)
    }

    var result WebServerConfiguration
    err := db.QueryRow("SELECT Port, SslIsEnable, SslCertPath, SslKeyPath, SesionTimeout, InitUsername, InitPassword FROM config_web").Scan(&result.Port, &result.SslIsEnable, &result.SslCertPath, &result.SslKeyPath, &result.SesionTimeout, &result.InitUsername, &result.InitPassword)
    if err != nil {
        t.Errorf("Failed to get web server config: %v", err)
    }

    if result.Port != webServerConf.Port || result.SslIsEnable != webServerConf.SslIsEnable || result.SslCertPath != webServerConf.SslCertPath || result.SslKeyPath != webServerConf.SslKeyPath || result.SesionTimeout != webServerConf.SesionTimeout || result.InitUsername != webServerConf.InitUsername || result.InitPassword != webServerConf.InitPassword {
        t.Errorf("Retrieved web server config does not match inserted config. Got %+v, want %+v", result, webServerConf)
    }
}

func TestInsertPrometheusConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    prometheusConf := PrometheusConfiguration{
        Url:          "http://localhost:9090",
        MetricName:   "test_metric",
        Auth:         true,
        Username:     "user",
        Password:     "pass",
        RetriesCount: 3,
    }

    if err := InsertPrometheusConfig(db, prometheusConf); err != nil {
        t.Errorf("InsertPrometheusConfig failed: %v", err)
    }

    var result PrometheusConfiguration
    err := db.QueryRow("SELECT Url, MetricName, Auth, Username, Password, RetriesCount FROM config_prometheus").Scan(&result.Url, &result.MetricName, &result.Auth, &result.Username, &result.Password, &result.RetriesCount)
    if err != nil {
        t.Errorf("Failed to get Prometheus config: %v", err)
    }

    if result.Url != prometheusConf.Url || result.MetricName != prometheusConf.MetricName || result.Auth != prometheusConf.Auth || result.Username != prometheusConf.Username || result.Password != prometheusConf.Password || result.RetriesCount != prometheusConf.RetriesCount {
        t.Errorf("Retrieved Prometheus config does not match inserted config. Got %+v, want %+v", result, prometheusConf)
    }
}

func TestInsertPrometeusLabelsConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    labelsConf := PrometheusLabelConfiguration{
        Opcode:              true,
        Authoritative:       false,
        Truncated:           true,
        Rcode:               false,
        RecursionDesired:    true,
        RecursionAvailable:  false,
        AuthenticatedData:   true,
        CheckingDisabled:    false,
        PollingRate:         true,
        Recursion:           true,
    }

    if err := InsertPrometeusLabelsConfig(db, labelsConf); err != nil {
        t.Errorf("InsertPrometeusLabelsConfig failed: %v", err)
    }

    rows, err := db.Query("SELECT label, isEnable FROM prometheus_label_config")
    if err != nil {
        t.Fatalf("Failed to query prometheus_label_config: %v", err)
    }
    defer rows.Close()

    retrievedLabels := make(map[string]bool)
    for rows.Next() {
        var label string
        var isEnable bool
        if err := rows.Scan(&label, &isEnable); err != nil {
            t.Errorf("Failed to scan row: %v", err)
            continue
        }
        retrievedLabels[label] = isEnable
    }

    val := reflect.ValueOf(labelsConf)
    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        name := val.Type().Field(i).Name
        if field.Kind() == reflect.Bool && retrievedLabels[name] != field.Bool() {
            t.Errorf("Label %s does not match. Expected: %v, got: %v", name, field.Bool(), retrievedLabels[name])
        }
    }
}

func TestInsertResolversConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    resolversConf := ResolversConfiguration{
        Path:            "/test/resolver/path",
        PullTimeout:     30,
        Delimeter:       ",",
        ExtraDelimeter:  ";",
        Hash:            "resolverhash",
        LastCheck:       time.Now().Unix(),
        LastUpdate:      time.Now().Unix(),
    }

    if err := InsertResolversConfig(db, resolversConf); err != nil {
        t.Errorf("InsertResolversConfig failed: %v", err)
    }

    var config ResolversConfiguration
    row := db.QueryRow("SELECT Path, PullTimeout, Delimeter, ExtraDelimeter, Hash, LastCheck, LastUpdate FROM config_resolver")
    if err := row.Scan(&config.Path, &config.PullTimeout, &config.Delimeter, &config.ExtraDelimeter, &config.Hash, &config.LastCheck, &config.LastUpdate); err != nil {
        t.Fatalf("Failed to scan config_resolver: %v", err)
    }

    if !reflect.DeepEqual(resolversConf, config) {
        t.Errorf("Retrieved config does not match. Expected: %+v, got: %+v", resolversConf, config)
    }
}

func TestInsertSyncConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    syncConf := SyncConfiguration{
        IsEnable: true,
        Members: []MemberConfiguration{
            {
                Hostname: "testHost1",
                Port:     "8080",
                IsLocal:  false,
            },
            {
                Hostname: "testHost2",
                Port:     "9090",
                IsLocal:  false,
            },
        },
    }

    if err := InsertSyncConfig(db, syncConf); err != nil {
        t.Errorf("InsertSyncConfig failed: %v", err)
    }

    // Verify config_sync table
    var isEnable int
    row := db.QueryRow("SELECT is_enable FROM config_sync LIMIT 1")
    if err := row.Scan(&isEnable); err != nil {
        t.Fatalf("Failed to scan config_sync: %v", err)
    }
    if isEnable != 1 {
        t.Errorf("Incorrect is_enable value in config_sync. Expected: 1, got: %v", isEnable)
    }

    // Verify config_sync_members table
    rows, err := db.Query("SELECT hostname, port, IsLocal FROM config_sync_members")
    if err != nil {
        t.Fatalf("Failed to query config_sync_members: %v", err)
    }
    defer rows.Close()

    members := make([]MemberConfiguration, 0)
    for rows.Next() {
        var member MemberConfiguration
        var isLocal int
        if err := rows.Scan(&member.Hostname, &member.Port, &isLocal); err != nil {
            t.Errorf("Failed to scan member: %v", err)
            continue
        }
        member.IsLocal = isLocal == 1
        members = append(members, member)
    }

    if err := rows.Err(); err != nil {
        t.Errorf("Rows iteration error: %v", err)
    }

    // Compare each member
    for i, expectedMember := range syncConf.Members {
        if !reflect.DeepEqual(expectedMember, members[i]) {
            t.Errorf("Member at index %d does not match. Expected: %+v, got: %+v", i, expectedMember, members[i])
        }
    }
}

func TestInsertResolver(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Define a resolver for testing
    resolver := Resolver{
        Server:             "TestServer",
        IPAddress:          "192.168.1.1",
        Domain:             "example.com",
        Location:           "TestLocation",
        Site:               "TestSite",
        ServerSecurityZone: "TestZone",
        Prefix:             "TestPrefix",
        Protocol:           "TCP",
        Zonename:           "TestZoneName",
        Recursion:          true,
        QueryCount:         10,
        ServiceMode:        true,
    }

    // Test InsertResolver
    if err := InsertResolver(db, resolver); err != nil {
        t.Errorf("InsertResolver failed: %v", err)
    }

    // Query the database to verify insertion
    var r Resolver
    row := db.QueryRow("SELECT Server, IPAddress, Domain, Location, Site, ServerSecurityZone, Prefix, Protocol, Zonename, Recursion, QueryCount, ServiceMode FROM Resolvers WHERE Server = ?", resolver.Server)
    if err := row.Scan(&r.Server, &r.IPAddress, &r.Domain, &r.Location, &r.Site, &r.ServerSecurityZone, &r.Prefix, &r.Protocol, &r.Zonename, &r.Recursion, &r.QueryCount, &r.ServiceMode); err != nil {
        t.Errorf("Failed to scan database result: %v", err)
    }

    // Compare retrieved data with the original data
    if !reflect.DeepEqual(resolver, r) {
        t.Errorf("Retrieved data does not match inserted data. Expected: %+v, got: %+v", resolver, r)
    }
}

func TestInsertResolvers(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Define resolvers for testing
    resolvers := []Resolver{
        {
            Server:             "TestServer1",
            IPAddress:          "192.168.1.1",
            Domain:             "example.com",
            Location:           "TestLocation1",
            Site:               "TestSite1",
            ServerSecurityZone: "TestZone1",
            Prefix:             "TestPrefix1",
            Protocol:           "TCP",
            Zonename:           "TestZoneName1",
            Recursion:          true,
            QueryCount:         10,
            ServiceMode:        true,
        },
        {
            Server:             "TestServer2",
            IPAddress:          "192.168.1.2",
            Domain:             "example.org",
            Location:           "TestLocation2",
            Site:               "TestSite2",
            ServerSecurityZone: "TestZone2",
            Prefix:             "TestPrefix2",
            Protocol:           "UDP",
            Zonename:           "TestZoneName2",
            Recursion:          false,
            QueryCount:         20,
            ServiceMode:        false,
        },
    }

    // Test InsertResolvers
    if err := InsertResolvers(db, resolvers); err != nil {
        t.Errorf("InsertResolvers failed: %v", err)
    }

    // Query the database to verify insertion
    for _, expected := range resolvers {
        var r Resolver
        row := db.QueryRow("SELECT Server, IPAddress, Domain, Location, Site, ServerSecurityZone, Prefix, Protocol, Zonename, Recursion, QueryCount, ServiceMode FROM Resolvers WHERE Server = ?", expected.Server)
        if err := row.Scan(&r.Server, &r.IPAddress, &r.Domain, &r.Location, &r.Site, &r.ServerSecurityZone, &r.Prefix, &r.Protocol, &r.Zonename, &r.Recursion, &r.QueryCount, &r.ServiceMode); err != nil {
            t.Errorf("Failed to scan database result for server %s: %v", expected.Server, err)
        }

        // Compare retrieved data with the original data
        if !reflect.DeepEqual(expected, r) {
            t.Errorf("Retrieved data does not match inserted data for server %s. Expected: %+v, got: %+v", expected.Server, expected, r)
        }
    }
}

func TestInsertWatcherConfig(t *testing.T) {
    // Setup in-memory database
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open in-memory db: %v", err)
    }
    defer db.Close()

    // Initialize the database with the required table
    if err := InitDB(db); err != nil {
        t.Fatalf("Failed to initialize db: %v", err)
    }

    // Define a watcher configuration for testing
    conf := WatcherConfiguration{
        Location:     "TestLocation",
        SecurityZone: "TestSecurityZone",
    }

    // Test InsertWatcherConfig
    if err := InsertWatcherConfig(db, conf); err != nil {
        t.Errorf("InsertWatcherConfig failed: %v", err)
    }

    // Query the database to verify insertion
    var result WatcherConfiguration
    err = db.QueryRow("SELECT Location, SecurityZone FROM config_watcher").Scan(&result.Location, &result.SecurityZone)
    if err != nil {
        t.Errorf("Failed to get watcher config: %v", err)
    }

    // Compare retrieved data with the original data
    if result.Location != conf.Location || result.SecurityZone != conf.SecurityZone {
        t.Errorf("Retrieved config does not match inserted config. Got %+v, want %+v", result, conf)
    }
}

func TestGetMainConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert a test configuration into the database
    testConfig := MainConfiguration{
        DBname:         "testDB",
        Hostname:       "testHost",
        IPAddress:      "192.168.1.1",
        ConfPath:       "/test/path",
        Sync:           true,
        UpdateInterval: 10,
        Hash:           "testhash",
        LastCheck:      time.Now().Unix(),
        LastUpdate:     time.Now().Unix(),
    }

    if err := InsertMainConfig(db, testConfig); err != nil {
        t.Fatalf("Failed to insert test main configuration: %v", err)
    }

    // Test GetMainConfig
    retrievedConfig, err := GetMainConfig(db)
    if err != nil {
        t.Errorf("GetMainConfig failed: %v", err)
    }

    // Compare retrieved data with the original data
    if !reflect.DeepEqual(testConfig, retrievedConfig) {
        t.Errorf("Retrieved main configuration does not match the expected one. Expected: %+v, got: %+v", testConfig, retrievedConfig)
    }
}

func TestGetResolverConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert a sample resolver configuration
    insertSQL := `INSERT INTO config_resolver (Path, PullTimeout, Delimeter, ExtraDelimeter, Hash, LastCheck, LastUpdate) VALUES (?, ?, ?, ?, ?, ?, ?)`
    _, err := db.Exec(insertSQL, "/test/resolver/path", 60, ",", ";", "resolverhash", 1625076000, 1625077000)
    if err != nil {
        t.Fatalf("Failed to insert sample resolver config: %v", err)
    }

    // Test GetResolverConfig
    gotConfig, err := GetResolverConfig(db)
    if err != nil {
        t.Errorf("GetResolverConfig failed: %v", err)
    }

    wantConfig := ResolversConfiguration{
        Path:           "/test/resolver/path",
        PullTimeout:    60,
        Delimeter:      ",",
        ExtraDelimeter: ";",
        Hash:           "resolverhash",
        LastCheck:      1625076000,
        LastUpdate:     1625077000,
    }

    if !reflect.DeepEqual(gotConfig, wantConfig) {
        t.Errorf("GetResolverConfig returned %+v, want %+v", gotConfig, wantConfig)
    }
}

func TestGetLogConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert sample log configurations
    InsertLogConfig(db, LogConfiguration{Path: "/log/path1", MinSeverity: "info", MaxAge: 30, MaxSize: 100, MaxFiles: 5}, 0) // Type 0
    InsertLogConfig(db, LogConfiguration{Path: "/log/path2", MinSeverity: "warn", MaxAge: 60, MaxSize: 200, MaxFiles: 10}, 1) // Type 1

    // Test GetLogConfig for type 0
    logConfig, err := GetLogConfig(db, 0)
    if err != nil {
        t.Errorf("GetLogConfig for type 0 failed: %v", err)
    }
    expectedLogConfig := LogConfiguration{Path: "/log/path1", MinSeverity: "info", MaxAge: 30, MaxSize: 100, MaxFiles: 5}
    if !reflect.DeepEqual(logConfig, expectedLogConfig) {
        t.Errorf("GetLogConfig for type 0 returned %+v, want %+v", logConfig, expectedLogConfig)
    }

    // Test GetLogConfig for type 1
    auditConfig, err := GetLogConfig(db, 1)
    if err != nil {
        t.Errorf("GetLogConfig for type 1 failed: %v", err)
    }
    expectedAuditConfig := LogConfiguration{Path: "/log/path2", MinSeverity: "warn", MaxAge: 60, MaxSize: 200, MaxFiles: 10}
    if !reflect.DeepEqual(auditConfig, expectedAuditConfig) {
        t.Errorf("GetLogConfig for type 1 returned %+v, want %+v", auditConfig, expectedAuditConfig)
    }
}

func TestGetWebServerConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert sample web server configuration
    InsertWebserverConfig(db, WebServerConfiguration{
        Port:          "8080",
        SslIsEnable:   true,
        SslCertPath:   "/ssl/cert",
        SslKeyPath:    "/ssl/key",
        SesionTimeout: 60,
        InitUsername:  "admin",
        InitPassword:  "password",
    })

    // Test GetWebServerConfig
    webServerConfig, err := GetWebServerConfig(db)
    if err != nil {
        t.Fatalf("Failed to get web server config: %v", err)
    }

    expectedConfig := WebServerConfiguration{
        Port:          "8080",
        SslIsEnable:   true,
        SslCertPath:   "/ssl/cert",
        SslKeyPath:    "/ssl/key",
        SesionTimeout: 60,
        InitUsername:  "admin",
        InitPassword:  "password",
    }

    if !reflect.DeepEqual(webServerConfig, expectedConfig) {
        t.Errorf("Retrieved web server config does not match. Expected %+v, got %+v", expectedConfig, webServerConfig)
    }
}

func TestGetPrometheusConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert sample Prometheus configuration
    InsertPrometheusConfig(db, PrometheusConfiguration{
        Url: "http://example.com",
        MetricName: "testMetric",
        Auth: true,
        Username: "user",
        Password: "pass",
        RetriesCount: 3,
        BuferSize: 2,
    })

    // Test GetPrometheusConfig
    promConfig, err := GetPrometheusConfig(db)
    if err != nil {
        t.Fatalf("Failed to get Prometheus config: %v", err)
    }

    expectedConfig := PrometheusConfiguration{
        Url: "http://example.com",
        MetricName: "testMetric",
        Auth: true,
        Username: "user",
        Password: "pass",
        RetriesCount: 3,
        BuferSize: 2,
    }

    if !reflect.DeepEqual(promConfig, expectedConfig) {
        t.Errorf("Retrieved Prometheus config does not match. Expected %+v, got %+v", expectedConfig, promConfig)
    }
}

func TestGetPrometheusLabelConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert sample Prometheus label configuration
    InsertPrometeusLabelsConfig(db, PrometheusLabelConfiguration{
        Opcode: true,
        Authoritative: false,
        // ... other label configurations
    })

    // Test GetPrometheusLabelConfig
    promLabelConfig, err := GetPrometheusLabelConfig(db)
    if err != nil {
        t.Fatalf("Failed to get Prometheus label config: %v", err)
    }

    expectedConfig := PrometheusLabelConfiguration{
        Opcode: true,
        Authoritative: false,
        // ... other label configurations
    }

    if !reflect.DeepEqual(promLabelConfig, expectedConfig) {
        t.Errorf("Retrieved Prometheus label config does not match. Expected %+v, got %+v", expectedConfig, promLabelConfig)
    }
}

func TestGetWatcherConfig(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert sample watcher configuration data
    _, err := db.Exec("INSERT INTO config_watcher (Location, SecurityZone) VALUES (?, ?)", "TestLocation", "TestSecurityZone")
    if err != nil {
        t.Fatalf("Failed to insert sample watcher config: %v", err)
    }

    // Test GetWatcherConfig
    config, err := GetWatcherConfig(db)
    if err != nil {
        t.Fatalf("GetWatcherConfig failed: %v", err)
    }

    // Check if the retrieved configuration matches the inserted data
    if config.Location != "TestLocation" || config.SecurityZone != "TestSecurityZone" {
        t.Errorf("Retrieved configuration does not match expected. Got %+v, want Location: %s, SecurityZone: %s", config, "TestLocation", "TestSecurityZone")
    }
}

func TestGetResolvers(t *testing.T) {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open in-memory db: %v", err)
    }
    defer db.Close()

    if err := InitDB(db); err != nil {
        t.Fatalf("Failed to initialize db: %v", err)
    }

    // Insert test data
    testData := []Resolver{
        {
            Server: "TestServer1",
            IPAddress: "192.168.1.1",
            Domain: "test.com",
            Location: "TestLocation1",
            Site: "TestSite1",
            ServerSecurityZone: "Zone1",
            Prefix: "Prefix1",
            Protocol: "Protocol1",
            Zonename: "ZoneName1",
            Recursion: true,
            QueryCount: 10,
            ServiceMode: true,
        },
        // Add more entries as needed
    }
    InsertResolvers(db, testData)

    // Test GetResolvers
    resolvers, err := GetResolvers(db)
    if err != nil {
        t.Fatalf("Failed to get resolvers: %v", err)
    }

    // Define expected data
    expected := []Resolver{
        {
            Server: "TestServer1",
            IPAddress: "192.168.1.1",
            Domain: "test.com",
            Location: "TestLocation1",
            Site: "TestSite1",
            ServerSecurityZone: "Zone1",
            Prefix: "Prefix1",
            Protocol: "Protocol1",
            Zonename: "ZoneName1",
            Recursion: true,
            QueryCount: 10,
            ServiceMode: true,
        },
        // Add more entries as needed
    }

    // Compare expected and actual data
    if !reflect.DeepEqual(resolvers, expected) {
        t.Errorf("Expected resolvers %+v, got %+v", expected, resolvers)
    }
}

func TestUpdateLastCheck(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert sample data into config_main
    insertSampleMainConfig(db, t)

    dbname := "testDB"
    newLastCheck := time.Now().Unix()

    // Call UpdateLastCheck
    if err := UpdateLastCheck(db, dbname, newLastCheck); err != nil {
        t.Fatalf("UpdateLastCheck failed: %v", err)
    }

    // Query the database to verify the update
    var lastCheck int64
    err := db.QueryRow("SELECT LastCheck FROM config_main WHERE DBname = ?", dbname).Scan(&lastCheck)
    if err != nil {
        t.Fatalf("Failed to query updated last check: %v", err)
    }

    if lastCheck != newLastCheck {
        t.Errorf("Expected LastCheck to be %d, got %d", newLastCheck, lastCheck)
    }
}

func insertSampleMainConfig(db *sql.DB, t *testing.T) {
    _, err := db.Exec(`INSERT INTO config_main (DBname, LastCheck) VALUES (?, ?)`, "testDB", time.Now().Unix())
    if err != nil {
        t.Fatalf("Failed to insert sample main config: %v", err)
    }
}

func TestUpsertSyncMember(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Prepare a sample sync member configuration
    member := MemberConfiguration{
        SyncID:          1,
        Hostname:        "testHost",
        Port:            "8080",
        SeverLastCheck:  time.Now().Unix(),
        ConfigHash:      "hash123",
        ConfigLastCheck: time.Now().Unix(),
        ConfigLastUpdate: time.Now().Unix(),
        ResolvHash:      "resolveHash123",
        ResolvLastCheck: time.Now().Unix(),
        ResolvLastUpdate: time.Now().Unix(),
        IsLocal:         true,
    }

    // Call UpsertSyncMember
    if err := UpsertSyncMember(db, member); err != nil {
        t.Fatalf("UpsertSyncMember failed: %v", err)
    }

    // Query the database to verify the insertion
    var retrievedMember MemberConfiguration
    err := db.QueryRow("SELECT sync_id, hostname, port, SeverLastCheck, ConfigHash, ConfigLastCheck, ConfigLastUpdate, ResolvHash, ResolvLastCheck, ResolvLastUpdate, IsLocal FROM config_sync_members WHERE hostname = ?", member.Hostname).Scan(&retrievedMember.SyncID, &retrievedMember.Hostname, &retrievedMember.Port, &retrievedMember.SeverLastCheck, &retrievedMember.ConfigHash, &retrievedMember.ConfigLastCheck, &retrievedMember.ConfigLastUpdate, &retrievedMember.ResolvHash, &retrievedMember.ResolvLastCheck, &retrievedMember.ResolvLastUpdate, &retrievedMember.IsLocal)
    if err != nil {
        t.Fatalf("Failed to query sync member: %v", err)
    }

    // Compare retrieved data with the original data
    if !reflect.DeepEqual(retrievedMember, member) {
        t.Errorf("Retrieved member does not match inserted member. Got %+v, want %+v", retrievedMember, member)
    }
}

func TestUpdateSyncMemberTimestamps(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()
    // Insert a sample sync member
    insertSyncMember(db, t, MemberConfiguration{
        SyncID:          1,
        Hostname:        "testHost",
        Port:            "8080",
        SeverLastCheck:  time.Now().Unix(),
        ConfigHash:      "hash123",
        ConfigLastCheck: time.Now().Unix(),
        ConfigLastUpdate: time.Now().Unix(),
        ResolvHash:      "resolveHash123",
        ResolvLastCheck: time.Now().Unix(),
        ResolvLastUpdate: time.Now().Unix(),
        IsLocal:         true,
        SyncEnable:      true,
    })
    // Define the timestamps for update
    newLastCheck := time.Now().Unix()
    newLastUpdate := time.Now().Unix()

    // Update the sync member timestamps
    err := UpdateSyncMemberTimestamps(db, "testHost", newLastCheck, newLastUpdate, newLastCheck, newLastUpdate, newLastCheck, true)
    if err != nil {
        t.Fatalf("Failed to update sync member timestamps: %v", err)
    }

    // Query the database to verify the update
    var member MemberConfiguration
    err = db.QueryRow("SELECT SeverLastCheck, ConfigLastCheck, ConfigLastUpdate, ResolvLastCheck, ResolvLastUpdate, SyncEnable FROM memebers_view WHERE hostname = ?", "testHost").Scan(&member.SeverLastCheck, &member.ConfigLastCheck, &member.ConfigLastUpdate, &member.ResolvLastCheck, &member.ResolvLastUpdate, &member.SyncEnable)
    if err != nil {
        t.Fatalf("Failed to get sync member: %v", err)
    }

    // Check if the timestamps are updated correctly
    if member.SeverLastCheck != newLastCheck || member.ConfigLastCheck != newLastCheck || member.ConfigLastUpdate != newLastUpdate || member.ResolvLastCheck != newLastCheck || member.ResolvLastUpdate != newLastUpdate {
        t.Errorf("Sync member timestamps not updated correctly. Got %+v", member)
    }
}

func insertSyncMember(db *sql.DB, t *testing.T, member MemberConfiguration) {
    _, err := db.Exec("INSERT INTO config_sync_members (sync_id, hostname, port, SeverLastCheck, ConfigHash, ConfigLastCheck, ConfigLastUpdate, ResolvHash, ResolvLastCheck, ResolvLastUpdate, IsLocal) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", member.SyncID, member.Hostname, member.Port, member.SeverLastCheck, member.ConfigHash, member.ConfigLastCheck, member.ConfigLastUpdate, member.ResolvHash, member.ResolvLastCheck, member.ResolvLastUpdate, member.IsLocal)
    if err != nil {
        t.Fatalf("Failed to insert sync member: %v", err)
    }
}

func TestUpdateMainConfigTimestamps(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert a sample main configuration
    InsertMainConfig(db, MainConfiguration{
        DBname:         "testDB",
        Hostname:       "testHost",
        IPAddress:      "192.168.1.1",
        ConfPath:       "/test/path",
        Sync:           true,
        UpdateInterval: 10,
        Hash:           "testhash",
        LastCheck:      time.Now().Unix(),
        LastUpdate:     time.Now().Unix(),
    })

    // Define new timestamps for update
    newLastCheck := time.Now().Unix()
    newLastUpdate := time.Now().Unix()

    // Update the main configuration timestamps
    err := UpdateMainConfigTimestamps(db, "testDB", newLastCheck, newLastUpdate)
    if err != nil {
        t.Fatalf("Failed to update main config timestamps: %v", err)
    }

    // Query the database to verify the update
    var config MainConfiguration
    err = db.QueryRow("SELECT LastCheck, LastUpdate FROM config_main WHERE DBname = ?", "testDB").Scan(&config.LastCheck, &config.LastUpdate)
    if err != nil {
        t.Fatalf("Failed to get main config: %v", err)
    }

    // Check if the timestamps are updated correctly
    if config.LastCheck != newLastCheck || config.LastUpdate != newLastUpdate {
        t.Errorf("Main config timestamps not updated correctly. Got LastCheck: %v, LastUpdate: %v, Want LastCheck: %v, LastUpdate: %v", config.LastCheck, config.LastUpdate, newLastCheck, newLastUpdate)
    }
}

func TestUpdateResolversConfigTimestamps(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert a sample resolver configuration
    InsertResolversConfig(db, ResolversConfiguration{
        Path:           "test/path",
        PullTimeout:    60,
        Delimeter:      ",",
        ExtraDelimeter: ";",
        Hash:           "resolverhash",
        LastCheck:  time.Now().Unix(),
        LastUpdate: time.Now().Unix(),
    })

    // Define new timestamps for update
    newLastCheck := time.Now().Unix()
    newLastUpdate := time.Now().Unix()

    // Update the resolver configuration timestamps
    err := UpdateResolversConfigTimestamps(db, "test/path", newLastCheck, newLastUpdate)
    if err != nil {
        t.Fatalf("Failed to update resolver config timestamps: %v", err)
    }

    // Query the database to verify the update
    var config ResolversConfiguration
    err = db.QueryRow("SELECT LastCheck, LastUpdate FROM config_resolver WHERE Path = ?", "test/path").Scan(&config.LastCheck, &config.LastUpdate)
    if err != nil {
        t.Fatalf("Failed to get resolver config: %v", err)
    }

    // Check if the timestamps are updated correctly
    if config.LastCheck != newLastCheck || config.LastUpdate != newLastUpdate {
        t.Errorf("Resolver config timestamps not updated correctly. Got LastCheck: %v, LastUpdate: %v, Want LastCheck: %v, LastUpdate: %v", config.LastCheck, config.LastUpdate, newLastCheck, newLastUpdate)
    }
}
