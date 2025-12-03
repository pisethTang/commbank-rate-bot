package config

import "time"

// Config holds all application configuration
type Config struct {
	TargetRate    float64
	CheckInterval time.Duration
}

// Initialize returns the configuration
func Initialize() Config {
	return Config{
		TargetRate:    1.20,             // Target USD to AUD rate to trigger notification
		CheckInterval: 60 * time.Minute, // Interval between checks
	}
}
