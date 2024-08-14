package redis

// Config represents the configuration settings for connecting to a Redis database.
type Config struct {
	Host        string
	Port        int
	ConnTimeout int
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
	Address     string
}
