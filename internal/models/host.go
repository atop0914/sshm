package models

// AuthType represents the authentication method
type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeKey      AuthType = "key"
	AuthTypeAgent    AuthType = "agent"
)

// Host represents an SSH host entry
type Host struct {
	ID              string    `json:"id" yaml:"id"`
	Name            string    `json:"name" yaml:"name"`
	Host            string    `json:"host" yaml:"host"`
	Port            int       `json:"port" yaml:"port"`
	User            string    `json:"user" yaml:"user"`
	Password        string    `json:"password,omitempty" yaml:"password,omitempty"`
	Identity        string    `json:"identity,omitempty" yaml:"identity,omitempty"`
	AuthType        AuthType  `json:"auth_type,omitempty" yaml:"auth_type,omitempty"`
	Proxy           string    `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	Group           string    `json:"group,omitempty" yaml:"group,omitempty"`
	Tags            []string  `json:"tags,omitempty" yaml:"tags,omitempty"`
	ConnectionCount int       `json:"connection_count,omitempty" yaml:"connection_count,omitempty"`
	Profile         string    `json:"profile,omitempty" yaml:"profile,omitempty"` // Profile name to use for this host
}

// SSHConfig represents SSH configuration settings
type SSHConfig struct {
	Name               string `json:"name" yaml:"name"`
	IdentityFile       string `json:"identity_file,omitempty" yaml:"identity_file,omitempty"`
	ProxyCommand       string `json:"proxy_command,omitempty" yaml:"proxy_command,omitempty"`
	ForwardAgent       bool   `json:"forward_agent,omitempty" yaml:"forward_agent,omitempty"`
	ServerAliveInterval int   `json:"server_alive_interval,omitempty" yaml:"server_alive_interval,omitempty"`
}

// Config holds the entire application configuration
type Config struct {
	Hosts     []Host     `json:"hosts" yaml:"hosts"`
	Configs   []SSHConfig `json:"configs" yaml:"configs"`
	Profiles  []Profile  `json:"profiles" yaml:"profiles"`
}
