# High-Frequency DNS Resolution Program for Prometheus Integration
Program Overview:
Our cutting-edge program is designed to provide high-frequency DNS name resolution for server names from a predefined list. This data is then efficiently recorded in Prometheus, empowering users with real-time insights into their server infrastructure's performance and availability.

Key Program Features:
- High-Frequency Polling: Fast and frequent DNS polling (default: every 150ms).
- Dynamic Server Naming: Random server names with timestamps.
- IP Address Verification: Ensures that IP address 1.1.1.1 is resolved for each server name.
- Continuous Monitoring: Operates endlessly in a loop for uninterrupted monitoring.
- Records resolved server names, resolve time, rcode, protocol, truncated flag and timestamps in Prometheus for historical analysis.
- Runs in an infinite loop for continuous monitoring.

### Install
**requred go 1.21.1**
```bash
git clone https://github.com/nir0k/HighFrequencyDNSChecker.git
cd HighFrequencyDNSChecker
go build .
```

### Prepare
- Create and fill out the file `.env` in the same folder with the program. Complited example we found in the file `.env-example` in the project
- Create and fill out csv-file with DNS servers information. Complited example we found in the file `dns_servers.csv` in the project. If you using diferen filename, you are need change it in .env file


### Use
```bash
./HighFrequencyDNSChecker
```

### Custom Rcode:
- 3841 - Resolved IP-address not equals 1.1.1.1
- 3842 - DNS Server not answer


## Available parameters in configurations:

- DNS settings:
  - `DNS_RESOLVERPATH` - Path to file with list of DNS servers
  - `DNS_TIMEOUT` - DNS answer timeout in seconds

- Prometheus settings:
  - `PROM_URL` - Prometheus remote write url. example: http://prometheus:8428/api/v1/write
  - `PROM_METRIC` - Prometheus metric name
  - `PROM_AUTH` - Prometheus authentication. false or true. If true, values PROM_USER and PROM_PASS are required
  - `PROM_USER` - Prometheus username
  - `PROM_PASS` - Prometheus password
  - `PROM_RETRIES` - Count retries for post data in prometheus
  - Labels: 
    - `OPCODES` - OpCodes. Possible value: true or false
    - `AUTHORITATIVE` - Authoritative. Possible value: true or false
    - `TRUNCATED` - Truncated. Possible value: true or false
    - `RCODE` - Rcode. Possible value: true or false
    - `RECURSION_DESIRED` - RecursionDesired. Possible value: true or false
    - `RECURSION_AVAILABLE` - RecursionAvailable. Possible value: true or false
    - `AUTHENTICATE_DATA` - AuthenticatedData. Possible value: true or false
    - `CHECKING_DISABLED` - CheckingDisabled. Possible value: true or false
    - `POLLING_RATE` - Polling rate. Possible value: true or false
    - `RECURSION` - Recursion. Request with reqursion or not. Possible value: true or false

- Watcher settings:
  - `LOG_FILE` - Path to log file
  - `LOG_LEVEL` - Minimal severity level for logging. Possible values: debug, info, warning, error, fatal (default: warning)
  - `CONF_CHECK_INTERVAL` - Interval check changes in config in minutes
  - `BUFFER_SIZE` - Timeseries buffer size for sent to prometheus
  - `WATCHER_LOCATION` - Watcher location
  - `WATCHER_SECURITYZONE` - Watcher security zone
  

## CSV structure:

**Delimeter: `,` (comma)**


Example CSV:
```csv
server,server_ip,domain,suffix,location,site,protocol,zonename,query_count_rps,zonename_with_recursion,query_count_with_recursion_rps
google_dns_1,8.8.8.8,google.com1,dnsmon,testloc1,testsite1,udp,testzone1,5,testzone1_r,2
```

### Field descriptiom:

 - `server` - DNS Server name. Value type: String
 - `server_ip` - DNS Server IP address. Value type: String
 - `domain` - Domain. Value type: String
 - `suffix` - Suffix for create dunamic hostname fo resolve. Hostname create by this rule: `<unixtime with nanoseconds>.<suffix>.<zonename>`
 - `location` - DNS Server Location. Value type: String
 - `site` - DNS Server Site. Value type: String
 - `protocol` - Protocol Used for polling. Value type: String. Possible value: tcp, udp, udp4, udp6, tcp4, tcp6
 - `zonename` - DNS Zonename without recusrion. Value type: String
 - `query_count_rps` - Count request per secconds for DSN server polling without recursion. Value type: Intenger
 - `zonename_with_recursion` - DNS Zonename with recusrion. Value type: String
 - `query_count_with_recursion_rps` - Count request per secconds for DSN server polling with recursion. Value type: Intenger