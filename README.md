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


Available parameters in configurations:

- DNS settings:
  - `DNS_RESOLVERPATH` - Path to file with list of DNS servers
  - `DNS_HOSTPOSTFIX` - Prefix hostname to resolve. Example: test.local.com
  - `DNS_POLLING_RATE_NO_RECURSION` - polling rate without recusrion in number of checks per second
  - `DNS_POLLING_RATE_RECURSION` - polling rate with recusrion in number of checks per second
  - `DNS_TIMEOUT` - DNS answer timeout in seconds
  - `DNS_PROTOCOL` - Protocol. Possible value: tcp, udp, udp4, udp6, tcp4, tcp6
  - `DNS_CHECK_HOST_INCLUDE` - Add checking host in metric label. Possible value: true or false

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

- Watcher settings:
  - `LOG_FILE` - Path to log file
  - `LOG_LEVEL` - Minimal severity level for logging. Possible values: debug, info, warning, error, fatal (default: warning)
  - `CONF_CHECK_INTERVAL` - Interval check changes in config in minutes
  - `BUFFER_SIZE` - Timeseries buffer size for sent to prometheus
  - `WATCHER_LOCATION` - Watcher location
  - `WATCHER_SECURITYZONE` - Watcher security zone
  - `DUBLICATE_ALLOW` - Allow or not dublicate record