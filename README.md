# High-Frequency DNS Server Polling Utility

## Description:
This utility is designed for high-frequency verification of DNS server name resolutions from a pre-determined list. The results of these checks are saved to a Prometheus database using the remote write feature, ensuring access to data in real-time. This provides users with comprehensive tools for analyzing the efficiency and availability of server infrastructure, significantly enhancing their ability to effectively monitor and maintain system integrity.

## Key features of the utility:
- High-frequency polling (default every 150 ms)
- Dynamic formation of the name for DNS server resolution. (A prefix in the form of unixtime with milliseconds is added)
- Verification of the resolved address (ensures that the address 1.1.1.1 was resolved)
- Continuous monitoring. (implies 24/7 operation)
- Ability to select labels to send in the configuration
- Dynamic updating of the configuration from a file
- Dynamic updating of the list of servers polled from a csv file
- Synchronization of configurations and server lists with neighboring servers (list of servers and authorization token configured in the configuration file)
- Web interface:
  - Ability to view configuration, list of polled servers, latest application logs and audits, and synchronization status with other servers
  - Capability to edit the server list by uploading a file or line-by-line
  - Option to change configuration through the web interface
  - API documentation
- API:
  - Hash of configuration and server list files, and date of last modification
  - Retrieval of the configuration file
  - Retrieval of the server list file
- Logging of utility operations with a choice of logging level
- Audit log recording
- Log rotation according to set rules

### Requirements for running the utility:
**Attention, the utility's operation was only tested on MacOS and Linux operating systems**

- Configuration file in yaml format
- File with the list of servers being polled

### Example configuration file:
```yaml
General:
  confCheckInterval: 1  # Configuration check frequency (in minutes)
  location: K2          # Location of the server with the utility
  securityZone: PROD    # Security zone of the server with the utility

Audit:
  path: "audit.json"    # Path to the audit file
  minSeverity: "info"   # Minimum severity level
  maxAge: 30            # Maximum file lifespan (in days)
  maxSize: 10           # Maximum file size (in MB)
  maxFiles: 10          # Maximum number of files

WebServer:
  listenAddress: 0.0.0.0    # The interface on which the web-server will work
  port: 443                 # Port on which the web server will run
  sslIsEnable: true         # Enable or disable HTTPS (currently not working)
  sslCertPath: "cert.pem"   # Path to the certificate file
  sslKeyPath: "key.pem"     # Path to the private key file (should not be password protected)
  sesionTimeout: 600        # User session timeout (in seconds)
  username: "admin"         # User for web interface login
  password: "password"      # Password for connection

Sync:
  isEnable: true                    # Enable synchronization with neighbors
  token: "fvdknlvd9ergturoegkvnemc" # Token for synchronization (unlike the user token, it does not expire)
  SyncMembers:                          # List of neighbors to synchronize with
    - hostname: "127.0.0.1"
      port: 443
    - hostname: "10.10.10.10"
      port: 8081

Prometheus:                                     # Prometheus settings
  url: "http://prometheus:8428/api/v1/write"    # URL to connect to Prometheus
  metricName: "dns_resolve"                     # Metric name
  auth: false                                   # Enable/disable authorization
  username: "user"                              # Prometheus user for writing data to the DB
  password: "password"                          # Prometheus password
  retriesCount: 2                               # Number of attempts to send metrics
  buferSize: 2                                  # Metrics buffer size (how many metrics will be collected before sending to Prometheus)
  labels:                                       # Enable or disable additional labels (current settings reflected)
    opcode: false
    authoritative: false
    truncated: true
    rcode: true
    recursionDesired: false
    recursionAvailable: false
    authenticatedData: false
    checkingDisabled: false
    pollingRate: false
    recursion: true

Resolvers:                  
  path: "dns_servers.csv"   # Path to the file with the list of servers being polled
  pullTimeout: 2            # Maximum response wait time (in seconds)
  delimeter: ","            # Main delimiter in the CSV file
  extraDelimeter: "&"       # Additional delimiter for fields server_security_zone, query_count_rps, zonename_with_recursion, query_count_with_recursion_rps
```

### Example of a CSV file with a list of servers
```csv
server,server_ip,service_mode,domain,prefix,location,site,server_security_zone,protocol,zonename,query_count_rps,zonename_with_recursion,query_count_with_recursion_rps
TestServer,127.0.0.5,true,REG,dnsmon,DP4,null,REGION-LAN,udp,reg.ru,2,msk.ru&test.ru,2&2
new_server2,1.2.3.4,false,newdomain.com,prefix,location,site,zone,udp,zone1,10,test.ru&region.test2.ru,1&2
TestServe3,127.0.0.99,true,REG,dnsmon,DP4,null,REGION-LAN,udp,reg.ru,2,msk.ru&region.test.ru&test.com,2&3&3
Dublicate ,1.2.3.5,false,newdomain.com,prefix,location,site,zone,udp,zone1,10,test.ru,1
Dublicate ,1.2.3.5,false,newdomain.com,prefix,location,site,zone,udp,zone1,10,test.ru,1
```
- `server` - The name of the server, which will be displayed in labels (does not affect polling)
- `server_ip` - The IP address of the server, which is used to connect to the server
- `service_mode` - Service mode, servers with service mode enabled are not polled, once a second a metric with data indicating the server is in service mode is sent
- `domain` - Domain
- `prefix` - The prefix to which a random part is added for polling. The formation pattern: <unixtime with nanoseconds>.<suffix>.<zonename>
- `location` - The location of the DNS server
- `site` - The site of the DNS server
- `server_security_zone` - The server's security zone
- `protocol` - The protocol by which the polling will be conducted (udp, tcp, udp4, tcp4, udp6, tcp6)
- `zonename` - The security zone being polled without recursion; if there are multiple zones, they should be entered in one line using the "extraDelimeter" value from the configuration as a separator
- `query_count_rps` - The frequency of polling without recursion; if there are multiple zones, the polling values should be entered in one line using the "extraDelimeter" value from the configuration as a separator
- `zonename_with_recursion` - The security zone being polled with recursion; if there are multiple zones, they should be entered in one line using the "extraDelimeter" value from the configuration as a separator
- `query_count_with_recursion_rps` - The frequency of polling with recursion; if there are multiple zones, the polling values should be entered in one line using the "extraDelimeter" value from the configuration as a separator


### Compiling the utility:
```bash
make build
```
The compiled package will be available in `bin/HighFrequencyDNSChecker-linux-amd64`

### Launch:
To launch the utility, place the executable file in the same directory as:
- The configuration file `config.yaml`
- The server list file referred to by `Resolvers->path` in the configuration
- The web server certificate file referred to by `WebServer->sslCertPath` in the configuration
- The certificate key file referred to by `WebServer->sslKeyPath` in the configuration

```bash
chmod +x HighFrequencyDNSChecker-linux-amd64
./HighFrequencyDNSChecker-linux-amd64
```

Launch Parameters:
```bash
-config string
      Path to the configuration file (default "config.yaml")
-logMaxAge int
      Maximum log file age (default 10)
-logMaxFiles int
      Maximum number of log files (default 10)
-logMaxSize int
      Max size for log file (Mb) (default 10)
-logPath string
      Path to the log file (default "log")
-logSeverity string
      Min log severity (default "debug")
--help 
      Show help
```

Example of Starting the Utility:
```bash
chmod +x HighFrequencyDNSChecker-linux-amd64
./HighFrequencyDNSChecker-linux-amd64
```

### During operation, the utility creates the following files:
- Log file in JSON format
- Audit file in JSON format
- Logs are compressed into a gz archive upon rotation
 