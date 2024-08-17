package config

type Log struct {
	// "debug" | "info" | "warn" | "error"
	Level string
	// "json" or "text"
	Format string
}
