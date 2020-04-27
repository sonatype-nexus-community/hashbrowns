package types

// Config is basic config for hashbrowns
type Config struct {
	LogLevel    int
	Path        string
	User        string
	Token       string
	Server      string
	Application string
	Stage       string
	MaxRetries  int
}
