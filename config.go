package tiberious

type (
	// Config is the Tiberious configuration.
	Config struct {
		// EnableGuests allows guest users to connect on Tiberious when true.
		// Default true.
		EnableGuests bool
		// EnableArchive controls whether Tiberious will archive messages for
		// offline users to retrieve later.
		// Default true.
		EnableArchive bool
	}
)

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		EnableGuests:  true,
		EnableArchive: true,
	}
}
