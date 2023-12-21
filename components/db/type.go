package sqldb


type Config struct {
    General            MainConfiguration           `json:"general" yaml:"General"`
    Log                LogConfiguration            `json:"log" yaml:"Log"`
    Audit              LogConfiguration            `json:"audit" yaml:"Audit"`
    WebServer          WebServerConfiguration      `json:"webServer" yaml:"WebServer"`
    Sync               SyncConfiguration           `json:"sync" yaml:"Sync"`
    Prometheus         PrometheusConfiguration     `json:"prometheus" yaml:"Prometheus"`
    PrometheusLabels   PrometheusLabelConfiguration `json:"prometheusLabels" yaml:"PrometheusLabels"`
    Resolvers          ResolversConfiguration      `json:"resolvers" yaml:"Resolvers"`
    Watcher            WatcherConfiguration        `json:"watcher" yaml:"Watcher"`
}

type MainConfiguration struct {
    DBname         string `json:"dbName" yaml:"db_name"`
    Hostname       string `json:"hostname"`
    IPAddress      string `json:"ipAddress"`
    ConfPath       string `json:"confPath" yaml:"confPath"`
    Sync           bool   `json:"sync" yaml:"sync"`
    UpdateInterval int    `json:"updateInterval" yaml:"confCheckInterval"`
    Hash           string `json:"Hash"`
    LastCheck      int64  `json:"lastCheck"`
    LastUpdate     int64  `json:"lastUpdate"`
}

type LogConfiguration struct {
    Path        string `json:"path" yaml:"path"`
    MinSeverity string `json:"minSeverity" yaml:"minSeverity"`
    MaxAge      int    `json:"maxAge" yaml:"maxAge"`
    MaxSize     int    `json:"maxSize" yaml:"maxSize"`
    MaxFiles    int    `json:"maxFiles" yaml:"maxFiles"`
}

type WebServerConfiguration struct {
    Port          string `json:"port" yaml:"port"`
    SslIsEnable   bool   `json:"sslIsEnable" yaml:"sslIsEnable"`
    SslCertPath   string `json:"sslCertPath" yaml:"sslCertPath"`
    SslKeyPath    string `json:"sslKeyPath" yaml:"sslKeyPath"`
    SesionTimeout int    `json:"sessionTimeout" yaml:"sesionTimeout"`
    InitUsername  string `json:"initUsername" yaml:"initUsername"`
    InitPassword  string `json:"initPassword" yaml:"initPassword"`
}

type SyncConfiguration struct {
    IsEnable bool                 `json:"-" yaml:"isEnable"`
    Token string                   `json:"-" yaml:"Token"`
    Members  []MemberConfiguration `json:"members" yaml:"members"`
}

type PrometheusConfiguration struct {
    Url           string `json:"url" yaml:"url"`
    MetricName    string `json:"metricName" yaml:"metricName"`
    Auth          bool   `json:"auth" yaml:"auth"`
    Username      string `json:"username" yaml:"username"`
    Password      string `json:"password" yaml:"password"`
    RetriesCount  int    `json:"retriesCount" yaml:"retriesCount"`
    BuferSize    int    `json:"buferSize" yaml:"buferSize"`
}

type PrometheusLabelConfiguration struct {
    Opcode             bool `json:"opcode" yaml:"opcode"`
    Authoritative      bool `json:"authoritative" yaml:"authoritative"`
    Truncated          bool `json:"truncated" yaml:"truncated"`
    Rcode              bool `json:"rcode" yaml:"rcode"`
    RecursionDesired   bool `json:"recursionDesired" yaml:"recursionDesired"`
    RecursionAvailable bool `json:"recursionAvailable" yaml:"recursionAvailable"`
    AuthenticatedData  bool `json:"authenticatedData" yaml:"authenticatedData"`
    CheckingDisabled   bool `json:"checkingDisabled" yaml:"checkingDisabled"`
    PollingRate        bool `json:"pollingRate" yaml:"pollingRate"`
    Recursion          bool `json:"recursion" yaml:"recursion"`
}


type ResolversConfiguration struct {
    Path           string `json:"path" yaml:"path"`
    PullTimeout    int    `json:"pullTimeout" yaml:"pullTimeout"`
    Delimeter      string `json:"delimiter" yaml:"delimeter"`
    ExtraDelimeter string `json:"extraDelimiter" yaml:"extraDelimeter"`
    Hash           string `json:"hash"`
    LastCheck      int64  `json:"lastCheck"`
    LastUpdate     int64  `json:"lastUpdate"`
}

type MemberConfiguration struct {
    SyncID          int `json:"-"`
    Hostname        string `json:"hostname" yaml:"hostname"`
    Port            string `json:"port" yaml:"port"`
    IPAddress       string `json:"-"`
    Location        string `json:"-"`
    SecurityZone    string `json:"-"`
    SeverLastCheck  int64 `json:"-"`
    ConfigHash      string `json:"-"`
    ConfigLastCheck int64 `json:"-"`
    ConfigLastUpdate int64 `json:"-"`
    ResolvHash      string `json:"-"`
    ResolvLastCheck int64 `json:"-"`
    ResolvLastUpdate int64 `json:"-"`
    IsLocal         bool `json:"-"`
    SyncEnable      bool `json:"-"`
}

type WatcherConfiguration struct {
    Location     string `json:"location" yaml:"location"`
    SecurityZone string `json:"securityZone" yaml:"securityZone"`
}

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

type Resolver struct {
    Server 				string
	IPAddress 			string
	Domain 				string
    Location 			string
    Site 				string
    ServerSecurityZone	string
    Prefix 				string
    Protocol 			string
    Zonename 			string
    Recursion 			bool
    QueryCount 			int
    ServiceMode 		bool
}

type MemberForSync struct {
    Hostname    string
    Port        string
}

type LogsPaths struct {
    Log string
    Audit string
}