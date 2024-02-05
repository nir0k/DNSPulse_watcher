# High-Frequency DNS Server Polling Utility

## Description:
This utility is designed for high-frequency verification of DNS server name resolutions from a pre-determined list. The results of these checks are saved to a Prometheus database using the remote write feature, ensuring access to data in real-time. This provides users with comprehensive tools for analyzing the efficiency and availability of server infrastructure, significantly enhancing their ability to effectively monitor and maintain system integrity.

## Key features of the utility:
- High-frequency polling (by default, every 150 ms)
- Dynamic formation of a name for DNS servers resolution. (A prefix in the form of unixtime with milliseconds is added)
- Verification of the resolved address (it is verified that the address 1.1.1.1 was resolved)
- Continuous monitoring. (implies 24/7 operation)
- Ability to choose the labels sent in the configuration
- Configuration synchronization with the central server
- Logging of the utility's operation with a choice of logging level
- Log rotation according to specified rules

### Requirements for running the utility:
**Attention, the utility's operation was only tested on MacOS and Linux operating systems**

- Configuration file in yaml format
- File with the list of servers being polled

### Example configuration file:
```yaml
General:
  location: K2          # Location of the server with the utility
  securityZone: PROD    # Security zone of the server with the utility

ConfigHub:
    host: localhost           # IP address or hostname of the central server
    port: 50051               # Port for connecting to the central server
    segment: <SegmentName>    # Name of the segment in which the Watcher is located
    token: "<Token>"          # Authorization token
    encryptionKey: "<Token>"  # Encryption key
    path: "data"              # Path to the directory for storing configuration data and the list of servers polled from the central server
```


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
 