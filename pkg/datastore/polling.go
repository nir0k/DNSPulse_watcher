package datastore

import (
	"DNSPulse_watcher/pkg/logger"
	"fmt"
	"strconv"
	"strings"
	"sync"

	pb "DNSPulse_watcher/pkg/gRPC"
)


type PollingHostStruct struct {
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
	pollingHosts []PollingHostStruct
    lastPollingHash string
	pollingHostsMutex sync.RWMutex
)

func GetCSVHash() string{
	pollingHostsMutex.RLock()
    defer pollingHostsMutex.RUnlock()
    return lastPollingHash
}

func LoadPollingHosts(data []*pb.Csv, newConfHash string) (bool, error) {

    var (
        hosts []PollingHostStruct
        csvHash string
    )
    conf := GetSegmentConfig().Polling
    csvHash = GetCSVHash()
    fmt.Printf("New hash: %s", newConfHash)
    if newConfHash == csvHash {
        logger.Logger.Debugf("PollingHosts not changes")
        return false, nil
    }
    
    uniqueChecker := make(map[string]bool)
    var duplicates []string
    for _, record := range data {
        key := record.Server + "-" + record.IpAddress + "-" + record.Domain
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
    
    hosts, err := parseZones(data, conf.ExtraDelimiter)
    if err != nil {
        return false, fmt.Errorf("error parse zones  %s : %v", conf.Path, err)
    }
	pollingHostsMutex.Lock()
    defer pollingHostsMutex.Unlock()
	pollingHosts = hosts
    lastPollingHash = newConfHash
    SetSegmentPollingHash(newConfHash)
    
    return true, nil
}

func parseZones(records []*pb.Csv, extraDelimeter string) ([]PollingHostStruct, error) {
    var resolvers []PollingHostStruct
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
            resolver := PollingHostStruct{
                Hostname: record.Server,
                IPAddress: record.IpAddress,
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
            resolver := PollingHostStruct{
                Hostname: record.Server,
                IPAddress: record.IpAddress,
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

func GetPollingHosts() *[]PollingHostStruct {
    pollingHostsMutex.RLock()
    defer pollingHostsMutex.RUnlock()
    return &pollingHosts
}