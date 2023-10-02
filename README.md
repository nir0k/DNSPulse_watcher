# HighFrequencyDNSChecker

Program for high-frequency resolution of server names on DNS servers from a list and recording the result in Prometheus.
Program Features:
- High-frequency polling (default every 150ms).
- A random server name is queried with a fixed prefix and it is checked that the IP address 1.1.1.1 is resolved. Server name create on this rule: <timestamp with miliseconds>.<hostname prefix>
- Work in an infinite loop.

### Install
**requred go 1.21.1**
```bash
git clone https://github.com/nir0k/HighFrequencyDNSChecker.git
cd frequency-dns-checker/main
go build .
```

### Prepare
Create and fill out the file .env in the same folder with the program. 
Complited example we found in the file .env-example in the project

### Use
```bash
./frequency-dns-checker
```


### Custom Rcode:
- 3841 - Resolved IP-address not equals 1.1.1.1
- 3842 - DNS Server not answer
