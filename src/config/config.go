package config

import (
	"flag"
	"sync"
)

var (
	cfg          Config
	once         sync.Once
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

// Init initializes the configuration and sets the active server.
func Init(server string) {
	once.Do(load)
	setActiveServer(server)
}

func Get() *ServerConfig {
	return activeConfig
}

func GetAdminEmail() string {
	return cfg.AdminEmail
}

func GetAdminPassword() string {
	return cfg.AdminPassword
}

func GetAdminConsole() *ServerConfig {
	return &cfg.AdminConsole
}

func GetAppName() string {
	return cfg.AppName
}

func GetAuthServer() *ServerConfig {
	return &cfg.AuthServer
}

func GetDatabase() *DatabaseConfig {
	return &cfg.Database
}

// setActiveServer sets the active server configuration.
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

func load() {
	// Auth server
	flag.StringVar(&cfg.AuthServer.BaseURL, "authserver-baseurl", cfg.AuthServer.BaseURL, "Goiabada auth server base URL")
	flag.StringVar(&cfg.AuthServer.InternalBaseURL, "authserver-internalbaseurl", cfg.AuthServer.InternalBaseURL, "Goiabada auth server internal base URL")
	flag.StringVar(&cfg.AuthServer.ListenHostHttps, "authserver-listen-host-https", cfg.AuthServer.ListenHostHttps, "Auth server https host")
	flag.IntVar(&cfg.AuthServer.ListenPortHttps, "authserver-listen-port-https", cfg.AuthServer.ListenPortHttps, "Auth server https port")
	flag.StringVar(&cfg.AuthServer.ListenHostHttp, "authserver-listen-host-http", cfg.AuthServer.ListenHostHttp, "Auth server http host")
	flag.IntVar(&cfg.AuthServer.ListenPortHttp, "authserver-listen-port-http", cfg.AuthServer.ListenPortHttp, "Auth server http port")
	flag.BoolVar(&cfg.AuthServer.TrustProxyHeaders, "authserver-trust-proxy-headers", cfg.AuthServer.TrustProxyHeaders, "Trust HTTP headers from reverse proxy in Auth server? (True-Client-IP, X-Real-IP or the X-Forwarded-For headers)")
	flag.BoolVar(&cfg.AuthServer.SetCookieSecure, "authserver-set-cookie-secure", cfg.AuthServer.SetCookieSecure, "Set secure flag on cookies for auth server")
	flag.BoolVar(&cfg.AuthServer.LogHttpRequests, "authserver-log-http-requests", cfg.AuthServer.LogHttpRequests, "Log HTTP requests for auth server")
	flag.StringVar(&cfg.AuthServer.CertFile, "authserver-certfile", cfg.AuthServer.CertFile, "Certificate file for HTTPS (auth server)")
	flag.StringVar(&cfg.AuthServer.KeyFile, "authserver-keyfile", cfg.AuthServer.KeyFile, "Key file for HTTPS (auth server)")
	flag.BoolVar(&cfg.AuthServer.LogSQL, "authserver-log-sql", cfg.AuthServer.LogSQL, "Log SQL queries for auth server")
	flag.BoolVar(&cfg.AuthServer.AuditLogsInConsole, "authserver-audit-logs-in-console", cfg.AuthServer.AuditLogsInConsole, "Enable audit logs in console output for auth server")
	flag.StringVar(&cfg.AuthServer.StaticDir, "authserver-staticdir", cfg.AuthServer.StaticDir, "Static files directory for auth server")
	flag.StringVar(&cfg.AuthServer.TemplateDir, "authserver-templatedir", cfg.AuthServer.TemplateDir, "Template files directory for auth server")

	// Admin console
	flag.StringVar(&cfg.AdminConsole.BaseURL, "adminconsole-baseurl", cfg.AdminConsole.BaseURL, "Goiabada admin console base URL")
	flag.StringVar(&cfg.AdminConsole.InternalBaseURL, "adminconsole-internalbaseurl", cfg.AdminConsole.InternalBaseURL, "Goiabada admin console internal base URL")
	flag.StringVar(&cfg.AdminConsole.ListenHostHttps, "adminconsole-listen-host-https", cfg.AdminConsole.ListenHostHttps, "Admin console https host")
	flag.IntVar(&cfg.AdminConsole.ListenPortHttps, "adminconsole-listen-port-https", cfg.AdminConsole.ListenPortHttps, "Admin console https port")
	flag.StringVar(&cfg.AdminConsole.ListenHostHttp, "adminconsole-listen-host-http", cfg.AdminConsole.ListenHostHttp, "Admin console http host")
	flag.IntVar(&cfg.AdminConsole.ListenPortHttp, "adminconsole-listen-port-http", cfg.AdminConsole.ListenPortHttp, "Admin console http port")
	flag.BoolVar(&cfg.AdminConsole.TrustProxyHeaders, "adminconsole-trust-proxy-headers", cfg.AdminConsole.TrustProxyHeaders, "Trust HTTP headers from reverse proxy in Admin console? (True-Client-IP, X-Real-IP or the X-Forwarded-For headers)")
	flag.BoolVar(&cfg.AdminConsole.SetCookieSecure, "adminconsole-set-cookie-secure", cfg.AdminConsole.SetCookieSecure, "Set secure flag on cookies for admin console")
	flag.BoolVar(&cfg.AdminConsole.LogHttpRequests, "adminconsole-log-http-requests", cfg.AdminConsole.LogHttpRequests, "Log HTTP requests for admin console")
	flag.StringVar(&cfg.AdminConsole.CertFile, "adminconsole-certfile", cfg.AdminConsole.CertFile, "Certificate file for HTTPS (admin console)")
	flag.StringVar(&cfg.AdminConsole.KeyFile, "adminconsole-keyfile", cfg.AdminConsole.KeyFile, "Key file for HTTPS (admin console)")
	flag.BoolVar(&cfg.AdminConsole.LogSQL, "adminconsole-log-sql", cfg.AdminConsole.LogSQL, "Log SQL queries for admin console")
	flag.BoolVar(&cfg.AdminConsole.AuditLogsInConsole, "adminconsole-audit-logs-in-console", cfg.AdminConsole.AuditLogsInConsole, "Enable audit logs in console output for admin console")
	flag.StringVar(&cfg.AdminConsole.StaticDir, "adminconsole-staticdir", cfg.AdminConsole.StaticDir, "Static files directory for admin console")
	flag.StringVar(&cfg.AdminConsole.TemplateDir, "adminconsole-templatedir", cfg.AdminConsole.TemplateDir, "Template files directory for admin console")

	// Database
	flag.StringVar(&cfg.Database.Type, "db-type", cfg.Database.Type, "Database type. Options: mysql, sqlite")
	flag.StringVar(&cfg.Database.Username, "db-username", cfg.Database.Username, "Database username")
	flag.StringVar(&cfg.Database.Password, "db-password", cfg.Database.Password, "Database password")
	flag.StringVar(&cfg.Database.Host, "db-host", cfg.Database.Host, "Database host")
	flag.IntVar(&cfg.Database.Port, "db-port", cfg.Database.Port, "Database port")
	flag.StringVar(&cfg.Database.Name, "db-name", cfg.Database.Name, "Database name")
	flag.StringVar(&cfg.Database.DSN, "db-dsn", cfg.Database.DSN, "Database DSN (only for sqlite)")

	// Initial setup
	flag.StringVar(&cfg.AdminEmail, "admin-email", cfg.AdminEmail, "Default admin email")
	flag.StringVar(&cfg.AdminPassword, "admin-password", cfg.AdminPassword, "Default admin password")
	flag.StringVar(&cfg.AppName, "appname", cfg.AppName, "Default app name")

	flag.Parse()
}
