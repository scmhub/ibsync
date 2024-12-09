package ibsync

import (
	"math/rand"
	"time"
)

var log = Logger()

const (
	// Default values for the connection parameters.
	TIMEOUT = 30 * time.Second // Default timeout duration
	HOST    = "127.0.0.1"      // Default host
	PORT    = 7497             // Default port
)

// Config holds the connection parameters for the client.
// This struct centralizes all the configurable options for creating a connection.
type Config struct {
	Host     string        // Host address for the connection
	Port     int           // Port number for the connection
	ClientID int64         // Client ID, default is randomized for uniqueness
	InSync   bool          // Stay in sync with server
	Timeout  time.Duration // Timeout for the connection
	ReadOnly bool          // Indicates if the client should be in read-only mode
	Account  string        // Optional account identifier
}

// NewConfig creates a new Config with default values, and applies any functional options.
// The functional options allow customization of the config without directly modifying fields.
func NewConfig(options ...func(*Config)) *Config {
	config := &Config{
		Host:     HOST,                    // Default host
		Port:     PORT,                    // Default port
		ClientID: rand.Int63n(999999) + 1, // Random default client ID to avoid collisions. +1 for non 0 id.
		InSync:   true,                    // Default true. Client is kept in sync with the TWS/IBG application
		Timeout:  TIMEOUT,                 // Default timeout
	}

	// Apply any functional options passed to the NewConfig function
	for _, option := range options {
		option(config)
	}

	return config
}

// WithHost is a functional option to set a custom host for the Config.
func WithHost(host string) func(*Config) {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort is a functional option to set a custom port for the Config.
func WithPort(port int) func(*Config) {
	return func(c *Config) {
		c.Port = port
	}
}

// WithClientID is a functional option to set a specific ClientID.
// Useful if you want to set a fixed client ID instead of a random one.
func WithClientID(id int64) func(*Config) {
	return func(c *Config) {
		c.ClientID = id
	}
}

// WithClientZero is a shortcut functional option that sets the ClientID to 0.
// ClientID = 0 is a privileged ClientID that gives your API session access to all order updates,
// including those entered manually in the TWS or by other API clients.
func WithClientZero() func(*Config) {
	return func(c *Config) {
		c.ClientID = 0
	}

}

// WithoutSync is a functional option that sets InSybc
func WithoutSync() func(*Config) {
	return func(c *Config) {
		c.InSync = false
	}
}

// WithTimeout is a functional option to customize the timeout for the connection.
func WithTimeout(timeout time.Duration) func(*Config) {
	return func(c *Config) {
		c.Timeout = timeout
	}
}
