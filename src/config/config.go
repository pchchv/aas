package config

var (
	cfg          Config
	activeConfig *ServerConfig
)

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

type Config struct {
	AppName       string
	Database      DatabaseConfig
	AdminEmail    string
	AuthServer    ServerConfig
	AdminConsole  ServerConfig
	AdminPassword string
}

func Get() *ServerConfig {
	return activeConfig
}

// setActiveServer sets the active server configuration
func setActiveServer(server string) {
	switch server {
	case "AuthServer":
		activeConfig = &cfg.AuthServer
	case "AdminConsole":
		activeConfig = &cfg.AdminConsole
	default:
		panic("Invalid active server configuration specified")
	}
}
