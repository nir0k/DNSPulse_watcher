# HighFrequencyDNSChecker

Program for high-frequency resolution of server names on DNS servers from a list and recording the result in Prometheus.

Program Features:
- High-frequency polling (default every 150ms).
- A random server name is queried with a fixed prefix and it is checked that the IP address 1.1.1.1 is resolved. Server name create on this rule: `<timestamp with miliseconds>.<hostname prefix>`
- Work in an infinite loop.

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
`DNS_RESOLVERPATH` - Path to file with list of DNS servers
`DNS_HOSTPOSTFIX` - Prefix hostname to resolve. Example: test.local.com
`DNS_POLLING_RATE_NO_RECURSION` - polling rate without recusrion in number of checks per second
`DNS_POLLING_RATE_RECURSION` - polling rate with recusrion in number of checks per second
`DNS_TIMEOUT` - DNS answer timeout in seconds
`DNS_PROTOCOL` - Protocol. Possible value: tcp, udp, udp4, udp6, tcp4, tcp6

- Prometheus settings:
`PROM_URL` - Prometheus remote write url. example: http://prometheus:8428/api/v1/write
`PROM_METRIC` - Prometheus metric name
`PROM_AUTH` - Prometheus authentication. false or true. If true, values PROM_USER and PROM_PASS are required
`PROM_USER` - Prometheus username
`PROM_PASS` - Prometheus password
`PROM_RETRIES` - Count retries for post data in prometheus

- Watcher settings:
`LOG_FILE` - Path to log file
`LOG_LEVEL` - Minimal severity level for logging. Possible values: debug, info, warning, error, fatal (default: warning)
`CONF_CHECK_INTERVAL` - Interval check changes in config in minutes
`BUFFER_SIZE` - Timeseries buffer size for sent to prometheus
`WATCHER_LOCATION` - Watcher location
`WATCHER_SECURITYZONE` - Watcher security zone
`DUBLICATE_ALLOW` - Allow or not dublicate record