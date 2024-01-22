package datastore

import (
	"HighFrequencyDNSChecker/components/logger"
	"HighFrequencyDNSChecker/components/tools"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
)

type Csv struct {
	Server					string `csv:"server"`
	IPAddress				string `csv:"server_ip"`
	Domain					string `csv:"domain"`
    Location				string `csv:"location"`
    Site					string `csv:"site"`
    ServerSecurityZone		string `csv:"server_security_zone"`
    Prefix					string `csv:"prefix"`
    Protocol				string `csv:"protocol"`
    Zonename				string `csv:"zonename"`
    QueryCount				string `csv:"query_count_rps"`
    ZonenameWithRecursion	string `csv:"zonename_with_recursion"`
    QueryCountWithRecursion	string `csv:"query_count_with_recursion_rps"`
    ServiceMode				string `csv:"service_mode"`
}

type PollingHost struct {
    Hostname 		string
	IPAddress 		string
	Domain 			string
    Location 		string
    Site 			string
    SecurityZone	string
    Prefix 			string
    Protocol 		string
    Zonename 		string
    Recursion 		bool
    QueryCount 		int
    ServiceMode 	bool
}

var (
	pollingHosts []PollingHost
    lastCsvHash HashStruct
	pollingHostsMutex sync.RWMutex

)

func ReadResolversFromCSV() (bool, error) {
    var (
		delimeter rune
		servers []PollingHost
	)
	conf := GetConfig().Polling
	if len(conf.Delimeter) > 0 {
		delimeter = rune(conf.Delimeter[0])
	} else {
		return false, errors.New("string is empty, cannot parse to rune")
	}

    fileHash, err := tools.CalculateHash(string(conf.Path))
    if err != nil {
        logger.Logger.Errorf("Error Calculate hash to file '%s' err: %v\n", configFile, err)
        return false, err
    }

    if lastCsvHash.LastHash == fileHash {
        logger.Logger.Debug("CSV file has not been changed")
        return false, nil
    }
    logger.Logger.Infof("CSV file has been changed")

	resolversFromCsv := []Csv{}
    clientsFile, err := os.OpenFile(conf.Path, os.O_RDWR, os.ModePerm)
	if err != nil {
        return false, err
	}
	defer clientsFile.Close()
    
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
        r := csv.NewReader(in)
        r.LazyQuotes = true
        r.Comma = delimeter
        return r
    })
	if err := gocsv.UnmarshalFile(clientsFile, &resolversFromCsv); err != nil {
        return false, fmt.Errorf("error Unmarshal file %s : %v", conf.Path, err)
	}

	uniqueChecker := make(map[string]bool)
    var duplicates []string
    for _, record := range resolversFromCsv {
        key := record.Server + "-" + record.IPAddress + "-" + record.Domain
        if _, exists := uniqueChecker[key]; exists {
            // Collect information about the duplicate
            duplicates = append(duplicates, key)
            continue
        }
        uniqueChecker[key] = true
    }

    if len(duplicates) > 0 {
        // Return an error with details of duplicates
        return false, fmt.Errorf("duplicates found in CSV file: %s", strings.Join(duplicates, ", "))
    }
    
    servers, err = parseZones(resolversFromCsv, conf.ExtraDelimeter)
    if err != nil {
        return false, fmt.Errorf("error parse zones  %s : %v", conf.Path, err)
    }
	pollingHostsMutex.Lock()
    defer pollingHostsMutex.Unlock()
	pollingHosts = servers
    lastCsvHash.LastHash = fileHash
    lastCsvHash.LastUpdate = time.Now().Unix()
    return true, nil
}

func parseZones(records []Csv, extraDelimeter string) ([]PollingHost, error) {
    var resolvers []PollingHost
    for _, record := range records {
        
        zoneNames :=  strings.Split(record.Zonename, extraDelimeter)
        queryRPSs := strings.Split(record.QueryCount, extraDelimeter)
        service_mode, err := strconv.ParseBool(record.ServiceMode)
        if err != nil {
            service_mode = false
        }
        for i, zonename := range zoneNames {
            if zonename == "" {
                continue
            }
            queryRPSInt, err := strconv.Atoi(queryRPSs[i])
            if err != nil {
                queryRPSInt = 5
            }
            resolver := PollingHost{
                Hostname: record.Server,
                IPAddress: record.IPAddress,
                Domain: record.Domain,
                Location: record.Location,
                Site: record.Site,
                SecurityZone: record.ServerSecurityZone,
                Prefix: record.Prefix,
                Protocol: record.Protocol,
                Zonename: zonename,
                Recursion: false,
                QueryCount: queryRPSInt,
                ServiceMode: service_mode,
            }
            resolvers = append(resolvers, resolver)
        }

        zoneNamesRecursion :=  strings.Split(record.ZonenameWithRecursion, extraDelimeter)
        queryRPSsRecursion := strings.Split(record.QueryCountWithRecursion, extraDelimeter)
        for i, zonename := range zoneNamesRecursion {
            if zonename == "" {
                continue
            }
            queryRPSInt, err := strconv.Atoi(queryRPSsRecursion[i])
            if err != nil {
                queryRPSInt = 2
            }
            resolver := PollingHost{
                Hostname: record.Server,
                IPAddress: record.IPAddress,
                Domain: record.Domain,
                Location: record.Location,
                Site: record.Site,
                SecurityZone: record.ServerSecurityZone,
                Prefix: record.Prefix,
                Protocol: record.Protocol,
                Zonename: zonename,
                Recursion: true,
                QueryCount: queryRPSInt,
                ServiceMode: service_mode,
            }
            resolvers = append(resolvers, resolver)
        }
    }
    return resolvers, nil
}

func GetPollingHosts() []PollingHost {
    configMutex.RLock()
    defer configMutex.RUnlock()
    copiedHosts := make([]PollingHost, len(pollingHosts))
    copy(copiedHosts, pollingHosts)

    return copiedHosts
}

func GetCSVHash() HashStruct{
	configMutex.RLock()
    defer configMutex.RUnlock()
    return lastCsvHash
}