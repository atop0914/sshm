package models

// Profile represents SSH connection profile settings
type Profile struct {
	Name                string `json:"name" yaml:"name"`
	Timeout            int    `json:"timeout" yaml:"timeout"`                     // Connection timeout in seconds
	KeepAliveInterval  int    `json:"keepalive_interval" yaml:"keepalive_interval"` // Keep-alive interval in seconds
	KeepAliveCountMax  int    `json:"keepalive_count_max" yaml:"keepalive_count_max"` // Max keep-alive count before disconnect
	ServerAliveEnabled bool   `json:"server_alive_enabled" yaml:"server_alive_enabled"` // Enable server alive messages
}

// DefaultProfile returns the default profile settings
func DefaultProfile() Profile {
	return Profile{
		Name:                "default",
		Timeout:            30,
		KeepAliveInterval:  15,
		KeepAliveCountMax:  3,
		ServerAliveEnabled: true,
	}
}
