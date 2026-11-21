package postgres

type ConnectionConfig struct {
	Host     string
	Port     int32
	SSLMode  string
	Username string
	Password string
}

type DatabaseConfig struct {
	Name     string
	Username string
	Password string
}
