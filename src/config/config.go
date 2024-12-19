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
