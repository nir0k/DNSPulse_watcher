package config

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	"os"
	"strconv"
	"testing"
)

func TestReadResolversFromCSV(t *testing.T) {
    // Sample CSV content
    content := `
server,server_ip,service_mode,domain,prefix,location,site,server_security_zone,protocol,zonename,query_count_rps,zonename_with_recursion,query_count_with_recursion_rps
TestServer,127.0.0.5,true,REG,dnsmon,DP4,null,REGION-LAN,udp,reg.ru,2,msk.ru&test.ru,2&2
new_server2,1.2.3.4,false,newdomain.com,prefix,location,site,zone,udp,zone1,10,test.ru&region.test2.ru,1&2
`

    // Create a temporary CSV file
    tmpfile, err := os.CreateTemp("", "example.*.csv")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name()) // clean up

    if _, err := tmpfile.WriteString(content); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    // Test configuration
    conf := sqldb.ResolversConfiguration{
        Path:           tmpfile.Name(),
        Delimeter:      ",",
        ExtraDelimeter: "&",
    }

    // Call the function under test
    resolvers, err := ReadResolversFromCSV(conf)
    if err != nil {
        t.Fatalf("ReadResolversFromCSV() error = %v", err)
    }

    // Expected length of the resolvers slice
    expectedLength := 6 // Adjust according to expected number of resolvers parsed from CSV
    if len(resolvers) != expectedLength {
        t.Errorf("Expected %d resolvers, got %d", expectedLength, len(resolvers))
    }

    // Asserting individual fields (adjust according to your CSV content)
    if resolvers[0].Server != "TestServer" {
        t.Errorf("Expected Server 'TestServer', got '%s'", resolvers[0].Server)
    }
    // ... Add more assertions for other fields and resolvers
}

func TestParseZones(t *testing.T) {
    records := []sqldb.Csv{
        {
            Server:                  "TestServer",
            IPAddress:               "127.0.0.5",
            ServiceMode:             "true",
            Domain:                  "REG",
            Prefix:                  "dnsmon",
            Location:                "DP4",
            Site:                    "null",
            ServerSecurityZone:      "REGION-LAN",
            Protocol:                "udp",
            Zonename:                "reg.ru",
            QueryCount:              "2",
            ZonenameWithRecursion:   "msk.ru&test.ru",
            QueryCountWithRecursion: "2&2",
        },
        {
            Server:                  "new_server2",
            IPAddress:               "1.2.3.4",
            ServiceMode:             "false",
            Domain:                  "newdomain.com",
            Prefix:                  "prefix",
            Location:                "location",
            Site:                    "site",
            ServerSecurityZone:      "zone",
            Protocol:                "udp",
            Zonename:                "zone1",
            QueryCount:              "10",
            ZonenameWithRecursion:   "test.ru&region.test2.ru",
            QueryCountWithRecursion: "1&2",
        },
        // Add more records as needed
    }

    resolvers, err := parseZones(records, "&")
    if err != nil {
        t.Fatalf("ParseZones() error = %v", err)
    }

    expectedLength := 6 // Adjust based on your expected number of Resolver entries
    if len(resolvers) != expectedLength {
        t.Errorf("Expected %d resolvers, got %d", expectedLength, len(resolvers))
    }

    // Test first resolver from the first record
    if resolvers[0].Server != "TestServer" || resolvers[0].IPAddress != "127.0.0.5" {
        t.Errorf("Incorrect data for first resolver: got %+v", resolvers[0])
    }

    // Test if ServiceMode is correctly parsed
    expectedServiceMode, _ := strconv.ParseBool(records[0].ServiceMode)
    if resolvers[0].ServiceMode != expectedServiceMode {
        t.Errorf("Expected ServiceMode %v, got %v for first resolver", expectedServiceMode, resolvers[0].ServiceMode)
    }

    // Add more assertions as needed for other fields or other resolvers
}