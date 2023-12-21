package config

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)


func ReadResolversFromCSV(conf sqldb.ResolversConfiguration) ([]sqldb.Resolver, error) {
    var (
		delimeter rune
		servers []sqldb.Resolver
	)
	if len(conf.Delimeter) > 0 {
		delimeter = rune(conf.Delimeter[0])
	} else {
		return []sqldb.Resolver{}, errors.New("string is empty, cannot parse to rune")
	}

	resolversFromCsv := []sqldb.Csv{}
    clientsFile, err := os.OpenFile(conf.Path, os.O_RDWR, os.ModePerm)
	if err != nil {
        return []sqldb.Resolver{}, err
	}
	defer clientsFile.Close()
    
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
        r := csv.NewReader(in)
        r.LazyQuotes = true
        r.Comma = delimeter
        return r
    })
	if err := gocsv.UnmarshalFile(clientsFile, &resolversFromCsv); err != nil {
		// log.AppLog.Error("Error Unmarshal file ", conf.Path,"| error: ", err)
        return []sqldb.Resolver{}, fmt.Errorf("error Unmarshal file %s : %v", conf.Path, err)
	}
    
    servers, err = parseZones(resolversFromCsv, conf.ExtraDelimeter)
    if err != nil {
        // log.AppLog.Error("Error parse zones '", conf.Path, "'. error:", err)
        return []sqldb.Resolver{}, fmt.Errorf("error parse zones  %s : %v", conf.Path, err)
    }
    
    return servers, nil
}

func parseZones(records []sqldb.Csv, extraDelimeter string) ([]sqldb.Resolver, error) {
    var resolvers []sqldb.Resolver
    for _, record := range records {
        
        zoneNames :=  strings.Split(record.Zonename, extraDelimeter)
        queryRPSs := strings.Split(record.QueryCount, extraDelimeter)
        service_mode, err := strconv.ParseBool(record.ServiceMode)
        if err != nil {
            // log.AppLog.Warning("Warning: Error parse server mode value for server: '", record.Server, "', value 'service_mode': ", record.Service_mode, "err:", err)
            service_mode = false
        }
        for i, zonename := range zoneNames {
            if zonename == "" {
                continue
            }
            queryRPSInt, err := strconv.Atoi(queryRPSs[i])
            if err != nil {
                // log.AppLog.Warning("Warning: Error parse query count rps value for server: '", record.Server, "', value 'query_count_rps': ", record.Query_count_rps, "err:", err)
                queryRPSInt = 5
            }
            resolver := sqldb.Resolver{
                Server: record.Server,
                IPAddress: record.IPAddress,
                Domain: record.Domain,
                Location: record.Location,
                Site: record.Site,
                ServerSecurityZone: record.ServerSecurityZone,
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
                // log.AppLog.Warning("Warning: Error parse query count rps value for server: '", record.Server, "', value 'query_count_rps': ", record.Query_count_rps, "err:", err)
                queryRPSInt = 2
            }
            resolver := sqldb.Resolver{
                Server: record.Server,
                IPAddress: record.IPAddress,
                Domain: record.Domain,
                Location: record.Location,
                Site: record.Site,
                ServerSecurityZone: record.ServerSecurityZone,
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