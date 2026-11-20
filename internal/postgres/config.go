package postgres

type ConnectionConfig struct {
	Host          string
	Port          int32
	SSLMode       string
	AdminUsername string
	AdminPassword string
}

type DatabaseConfig struct {
	Name     string
	Username string
	Password string
}
