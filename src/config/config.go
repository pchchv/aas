package config

type ServerConfig struct {
	BaseURL            string
	LogSQL             bool
	KeyFile            string
	CertFile           string
	InternalBaseURL    string
	ListenHostHttps    string
	ListenPortHttps    int
	ListenHostHttp     string
	ListenPortHttp     int
	TrustProxyHeaders  bool
	SetCookieSecure    bool
	LogHttpRequests    bool
	StaticDir          string
	TemplateDir        string
	AuditLogsInConsole bool
}

type DatabaseConfig struct {
	DSN      string
	Type     string
	Name     string
	Host     string
	Port     int
	Password string
	Username string
}
