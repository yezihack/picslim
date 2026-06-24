package config

type Config struct {
	EnableLocalHTTP bool
	HTTPPort        int
	LogLevel        string
}

func LoadDefault() Config {
	return Config{
		EnableLocalHTTP: false,
		HTTPPort:        18080,
		LogLevel:        "info",
	}
}
