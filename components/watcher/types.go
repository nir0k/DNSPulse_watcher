package watcher

type Csv struct {
	Server			            string `csv:"server"`
	Server_ip		            string `csv:"server_ip"`
	Domain			            string `csv:"domain"`
    Location                    string `csv:"location"`
    Site                        string `csv:"site"`
    Server_security_zone        string `csv:"server_security_zone"`
    Prefix                      string `csv:"prefix"`
    Protocol                    string `csv:"protocol"`
    Zonename		            string `csv:"zonename"`
    Query_count_rps             string `csv:"query_count_rps"`
    Zonename_with_recursion     string `csv:"zonename_with_recursion"`
    Query_count_with_recursion  string `csv:"query_count_with_recursion_rps"`
    Service_mode            string `csv:"service_mode"`
}

type Resolver struct {
    Server string
	Server_ip string
	Domain string
    Location string
    Site string
    Server_security_zone string
    Prefix string
    Protocol string
    Zonename string
    Recursion bool
    Query_count_rps int
    Service_mode bool
}


type prometheus struct {
    Url string
    Auth  bool
    Username string
    Password string
    Metric string
    Retries int
    Metrics metrics
}


type metrics struct {
    Rcode bool
    Opscodes bool
    Authoritative bool
    Truncated bool
    RecursionDesired bool
    RecursionAvailable bool
    AuthenticatedData bool
    CheckingDisabled bool
    Polling_rate bool
    Recusrion bool
}


type dns_param struct {
    Timeout  int
    Dns_servers_path string
    Dns_servers_file_md5hash string
    Delimeter rune
    Delimeter_for_additional_field string
}


type config struct {
    Conf_path string
    Conf_md5hash string
    Check_interval int
    Buffer_size int
    Ip string
    Hostname string
    Location string
    SecurityZone string
}
