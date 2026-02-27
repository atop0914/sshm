package models

// Host represents an SSH host entry
type Host struct {
	ID       string   `json:"id" yaml:"id"`
	Name     string   `json:"name" yaml:"name"`
	Host     string   `json:"host" yaml:"host"`
	Port     int      `json:"port" yaml:"port"`
	User     string   `json:"user" yaml:"user"`
	Identity string   `json:"identity,omitempty" yaml:"identity,omitempty"`
	Proxy    string   `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
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
	Hosts  []Host      `json:"hosts" yaml:"hosts"`
	Configs []SSHConfig `json:"configs" yaml:"configs"`
}
