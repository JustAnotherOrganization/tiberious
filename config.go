package tiberious

type (
	// Config is the Tiberious configuration.
	Config struct {
		// EnableGuests allows guest users to connect on Tiberious when true.
		// Default true.
		EnableGuests bool
	}
)

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		EnableGuests: true,
	}
}
